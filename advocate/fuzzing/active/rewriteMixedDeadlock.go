// Copyright (c) 2024 Erik Kassubek
//
// File: advocate/analysis/rewriter/mixedDeadlock.go
// Brief: Rewrite trace for mixed deadlocks
//
// Author: Erik Kassubek, Ilian Kohl
// Created: 2025-01-01
//
// License: BSD-3-Clause

package active

import (
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/bugs"
	"advocate/trace"
	"errors"
	"fmt"
)

// -----------------------------------------------------------------------
//
// rewriteMixedDeadlock rewrites trace to provoke mixed deadlock
// (Bug encoding TraceElement2, as produced by reportMixedDeadlockCycle)
//
// -----------------------------------------------------------------------
// Algorithm (mirrors rewriteCyclicDeadlock):
//
//  1. lastTime = max tPre across all cycle elements

//  2. ShortenTrace(lastTime, true) - drop everything after lastTime

//  3. For each element in the cycle:
//       CB (ElementChannel)-> ShortenRoutine(r, CB.tPre)
//                             (re-executes channel op, blocks while holding lock)

//       MB (ElementMutex)  -> ShortenRoutine(r, MB.tPre)
// 							   (first of a pair where next is channel, same routine)
//							   (re-executes lock acquire, blocks because InCS side holds it)

//       DC (ElementMutex)  -> ShortenRoutine(r, DC.tPre)
// 							   (lone DC with no adjacent channel of same routine)
// 							   (RD node: re-executes lock request, blocks)

//       DC following a CB (InCS acquire) -> no shortening; used for ordering only
//       informational channel (after MB) -> no shortening

//  4. Enforce acquire ordering for InCS pairs on same lock:
//  	- In  original trace one InCS routine acquired before the other
//        To provoke deadlock, routine that acquired LATER must now acquire FIRST
//      - Use ShiftRoutine to push originally-earlier routine's acquire to after
// 		  the originally-later one (like rewriteCyclicDeadlock)

//  5. Shift any unlock of a cycle lock appearing after the blocking point to before it

//  6. AddTraceElementReplay(lastTime+1, ExitCodeMixedDeadlock)
//
// -----------------------------------------------------------------------
// Example: MD2-1B (both InCS, buffered channel):
//
//	Original trace:
//	  Routine 2 (sender):   m.Lock(tPre=14)  c<-1(tPre=20)  m.Unlock(tPre=24)
//	  Routine 3 (receiver): m.Lock(tPre=36)  <-c(tPre=42)   m.Unlock(tPre=46)
//	  Sender acquired the lock first (14 < 36); send succeeded (buffered)
//	  Receiver acquired after, received from buffer; no deadlock
//
//	Cycle elements from reportMixedDeadlockCycle:
//	  CB: ElementChannel routine=2 tPre=20 (send — InCS blocking point)
//	  DC: ElementMutex   routine=2 tPre=14 (InCS acquire, lock ID + ordering)
//	  CB: ElementChannel routine=3 tPre=42 (recv — InCS blocking point)
//	  DC: ElementMutex   routine=3 tPre=36 (InCS acquire, lock ID + ordering)
//
//	Rewritten trace:
//	  Routine 3 acquires first -> m.Lock(36) -> <-c on empty buffer -> blocks
//	  Routine 2 tries m.Lock(14) -> blocks (routine 3 holds it)
//	  Deadlock: routine 3 blocks on empty channel; routine 2 blocks on mutex.
//
//	Shortening:
//	  ShortenRoutine(2, 20) -> routine 2 retains only: m.Lock(14)
//	  ShortenRoutine(3, 42) -> routine 3 retains only: m.Lock(36)
//
//	Ordering:
//	  DC for routine 2 has tPre=14, DC for routine 3 has tPre=36.
//	  Routine 2 acquired first (14 < 36) -> routine 3 must go first.
//	  ShiftRoutine(routine=2, startTPre=14, shift) so that routine 2's lock
//	  acquire is pushed after routine 3's, by finding the first element in
//	  routine 2's remaining trace that is concurrent with routine 3's acquire
//	  and shifting from there.
// -----------------------------------------------------------------------

