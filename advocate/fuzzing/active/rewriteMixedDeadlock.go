// advocate/fuzzing/active/rewriteMixedDeadlock.go
//
// Copyright (c) 2024 Erik Kassubek
//
// File: rewriteMixedDeadlock.go
//
// License: BSD-3-Clause

package active

import (
	"advocate/results/bugs"
	"advocate/trace"
	"errors"
	"fmt"
	"math"
)

// rewriteMixedDeadlock rewrites the trace to trigger a two‑cycle mixed deadlock
//
// Handles cases with a main routine (R1) that forks the communicating routines (R2, R3)
// Preserves all main routine operations (forks, wait groups, etc.)
// Only modifies the two communicating routines
func rewriteMixedDeadlock(tr *trace.Trace, bug bugs.Bug, code int) error {
	if len(bug.TraceElement2) != 4 {
		return fmt.Errorf("rewriteMixedDeadlock: expected 4 elements, got %d", len(bug.TraceElement2))
	}

	// Extract elements
	cdHolder, ok1 := bug.TraceElement2[0].(*trace.ElementChannel)
	lockHolder, ok2 := bug.TraceElement2[1].(*trace.ElementMutex)
	_, ok3 := bug.TraceElement2[2].(*trace.ElementChannel)
	lockWaiter, ok4 := bug.TraceElement2[3].(*trace.ElementMutex)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return errors.New("rewriteMixedDeadlock: incorrect element types in cycle")
	}

	holderRout := cdHolder.GetRoutine()
	waiterRout := lockWaiter.GetRoutine()

	// Find the main routine (the one that forks the others)
	// Usually the routine with the lowest ID that's not holder or waiter
	mainRout := 1
	for rid := range tr.GetTraces() {
		if rid != int(holderRout) && rid != int(waiterRout) && rid < mainRout {
			mainRout = rid
		}
	}

	fmt.Printf("rewriteMixedDeadlock: main=R%d, holder=R%d (lock=%d, chan=%d), waiter=R%d (lock=%d)\n",
		mainRout, holderRout, lockHolder.GetTPre(), cdHolder.GetTPre(), waiterRout, lockWaiter.GetTPre())

	// -----------------------------------------------------------------------
	// Step 1: Find the last time we need to keep
	// We need to keep: main routine's last element, holder's lock+chan, waiter's lock
	// -----------------------------------------------------------------------
	lastTime := max(lockHolder.GetTPost(), lockWaiter.GetTPost())

	// Also include main routine's last element if it's after
	if mainTrace := tr.GetRoutineTrace(mainRout); len(mainTrace) > 0 {
		if lastElem := mainTrace[len(mainTrace)-1]; lastElem.GetTPost() > lastTime {
			lastTime = lastElem.GetTPost()
		}
	}
	tr.ShortenTrace(lastTime, true)

	// -----------------------------------------------------------------------
	// Step 2: Shorten each routine
	// Main routine: keep everything (don't shorten)
	// Holder: keep up to its channel op
	// Waiter: keep only its lock acquire
	// -----------------------------------------------------------------------
	// Don't shorten main routine at all - keep all its operations
	// (forks, wait group add/done, etc. must remain)

	tr.ShortenRoutine(holderRout, cdHolder.GetTPost()+1)
	tr.ShortenRoutine(waiterRout, lockWaiter.GetTPost()+1)

	fmt.Printf("rewriteMixedDeadlock: main R%d unchanged, holder R%d kept to t=%d, waiter R%d kept to t=%d\n",
		mainRout, holderRout, cdHolder.GetTPost(), waiterRout, lockWaiter.GetTPost())

	// -----------------------------------------------------------------------
	// Step 3: Reorder so holder's lock happens BEFORE waiter's lock
	// For MD2-1B: waiter (sender) originally acquired first, need to move it after holder
	// -----------------------------------------------------------------------
	if lockWaiter.GetTPre() < lockHolder.GetTPre() {
		// Shift waiter's entire trace forward
		// Target: waiter's lock should happen after holder's lock completes
		targetTPre := lockHolder.GetTPost() + 1
		shift := targetTPre - lockWaiter.GetTPre()

		if shift > 0 {
			// Get the first element in waiter's trace (usually the lock acquire itself)
			waiterTrace := tr.GetRoutineTrace(waiterRout)
			if len(waiterTrace) == 0 {
				return fmt.Errorf("rewriteMixedDeadlock: waiter R%d has no trace", waiterRout)
			}

			// Find earliest element to shift from (could be fork or first operation)
			startElem := waiterTrace[0]
			startTPre := startElem.GetTPre()

			fmt.Printf("rewriteMixedDeadlock: shifting waiter R%d by %d (from t=%d to after holder lock at t=%d)\n",
				waiterRout, shift, lockWaiter.GetTPre(), lockHolder.GetTPost())

			tr.ShiftRoutine(waiterRout, startTPre, shift)

			// Also shift any concurrent elements after the lock
			tr.ShiftConcurrentOrAfterToAfter(lockWaiter)
		}
	} else {
		fmt.Printf("rewriteMixedDeadlock: holder already first (holder lock %d < waiter lock %d)\n",
			lockHolder.GetTPre(), lockWaiter.GetTPre())
	}

	// -----------------------------------------------------------------------
	// Step 4: Ensure holder's channel op happens AFTER its lock acquire
	// -----------------------------------------------------------------------
	if cdHolder.GetTPre() <= lockHolder.GetTPre() {
		newTPre := lockHolder.GetTPost() + 1
		fmt.Printf("rewriteMixedDeadlock: moving holder's channel op from tPre=%d to tPre=%d\n",
			cdHolder.GetTPre(), newTPre)
		cdHolder.SetTPre(newTPre)
		cdHolder.SetTPost(newTPre)
	}

	// -----------------------------------------------------------------------
	// Step 5: Mark blocking elements with tPost=0
	// Holder's channel op blocks, waiter's lock acquire blocks
	// -----------------------------------------------------------------------
	// Set tPost=0 on holder's channel op (receive blocks on empty buffer)
	setTPostZero(cdHolder)
	fmt.Printf("rewriteMixedDeadlock: holder R%d channel op blocked (tPost=0)\n", holderRout)

	// Set tPost=0 on waiter's lock acquire (blocks because holder holds lock)
	setTPostZero(lockWaiter)
	fmt.Printf("rewriteMixedDeadlock: waiter R%d lock acquire blocked (tPost=0)\n", waiterRout)

	// -----------------------------------------------------------------------
	// Step 6: Insert replay marker after all remaining operations
	// -----------------------------------------------------------------------
	newLastTime := 0
	for rid, traceSlice := range tr.GetTraces() {
		for _, elem := range traceSlice {
			t := elem.GetTSort()
			if t > newLastTime && t != math.MaxInt {
				newLastTime = t
			}
		}
		// Debug: print what's left in each routine
		fmt.Printf("rewriteMixedDeadlock: R%d has %d elements, last t=%d\n",
			rid, len(traceSlice), newLastTime)
	}

	tr.AddTraceElementReplay(newLastTime+1, code)
	fmt.Printf("rewriteMixedDeadlock: replay marker at t=%d, code=%d\n", newLastTime+1, code)

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func setTPostZero(elem trace.Element) {
	savedTPre := elem.GetTPre()
	elem.SetTSort(0)        // sets tPre=0, tPost=0
	elem.SetTPre(savedTPre) // restore tPre; tPost stays 0
	fmt.Printf("rewriteMixedDeadlock: setTPostZero on %T: tPre=%d -> tPost=0\n", elem, savedTPre)
}
