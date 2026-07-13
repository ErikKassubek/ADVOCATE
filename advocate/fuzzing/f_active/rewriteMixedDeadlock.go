// advocate/fuzzing/active/rewriteMixedDeadlock.go
//
// Copyright (c) 2024 Erik Kassubek
//
// File: rewriteMixedDeadlock.go
//
// License: BSD-3-Clause

package f_active

import (
	"advocate/trace"
	"advocate/utils/results/bugs"
	"errors"
	"fmt"
	"math"
)

func rewriteMixedDeadlock(tr *trace.Trace, bug bugs.Bug, code int) error {
	if len(bug.TraceElement2) != 4 {
		return fmt.Errorf("rewriteMixedDeadlock: expected 4 elements, got %d", len(bug.TraceElement2))
	}

	cdHolder, ok1 := bug.TraceElement2[0].(*trace.ElementChannel)
	lockHolder, ok2 := bug.TraceElement2[1].(*trace.ElementMutex)
	_, ok3 := bug.TraceElement2[2].(*trace.ElementChannel)
	lockWaiter, ok4 := bug.TraceElement2[3].(*trace.ElementMutex)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return errors.New("rewriteMixedDeadlock: incorrect element types in cycle")
	}

	holderRout := cdHolder.GetRoutine()
	waiterRout := lockWaiter.GetRoutine()

	mainRout := 0
	for rid := range tr.GetTraces() {
		if rid != int(holderRout) && rid != int(waiterRout) {
			if mainRout == 0 || rid < mainRout {
				mainRout = rid
			}
		}
	}
	if mainRout == 0 {
		mainRout = 1
	}

	//fmt.Printf("rewriteMixedDeadlock: main=R%d, holder=R%d (lock=%d, chan=%d), waiter=R%d (lock=%d)\n",
	//	mainRout, holderRout, lockHolder.GetT(tPre, ), cdHolder.GetT(tPre, ), waiterRout, lockWaiter.GetT(tPre, ))

	lastTime := max(lockHolder.GetT(trace.Commit), lockWaiter.GetT(trace.Commit))
	if mainTrace := tr.GetRoutineTrace(mainRout); mainTrace.Len() > 0 {
		if lastElem := mainTrace.Last(); lastElem.GetT(trace.Commit) > lastTime {
			lastTime = lastElem.GetT(trace.Commit)
		}
	}
	tr.ShortenTrace(lastTime, true)

	tr.ShortenRoutine(holderRout, cdHolder.GetT(trace.Commit)+1)
	tr.ShortenRoutine(waiterRout, lockWaiter.GetT(trace.Commit)+1)

	//fmt.Printf("rewriteMixedDeadlock: holder R%d kept to t=%d, waiter R%d kept to t=%d\n",
	//	holderRout, cdHolder.GetT(tPost, ), waiterRout, lockWaiter.GetT(tPost, ))

	// Reorder
	if lockWaiter.GetT(trace.Request) < lockHolder.GetT(trace.Request) {
		targetTPre := lockHolder.GetT(trace.Request) + 1
		shift := targetTPre - lockWaiter.GetT(trace.Request)
		if shift > 0 {
			waiterTrace := tr.GetRoutineTrace(waiterRout)
			if waiterTrace.Empty() {
				return fmt.Errorf("rewriteMixedDeadlock: waiter R%d has no trace", waiterRout)
			}
			startElem := waiterTrace.First()
			startTPre := startElem.GetT(trace.Request)
			//fmt.Printf("rewriteMixedDeadlock: shifting waiter R%d by %d\n", waiterRout, shift)
			tr.ShiftRoutine(waiterRout, startTPre, shift)
			tr.ShiftConcurrentOrAfterToAfter(lockWaiter)
		}
	}

	// Ensure holder's channel op is after lock acquire
	if cdHolder.GetT(trace.Request) <= lockHolder.GetT(trace.Request) {
		newTPre := lockHolder.GetT(trace.Commit) + 1
		cdHolder.SetT(trace.Request, newTPre)
		cdHolder.SetT(trace.Commit, newTPre)
	}

	// Clear channel state
	forceChannelBlock(cdHolder)

	// Calculate final time
	newLastTime := 0
	for _, routSlice := range tr.GetTraces() {
		for _, elem := range routSlice.Elems() {
			t := elem.GetT(trace.Sorting)
			if t > newLastTime && t != math.MaxInt {
				newLastTime = t
			}
		}
		//fmt.Printf("rewriteMixedDeadlock: R%d has %d elements, last t=%d\n",
		//	rid, len(traceSlice), newLastTime)
	}

	// Use SetTWithoutNotExecuted to set tPost=0 while preserving tPre behavior
	blockElement(cdHolder)
	blockElement(lockWaiter)

	tr.AddTraceElementReplay(newLastTime+1, code)
	//fmt.Printf("rewriteMixedDeadlock: replay marker at t=%d, code=%d\n", newLastTime+1, code)

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// blockElement to force an element to block (tPost=0) while preserving tPre
func blockElement(elem trace.Element) {
	savedTPre := elem.GetT(trace.Request)
	// SetTWithoutNotExecuted sets tPost=0 ONLY if the original tPost was non-zero
	// element will be marked as "never completed"
	elem.SetTWithoutNotExecuted(0)
	elem.SetT(trace.Request, savedTPre)
	//fmt.Printf("rewriteMixedDeadlock: blocked element %T (tPre=%d, tPost=0)\n", elem, savedTPre)
}

func forceChannelBlock(ch *trace.ElementChannel) {
	ch.SetPartner(nil)
	ch.SetOID(-1)
	ch.SetQCount(0)
	//fmt.Printf("rewriteMixedDeadlock: forceChannelBlock on ch=%d\n", ch.GetObjId())
}