// rewriteMixedDeadlock rewrites trace to provoke the mixed deadlock encoded in bug.TraceElement2 (cycle elements)
func rewriteMixedDeadlock(tr *trace.Trace, bug bugs.Bug, exitCode int) error {
	cycle := bug.TraceElement2
	if len(cycle) < 2 {
		return errors.New("mixed deadlock rewrite: need at least 2 cycle elements")
	}

	// Step 1: lastTime = max tPre across all cycle elements
	lastTime := findLastTime(cycle)
	if lastTime <= 0 {
		return errors.New("mixed deadlock rewrite: invalid lastTime")
	}

	// Step 2: drop everything after lastTime
	tr.ShortenTrace(lastTime, true)

	// Step 3: determine the blocking point for each routine and shorten & walk cycle list
	//
	// Encoding from reportMixedDeadlockCycle produces alternating pairs (channel, mutex)
	// or (mutex, channel) for CD nodes, and lone mutex elements for RD nodes
	//
	// Classification rules based on Go type of adjacent elements
	//   [ElementChannel at i, ElementMutex at i+1, same routine]
	//     -> CB + DC  (InCS): shorten to channel element's tPre
	//
	//   [ElementMutex at i, ElementChannel at i+1, same routine]
	//     -> MB + informational channel  (PCS): shorten to mutex element's tPre
	//
	//   [ElementMutex at i, no adjacent channel of same routine]
	//     -> lone DC  (RD node): shorten to mutex element's tPre
	n := len(cycle)

	// shortenTo: routine -> tPre to shorten to (the blocking element's tPre)
	shortenTo := make(map[int]int)

	// dcAcqTPre: lockID -> list of (routine, tPre-of-acquire) for InCS DC elements
	// To later enforce ordering between InCS pairs
	type dcInfo struct {
		routine int
		tPre    int
	}
	dcAcqByLock := make(map[int][]dcInfo) // lockID -> []dcInfo

	claimed := make([]bool, n)

	// First pass: identify adjacent same-routine pairs.
	for i := 0; i < n; i++ {
		if claimed[i] {
			continue
		}
		next := (i + 1) % n
		if claimed[next] {
			continue
		}
		ei := cycle[i]
		ej := cycle[next]
		if ei == nil || ej == nil || ei.GetRoutine() != ej.GetRoutine() {
			continue
		}

		chanI, isChannelI := ei.(*trace.ElementChannel)
		mutI, isMutexI := ei.(*trace.ElementMutex)
		chanJ, isChannelJ := ej.(*trace.ElementChannel)
		mutJ, isMutexJ := ej.(*trace.ElementMutex)

		switch {
		case isChannelI && isMutexJ && mutJ.IsLock():
			// CB + DC  (InCS)
			claimed[i] = true
			claimed[next] = true
			// Shorten to channel blocking point
			r := chanI.GetRoutine()
			t := chanI.GetTPre()
			if cur, set := shortenTo[r]; !set || t < cur {
				shortenTo[r] = t
			}
			// Record DC acquire info for ordering
			dcAcqByLock[mutJ.GetObjId()] = append(dcAcqByLock[mutJ.GetObjId()], dcInfo{
				routine: mutJ.GetRoutine(),
				tPre:    mutJ.GetTPre(),
			})

		case isMutexI && mutI.IsLock() && isChannelJ:
			// MB + informational channel  (PCS)
			claimed[i] = true
			claimed[next] = true
			// Shorten to mutex blocking point.
			r := mutI.GetRoutine()
			t := mutI.GetTPre()
			if cur, set := shortenTo[r]; !set || t < cur {
				shortenTo[r] = t
			}
			// MB lock ID is collected for unlock-shift but NOT for ordering
			// (only one side holds the lock in PCS patterns)
			_ = chanJ
		}
	}

	// Second pass: handle unclaimed elements (lone DC / RD nodes)
	for i := 0; i < n; i++ {
		if claimed[i] || cycle[i] == nil {
			continue
		}
		mut, isMutex := cycle[i].(*trace.ElementMutex)
		if !isMutex || !mut.IsLock() {
			continue
		}
		// Lone DC: RD node - shorten to lock-request tPre
		r := mut.GetRoutine()
		t := mut.GetTPre()
		if cur, set := shortenTo[r]; !set || t < cur {
			shortenTo[r] = t
		}
	}

	// Apply shortening
	for r, t := range shortenTo {
		tr.ShortenRoutine(r, t)
		fmt.Printf("[rewriteMixedDeadlock] ShortenRoutine(%d, %d)\n", r, t)
	}

	// Step 4: enforce acquire ordering for InCS pairs on the same lock
	//
	// For each lock that has exactly two InCS DC entries, we know which
	// routine acquired first in the original trace (smaller tPre)
	// To provoke deadlock we must make the later-acquiring routine go first
	for lockID, infos := range dcAcqByLock {
		if len(infos) != 2 {
			// Not a two-InCS pair on this lock - skip
			continue
		}
		_ = lockID

		// Identify which routine acquired first
		first, second := infos[0], infos[1]
		if second.tPre < first.tPre {
			first, second = second, first
		}
		// first acquired earlier; second acquired later
		// We want second to go before first -> shift first's remaining trace

		// Find the acquire element of 'second' in the (already shortened) trace
		var secondAcqElem trace.Element
		for _, e := range tr.GetRoutineTrace(second.routine) {
			mut, ok := e.(*trace.ElementMutex)
			if ok && mut.IsLock() && mut.GetTPre() == second.tPre {
				secondAcqElem = e
				break
			}
		}
		if secondAcqElem == nil {
			fmt.Printf("[rewriteMixedDeadlock] ordering: could not find acquire elem"+
				" for routine %d lock tPre=%d\n", second.routine, second.tPre)
			continue
		}

		// Find the first element in 'first' routine's remaining trace that is
		// concurrent with second's acquire (same as rewriteCyclicDeadlock)
		var concurrentStart trace.Element
		for _, e := range tr.GetRoutineTrace(first.routine) {
			if clock.GetHappensBefore(e.GetWVC(), secondAcqElem.GetWVC()) == hb.Concurrent {
				concurrentStart = e
				break
			}
		}
		if concurrentStart == nil {
			fmt.Printf("[rewriteMixedDeadlock] ordering: no concurrent start"+
				" for routine %d vs routine %d\n", first.routine, second.routine)
			continue
		}

		// Compute shift so that first's concurrent section starts after second's acquire
		shift := (secondAcqElem.GetTSort() - concurrentStart.GetTSort()) + 1

		rt := tr.GetRoutineTrace(first.routine)
		routineEnd := rt[len(rt)-1]
		tr.ShiftRoutine(first.routine, concurrentStart.GetTPre(), shift)
		if routineEnd.GetTPost() > lastTime {
			lastTime = routineEnd.GetTPost()
		}
		tr.ShiftConcurrentOrAfterToAfter(secondAcqElem)

		fmt.Printf("[rewriteMixedDeadlock] ordering: shifted routine %d by %d"+
			" to execute after routine %d\n", first.routine, shift, second.routine)
	}

	// Step 5: shift unlocks of cycle locks appearing after the blocking point to before it
	// Collect all lock IDs involved in the cycle for the unlock scan
	cycleLockIDs := make(map[int]struct{})
	for _, e := range cycle {
		if e == nil {
			continue
		}
		if mut, ok := e.(*trace.ElementMutex); ok && mut.IsLock() {
			cycleLockIDs[mut.GetObjId()] = struct{}{}
		}
	}

	for routine, blockTPre := range shortenTo {
		for _, cand := range tr.GetRoutineTrace(routine) {
			unlock, ok := cand.(*trace.ElementMutex)
			if !ok || unlock.IsLock() {
				continue
			}
			if _, isCycleLock := cycleLockIDs[unlock.GetObjId()]; !isCycleLock {
				continue
			}
			if unlock.GetTSort() < blockTPre {
				// Already before the blocking point - nothing to do
				continue
			}

			var concurrentStart trace.Element
			for _, e := range tr.GetRoutineTrace(routine) {
				if clock.GetHappensBefore(e.GetWVC(), unlock.GetWVC()) == hb.Concurrent {
					concurrentStart = e
					break
				}
			}
			if concurrentStart == nil {
				fmt.Printf("[rewriteMixedDeadlock] unlock shift: no concurrent start"+
					" for routine %d unlock lock %d\n", routine, unlock.GetObjId())
				continue
			}

			rt := tr.GetRoutineTrace(routine)
			routineEnd := rt[len(rt)-1]
			tr.ShiftRoutine(routine, concurrentStart.GetTPre(),
				(unlock.GetTSort()-concurrentStart.GetTSort())+1)
			if routineEnd.GetTPost() > lastTime {
				lastTime = routineEnd.GetTPost()
			}
			tr.ShiftConcurrentOrAfterToAfter(unlock)
		}
	}

	// Step 6: replay end marker
	tr.AddTraceElementReplay(lastTime+1, exitCode)

	fmt.Println("[rewriteMixedDeadlock] cycle elements:")
	for _, e := range cycle {
		if e == nil {
			continue
		}
		fmt.Printf("  routine=%d type=%T tPre=%d id=%d\n",
			e.GetRoutine(), e, e.GetTPre(), e.GetObjId())
	}

	return nil
}
