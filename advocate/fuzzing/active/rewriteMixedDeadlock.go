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
	//	mainRout, holderRout, lockHolder.GetTPre(), cdHolder.GetTPre(), waiterRout, lockWaiter.GetTPre())

	lastTime := max(lockHolder.GetTPost(), lockWaiter.GetTPost())
	if mainTrace := tr.GetRoutineTrace(mainRout); len(mainTrace) > 0 {
		if lastElem := mainTrace[len(mainTrace)-1]; lastElem.GetTPost() > lastTime {
			lastTime = lastElem.GetTPost()
		}
	}
	tr.ShortenTrace(lastTime, true)

	tr.ShortenRoutine(holderRout, cdHolder.GetTPost()+1)
	tr.ShortenRoutine(waiterRout, lockWaiter.GetTPost()+1)

	//fmt.Printf("rewriteMixedDeadlock: holder R%d kept to t=%d, waiter R%d kept to t=%d\n",
	//	holderRout, cdHolder.GetTPost(), waiterRout, lockWaiter.GetTPost())

	// Reorder
	if lockWaiter.GetTPre() < lockHolder.GetTPre() {
		targetTPre := lockHolder.GetTPost() + 1
		shift := targetTPre - lockWaiter.GetTPre()
		if shift > 0 {
			waiterTrace := tr.GetRoutineTrace(waiterRout)
			if len(waiterTrace) == 0 {
				return fmt.Errorf("rewriteMixedDeadlock: waiter R%d has no trace", waiterRout)
			}
			startElem := waiterTrace[0]
			startTPre := startElem.GetTPre()
			//fmt.Printf("rewriteMixedDeadlock: shifting waiter R%d by %d\n", waiterRout, shift)
			tr.ShiftRoutine(waiterRout, startTPre, shift)
			tr.ShiftConcurrentOrAfterToAfter(lockWaiter)
		}
	}

	// Ensure holder's channel op is after lock acquire
	if cdHolder.GetTPre() <= lockHolder.GetTPre() {
		newTPre := lockHolder.GetTPost() + 1
		cdHolder.SetTPre(newTPre)
		cdHolder.SetTPost(newTPre)
	}

	// Clear channel state
	forceChannelBlock(cdHolder)

	// Calculate final time
	newLastTime := 0
	for _, traceSlice := range tr.GetTraces() {
		for _, elem := range traceSlice {
			t := elem.GetTSort()
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
	savedTPre := elem.GetTPre()
	// SetTWithoutNotExecuted sets tPost=0 ONLY if the original tPost was non-zero
	// element will be marked as "never completed"
	elem.SetTWithoutNotExecuted(0)
	elem.SetTPre(savedTPre)
	//fmt.Printf("rewriteMixedDeadlock: blocked element %T (tPre=%d, tPost=0)\n", elem, savedTPre)
}

func forceChannelBlock(ch *trace.ElementChannel) {
	ch.SetPartner(nil)
	ch.SetOID(-1)
	ch.SetQCount(0)
	//fmt.Printf("rewriteMixedDeadlock: forceChannelBlock on ch=%d\n", ch.GetObjId())
}
