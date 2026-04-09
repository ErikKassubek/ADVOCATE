// Copyright (c) 2024 Erik Kassubek
//
// File: rewriteMixedDeadlock.go
// Brief: Rewrite trace for mixed deadlocks (channel + mutex cycles)
//
// Author: Ilian Kohl
//
// License: BSD-3-Clause

package active

import (
	"advocate/results/bugs"
	"advocate/trace"
	"advocate/utils/helper"
	"errors"
	"fmt"
)

// rewriteMixedDeadlock rewrites the trace to trigger the mixed deadlock.
//
// Input:
//
//	bug.TraceElement1 – *trace.ElementMutex, one per RDNode in cycle order
//	bug.TraceElement2 – *trace.ElementChannel, one per CDNode in cycle order
//	lockElems[i] and chanElems[i] are in the same goroutine.
//
// Cycle structure: [CD_0, RD_0, CD_1, RD_1, ..., CD_{n-1}, RD_{n-1}]
//
//	lockElems[i] = RD_i's lock acquire element
//	chanElems[i] = CD_i's channel op element
//	Inter-thread edge: RD_i -> CD_{(i+1)%n}
//	  goroutine i needs the lock held/needed by goroutine (i+1)%n.
//
// Deadlocking schedule:
//
//	The cycle is reported with the canonical root (smallest tPre CD),
//	producing a single rewrite per unique cycle.
//
//	We identify for each pair (i, (i+1)%n):
//	  Holder = goroutine (i+1)%n: holds lock while its channel op blocks.
//	  Waiter = goroutine i: tries to acquire lock, blocks before its channel op.
//
//	For CS holder: holds lock during channel op -> chan op blocks (no partner/
//	               empty buffer) -> keep acq(x) + chan_op, set chan_op.tPost=0.
//	For PCS waiter: needs to re-acquire lock before reaching channel op ->
//	                blocks on acq(x) -> keep only acq(x), set acq.tPost=0,
//	                remove chan_op from trace.
//
//	For n=2 with one CS and one PCS:
//	  MD2-2 (sender CS, receiver PCS):
//	    Sender is holder (CS): keep acq(x) + snd(c), snd.tPost=0 (blocks, no recv)
//	    Receiver is waiter (PCS): keep acq(x) only, acq.tPost=0 (blocks on lock)
//	  MD2-3 (sender PCS, receiver CS):
//	    Receiver is holder (CS): keep acq(x) + rcv(c), rcv.tPost=0
//	    Sender is waiter (PCS): keep acq(x) only, acq.tPost=0
//
//	Ordering: shift holder's elements to run before waiter's acq(x).
//
// tPost=0 sentinel: GetTSort()=MaxInt -> replayer blocks at this operation.
func rewriteMixedDeadlock(tr *trace.Trace, bug bugs.Bug, code int) error {
	lockElems := bug.TraceElement1
	chanElems := bug.TraceElement2

	if len(lockElems) == 0 {
		return errors.New("rewriteMixedDeadlock: no lock elements in bug")
	}
	if len(chanElems) == 0 {
		return errors.New("rewriteMixedDeadlock: no channel elements in bug")
	}
	if len(lockElems) != len(chanElems) {
		return fmt.Errorf("rewriteMixedDeadlock: lock count %d != channel count %d",
			len(lockElems), len(chanElems))
	}

	n := len(lockElems)

	// -----------------------------------------------------------------------
	// Step 1: find latest tPost, shorten overall trace.
	// -----------------------------------------------------------------------
	lastTime := 0
	for _, e := range lockElems {
		if e.GetTPost() > lastTime {
			lastTime = e.GetTPost()
		}
	}
	for _, e := range chanElems {
		if e.GetTPost() > lastTime {
			lastTime = e.GetTPost()
		}
	}
	fmt.Println("rewriteMixedDeadlock: lastTime =", lastTime)
	tr.ShortenTrace(lastTime, true)

	// -----------------------------------------------------------------------
	// Step 2: for each inter-thread pair (waiter i -> holder (i+1)%n),
	// shorten and mark blocking elements.
	//
	// Waiter goroutine i:
	//   - Keep only acq(x) = lockElems[i], remove chanElems[i] onward.
	//   - Set lockElems[i].tPost = 0 (blocks acquiring lock).
	//
	// Holder goroutine (i+1)%n:
	//   - Keep acq(x) = lockElems[(i+1)%n] AND chanElems[(i+1)%n].
	//   - Set chanElems[(i+1)%n].tPost = 0 (blocks on channel op).
	//
	// For n=2 each goroutine plays both roles. We process pair 0 only
	// (waiter=goroutine 0, holder=goroutine 1). This produces one consistent
	// deadlocking schedule. The canonical root deduplication ensures only one
	// rewrite is generated per unique cycle, making the second pair redundant.
	// -----------------------------------------------------------------------
	for i := 0; i < n; i++ {
		holderIdx := (i + 1) % n

		waiterLock := lockElems[i]
		waiterChan := chanElems[i]
		holderLock := lockElems[holderIdx]
		holderChan := chanElems[holderIdx]

		waiterRout := waiterLock.GetRoutine()
		holderRout := holderLock.GetRoutine()

		fmt.Printf("rewriteMixedDeadlock: pair %d: waiter=R%d holder=R%d\n", i, waiterRout, holderRout)

		// Shorten waiter to just before chanElems[i], keeps lockElems[i], drops rest.
		// ShortenRoutine(rout, t) removes tSort >= t.
		// waiterLock.tPost < waiterChan.tPost (same goroutine, earlier event).
		// We want to keep waiterLock but drop waiterChan+.
		// Pass waiterChan's tSort (= waiterChan.tPost since tPost≠0) to remove it.
		tr.ShortenRoutine(waiterRout, waiterChan.GetTSort())
		fmt.Printf("rewriteMixedDeadlock: waiter R%d shortened (drop chanElem tSort=%d)\n",
			waiterRout, waiterChan.GetTSort())

		// Shorten holder to just after chanElems[holderIdx] — keeps both M and C.
		// holderChan.tPost+1 removes everything after holderChan.
		tr.ShortenRoutine(holderRout, holderChan.GetTPost()+1)
		fmt.Printf("rewriteMixedDeadlock: holder R%d shortened to C.tPost=%d\n",
			holderRout, holderChan.GetTPost())

		// Mark waiterLock as blocking: tPost=0.
		setTPostZero(waiterLock)
		fmt.Printf("rewriteMixedDeadlock: waiterLock R%d tPre=%d tPost=%d\n",
			waiterRout, waiterLock.GetTPre(), waiterLock.GetTPost())

		// Mark holderChan as blocking: tPost=0.
		setTPostZero(holderChan)
		fmt.Printf("rewriteMixedDeadlock: holderChan R%d tPre=%d tPost=%d\n",
			holderRout, holderChan.GetTPre(), holderChan.GetTPost())
	}

	// -----------------------------------------------------------------------
	// Step 3: reorder, holder's elements before waiter's lock acquire.
	//
	// For each pair i, assign holder's elements tPre values just before the
	// waiter's lock acquire tPre. This forces the holder to acquire the lock
	// first in the replay ordering.
	//
	// holderLock gets a real tPost (non-zero) so it sorts by tPost and
	// appears before the waiter. holderChan has tPost=0 so it's at MaxInt
	// (end of trace) — the replayer will reach it and block there.
	//
	// We use SetTPre (not SetTSort) on holderLock to preserve its tPost=non-zero:
	//   - SetTPre only adjusts tPost upward if tPost < tPre.
	//   - We keep holderLock.tPost as-is (it completed in working trace).
	//   - We only change holderLock.tPre to the new ordering position.
	// -----------------------------------------------------------------------
	for i := 0; i < n; i++ {
		holderIdx := (i + 1) % n

		waiterLock := lockElems[i]
		holderLock := lockElems[holderIdx]
		holderRout := holderLock.GetRoutine()

		waiterTPre := waiterLock.GetTPre()
		holderElems := tr.GetRoutineTrace(holderRout)
		nHolder := len(holderElems)

		if nHolder == 0 {
			fmt.Printf("rewriteMixedDeadlock: holder R%d has empty trace — skip reorder\n", holderRout)
			continue
		}

		// Ensure room before waiterTPre for nHolder elements.
		if waiterTPre <= nHolder {
			shift := nHolder - waiterTPre + 1
			tr.ShiftTrace(waiterTPre, shift)
			// Update waiterLock's tPre reference after shift.
			waiterTPre = waiterLock.GetTPre() // re-read after shift
			fmt.Printf("rewriteMixedDeadlock: ShiftTrace by %d, new waiterTPre=%d\n",
				shift, waiterTPre)
		}

		// Assign holder elements new tPre: [waiterTPre-nHolder .. waiterTPre-1].
		// - Non-blocking elements (tPost≠0): use SetT to set both tPre and tPost
		//   so they sort correctly by tPost.
		// - Blocking elements (tPost=0): use SetTPre only — tPost stays 0,
		//   keeping GetTSort()=MaxInt.
		newT := waiterTPre - nHolder
		for _, hElem := range holderElems {
			oldTPre := hElem.GetTPre()
			if hElem.GetTPost() == 0 {
				hElem.SetTPre(newT)
			} else {
				hElem.SetT(newT)
			}
			fmt.Printf("rewriteMixedDeadlock: holder R%d elem tPre %d->%d tPost=%d\n",
				holderRout, oldTPre, newT, hElem.GetTPost())
			newT++
		}
	}

	// -----------------------------------------------------------------------
	// Step 4: insert replay exit-code marker.
	// -----------------------------------------------------------------------
	tr.AddTraceElementReplay(lastTime+1, code)

	fmt.Printf("rewriteMixedDeadlock: done. Marker at %d code %d (ExitCodeMixedDeadlock=%d)\n",
		lastTime+1, code, helper.ExitCodeMixedDeadlock)

	return nil
}

// setTPostZero sets tPost=0 on an element while preserving tPre.
//
// tPost=0 is ADVOCATE's sentinel for "operation never completed" ->
// GetTSort()=MaxInt -> replayer blocks at this operation.
//
// Method:
//
//	SetTSort(0) sets tPre=tPost=0.
//	SetTPre(savedTPre) restores tPre; since tPost==0, SetTPre's guard
//	(tPost!=0 && tPost<tPre) does NOT fire, so tPost stays at 0.
func setTPostZero(elem trace.Element) {
	savedTPre := elem.GetTPre()
	elem.SetTSort(0)        // tPre=0, tPost=0
	elem.SetTPre(savedTPre) // restore tPre; tPost stays 0
}
