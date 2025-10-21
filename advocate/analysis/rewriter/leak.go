// Copyright (c) 2024 Erik Kassubek
//
// File: leak.go
// Brief: Rewrite trace for leaked channel
//
// Author: Erik Kassubek
// Created: 2024-04-07
//
// License: BSD-3-Clause

package rewriter

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/bugs"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"errors"
)

// Rewrite a trace where a leaking routine was found.
// Different to most other rewrites, we don not try to get the program to run
// into a possible bug, but to take an actual leak (we only detect actual leaks,
// not possible leaks) and rewrite them in such a way, that the routine
// gets unstuck, meaning is not leaking any more.
// We detect leaks, that are stuck because of the following conditions:
//  - channel operation without a possible  partner (may be in select)
//  - channel operation with a possible partner, but no communication (may be in select)
//  - mutex operation without a post event
//  - wait group operation without a post event
//  - cond operation without a post event

// =============== Channel/Select ====================

// Rewrite a trace where a leaking unbuffered channel/select with possible partner was found.
//
// Parameter:
//   - tr *trace.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteUnbufChanLeak(tr *trace.Trace, bug bugs.Bug) error {
	// check if one or both of the bug elements are select
	t1Sel := false
	t2Sel := false
	switch bug.TraceElement1[0].(type) {
	case *trace.ElementSelect:
		t1Sel = true
	}
	switch bug.TraceElement2[0].(type) {
	case *trace.ElementSelect:
		t2Sel = true
	}

	if !t1Sel && !t2Sel { // both are channel operations
		return rewriteUnbufChanLeakChanChan(tr, bug)
	} else if !t1Sel && t2Sel { // first is channel operation, second is select
		return rewriteUnbufChanLeakChanSel(tr, bug)
	} else if t1Sel && !t2Sel { // first is select, second is channel operation
		return rewriteUnbufChanLeakSelChan(tr, bug)
	} // both are select
	return rewriteUnbufChanLeakSelSel(tr, bug)
}

// Rewrite a trace where a leaking unbuffered channel/select with possible partner was found
// if both elements are channel operations.
//
// Parameter:
//   - tr *analysis.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteUnbufChanLeakChanChan(tr *trace.Trace, bug bugs.Bug) error {
	stuck := bug.TraceElement1[0].(*trace.ElementChannel)
	possiblePartner := bug.TraceElement2[0].(*trace.ElementChannel)
	possiblePartnerPartner := possiblePartner.GetPartner()

	if possiblePartnerPartner != nil {
		hbInfo := clock.GetHappensBefore(possiblePartnerPartner.GetVC(), stuck.GetVC())
		if hbInfo == hb.Before {
			return errors.New("The actual partner of the potential partner is HB " +
				"before to the stuck element. Cannot rewrite trace.")
		}
	}

	// T = T1 ++ [f] ++ T2 ++ [g] ++ T3 ++ [e]

	// remove the potential partner partner from the trace
	if possiblePartnerPartner != nil {
		tr.RemoveElementFromTrace(possiblePartnerPartner.GetTID())
	}

	// T = T1 ++ [f] ++ T2 ++ T3 ++ [e]

	if stuck.GetType(true) == trace.ChannelRecv { // Case 3
		tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartner.GetTSort()) // bug.TraceElement1[0] = stuck

		// T = T1 ++ [f] ++ T2' ++ T3' ++ [e]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]

		// add replay signals
		tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)

	} else { // Case 4
		if possiblePartnerPartner != nil {
			tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartnerPartner.GetTSort()) // bug.TraceElement1[0] = stuck
		} else {
			tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], 0) // bug.TraceElement1[0] = stuck
		}

		// T = T1 ++ T2' ++ T3' ++ [e] ++ T4 ++ [f]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
		// and T4 = [h in T4 | h >= e]

		tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement2[0], stuck.GetTSort()) // bug.TraceElement2[0] = possiblePartner

		// T = T1 ++ T2' ++ T3' ++ [e] ++ T4' ++ [f]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
		// and T4' = [h in T4 | h >= e and h < f]

		// add replay signal
		tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)
	}

	return nil
}

// Rewrite a trace where a leaking unbuffered channel/select with possible partner was found
// if a channel is stuck and a select is a possible partner
//
// Parameter:
//   - tr *analysis.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteUnbufChanLeakChanSel(tr *trace.Trace, bug bugs.Bug) error {
	stuck := bug.TraceElement1[0].(*trace.ElementChannel)
	possiblePartner := bug.TraceElement2[0].(*trace.ElementSelect)
	possiblePartnerPartner := possiblePartner.GetPartner()

	if possiblePartnerPartner != nil {
		hbInfo := clock.GetHappensBefore(possiblePartnerPartner.GetVC(), stuck.GetVC())
		if hbInfo == hb.Before {
			return errors.New("The actual partner of the potential partner is not HB " +
				"concurrent to the stuck element. Cannot rewrite trace.")
		}
	}

	// T = T1 ++ [f] ++ T2 ++ [g] ++ T3 ++ [e]

	// remove the potential partner partner from the trace
	if possiblePartnerPartner != nil {
		tr.RemoveElementFromTrace(possiblePartnerPartner.GetTID())
	}

	// T = T1 ++ [f] ++ T2 ++ T3 ++ [e]

	if stuck.GetType(true) == trace.ChannelRecv { // Case 3
		tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartner.GetTSort()) // bug.TraceElement1[0] = stuck

		// T = T1 ++ [f] ++ T2' ++ T3' ++ [e]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]

		err := bug.TraceElement2[0].(*trace.ElementSelect).SetCase(stuck.GetID(), trace.ChannelSend)
		if err != nil {
			println(err.Error())
		}

		// add replay signal
		tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)

	} else { // Case 4
		if possiblePartnerPartner != nil {
			tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartnerPartner.GetTSort()) // bug.TraceElement1[0] = stuck
		} else {
			tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], 0) // bug.TraceElement1[0] = stuck
		}

		// T = T1 ++ T2' ++ T3' ++ [e] ++ T4 ++ [f]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
		// and T4 = [h in T4 | h >= e]

		tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement2[0], stuck.GetTSort()) // bug.TraceElement2[0] = possiblePartner

		// T = T1 ++ T2' ++ T3' ++ [e] ++ T4' ++ [f]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
		// and T4' = [h in T4 | h >= e and h < f]

		err := bug.TraceElement2[0].(*trace.ElementSelect).SetCase(stuck.GetID(), trace.ChannelRecv)
		if err != nil {
			println(err.Error())
		}

		// add replay signal
		tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)
	}

	return nil
}

// Rewrite a trace where a leaking unbuffered channel/select with possible partner was found
// if a select is stuck and a channel is a possible partner
//
// Parameter:
//   - tr *trace.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteUnbufChanLeakSelChan(tr *trace.Trace, bug bugs.Bug) error {
	stuck := bug.TraceElement1[0].(*trace.ElementSelect)
	possiblePartner := bug.TraceElement2[0].(*trace.ElementChannel)
	possiblePartnerPartner := possiblePartner.GetPartner()

	if possiblePartnerPartner != nil {
		hbInfo := clock.GetHappensBefore(possiblePartnerPartner.GetVC(), stuck.GetVC())
		if hbInfo == hb.Before {
			return errors.New("The actual partner of the potential partner is HB " +
				"before to the stuck element. Cannot rewrite trace.")
		}
	}

	// T = T1 ++ [f] ++ T2 ++ [g] ++ T3 ++ [e]

	// remove the potential partner partner from the trace
	if possiblePartnerPartner != nil {
		tr.RemoveElementFromTrace(possiblePartnerPartner.GetTID())
	}

	// T = T1 ++ [f] ++ T2 ++ T3 ++ [e]

	if possiblePartner.GetType(true) == trace.ChannelRecv {
		if possiblePartnerPartner != nil {
			tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartnerPartner.GetTSort()) // bug.TraceElement1[0] = stuck
		} else {
			tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], 0) // bug.TraceElement1[0] = stuck
		}

		// T = T1 ++ T2' ++ T3' ++ [e] ++ T4 ++ [f]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
		// and T4 = [h in T4 | h >= e]

		tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement2[0], stuck.GetTSort()) // bug.TraceElement2[0] = possiblePartner

		err := bug.TraceElement1[0].(*trace.ElementSelect).SetCase(stuck.GetID(), trace.ChannelSend)
		if err != nil {
			println(err.Error())
		}

		// T = T1 ++ T2' ++ T3' ++ [e] ++ T4' ++ [f]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
		// and T4' = [h in T4 | h >= e and h < f]
		// add replay signals
		tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)

	} else { // Case 3
		tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartner.GetTSort()) // bug.TraceElement1[0] = stuck

		// T = T1 ++ [f] ++ T2' ++ T3' ++ [e]
		// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]

		err := bug.TraceElement1[0].(*trace.ElementSelect).SetCase(stuck.GetID(), trace.ChannelRecv)
		if err != nil {
			println(err.Error())
		}

		// add replay signal
		tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)

	}

	return nil
}

// Rewrite a trace where a leaking unbuffered channel/select with possible partner was found
// if both elements are select operations.
//
// Parameter:
//   - tr *trace.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteUnbufChanLeakSelSel(tr *trace.Trace, bug bugs.Bug) error {
	stuck := bug.TraceElement1[0].(*trace.ElementSelect)
	possiblePartner := bug.TraceElement2[0].(*trace.ElementSelect)
	possiblePartnerPartner := possiblePartner.GetPartner()

	if possiblePartnerPartner != nil {
		hbInfo := clock.GetHappensBefore(possiblePartnerPartner.GetVC(), stuck.GetVC())
		if hbInfo == hb.Before {
			return errors.New("The actual partner of the potential partner is HB " +
				"before to the stuck element. Cannot rewrite trace.")
		}
	}

	// T = T1 ++ [f] ++ T2 ++ [g] ++ T3 ++ [e]

	// remove the potential partner partner from the trace
	if possiblePartnerPartner != nil {
		tr.RemoveElementFromTrace(possiblePartnerPartner.GetTID())
	}

	// find communication
	for _, c := range stuck.GetCases() {
		for _, d := range possiblePartner.GetCases() {
			if c.GetID() != d.GetID() {
				continue
			}

			if c.GetType(true) == d.GetType(true) {
				continue
			}

			// T = T1 ++ [f] ++ T2 ++ T3 ++ [e]

			if c.GetType(true) == trace.ChannelRecv { // Case 3
				tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartner.GetTSort()) // bug.TraceElement1[0] = stuck

				// T = T1 ++ [f] ++ T2' ++ T3' ++ [e]
				// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]

				err := bug.TraceElement1[0].(*trace.ElementSelect).SetCase(c.GetID(), trace.ChannelRecv)
				if err != nil {
					println(err.Error())
				}
				err = bug.TraceElement2[0].(*trace.ElementSelect).SetCase(d.GetID(), trace.ChannelSend)
				if err != nil {
					println(err.Error())
				}

				// add replay signal
				tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)
				return nil
			}

			// Case 4
			baseA.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement1[0], possiblePartner.GetTSort()) // bug.TraceElement1[0] = stuck

			// T = T1 ++ T2' ++ T3' ++ [e] ++ T4 ++ [f]
			// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
			// and T4 = [h in T4 | h >= e]

			tr.ShiftConcurrentOrAfterToAfterStartingFromElement(bug.TraceElement2[0], stuck.GetTSort()) // bug.TraceElement2[0] = possiblePartner

			// T = T1 ++ T2' ++ T3' ++ [e] ++ T4' ++ [f]
			// where T2' = [h in T2 | h < e] and T3' = [h in T3 | h < e]
			// and T4' = [h in T4 | h >= e and h < f]

			err := bug.TraceElement1[0].(*trace.ElementSelect).SetCase(c.GetID(), trace.ChannelSend)
			if err != nil {
				println(err.Error())
			}
			err = bug.TraceElement2[0].(*trace.ElementSelect).SetCase(d.GetID(), trace.ChannelRecv)
			if err != nil {
				println(err.Error())
			}

			// add replay signals
			tr.AddTraceElementReplay(max(bug.TraceElement1[0].GetTSort(), bug.TraceElement2[0].GetTSort())+1, helper.ExitCodeLeakUnbuf)

			return nil
		}
	}

	return errors.New("could not establish communication between two selects. Cannot rewrite trace")
}

// Rewrite a trace for a leaking buffered channel
//
// Parameter:
//   - tr *trace.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteBufChanLeak(tr *trace.Trace, bug bugs.Bug) error {
	stuck := bug.TraceElement1[0]
	possiblePartner := bug.TraceElement2[0]
	var possiblePartnerPartner *trace.ElementChannel
	switch z := possiblePartner.(type) {
	case *trace.ElementChannel:
		possiblePartnerPartner = z.GetPartner()
	case *trace.ElementSelect:
		possiblePartnerPartner = z.GetPartner()
	}

	if possiblePartnerPartner != nil {
		hbInfo := clock.GetHappensBefore(possiblePartnerPartner.GetVC(), stuck.GetVC())
		if hbInfo == hb.Before {
			return errors.New("The actual partner of the potential partner is HB " +
				"before to the stuck element. Cannot rewrite trace.")
		}
	}

	if possiblePartnerPartner != nil {
		// T = T1 ++ T2 ++ [e]
		tr.RemoveElementFromTrace(possiblePartnerPartner.GetTID())

		// T = T1 ++ T2' ++ [e]
		// where T2' = [ h | h in T2 and h <HB e]
		tr.ShiftConcurrentOrAfterToAfterStartingFromElement(stuck, possiblePartnerPartner.GetTSort())
	}

	bug.TraceElement1[0].SetTSort(possiblePartner.GetTSort() + 1)

	if possiblePartner.GetTSort() < stuck.GetTSort() {
		tr.AddTraceElementReplay(stuck.GetTSort()+1, helper.ExitCodeLeakBuf)
	} else {
		tr.AddTraceElementReplay(possiblePartner.GetTSort()+1, helper.ExitCodeLeakBuf)
	}

	return nil
}

// ================== Mutex ====================

// Rewrite a trace where a leaking mutex was found.
// The trace can only be rewritten, if the stuck lock operation is concurrent
// with the last lock operation on this mutex. If it is not concurrent, the
// rewrite fails. If a rewrite is possible, we try to run the stock lock operation
// before the last lock operation, so that the mutex is not blocked anymore.
// We therefore rewrite the trace from
//
// T_1 + [l'] + T_2 + [l] + T_3
//
// to
//
// T_1' + T_2' + [X_s, l, X_e]
//
// where l is the stuck lock, l' is the last lock, T_1, T_2, T_3 are the traces
// before, between and after the locks, T_1' and T_2' are the elements from T_1 and T_2, that
// are before (HB) l, X_s is the start and X_e is the stop signal, that releases the program from the
// guided replay.
//
// Parameter:
//   - tr *analysis.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteMutexLeak(tr *trace.Trace, bug bugs.Bug) error {
	log.Info("Start rewriting trace for mutex leak...")

	// get l and l'
	lockOp := bug.TraceElement1[0].(*trace.ElementMutex)
	lastLockOp := bug.TraceElement2[0].(*trace.ElementMutex)

	hbInfo := clock.GetHappensBefore(lockOp.GetVC(), lastLockOp.GetVC())
	if hbInfo != hb.Concurrent {
		return errors.New("the stuck mutex lock is not concurrent with the prior lock. Cannot rewrite trace")
	}

	// remove T_3 -> T_1 + [l'] + T_2 + [l]
	tr.ShortenTrace(lockOp.GetTSort(), true)

	// remove all elements, that are concurrent with l. This includes l'
	// -> T_1' + T_2' + [l]
	tr.RemoveConcurrent(bug.TraceElement1[0], 0)

	// set tPost of l to non zero
	lockOp.SetT(lockOp.GetTPre())

	// add the start and stop signal after l -> T_1' + T_2' + [X_s, l, X_e]
	tr.AddTraceElementReplay(lockOp.GetTPre()+1, helper.ExitCodeLeakMutex)

	return nil
}

// ================== WaitGroup ====================

// Rewrite a trace where a leaking waitgroup was found.
//
// Parameter:
//   - tr *trace.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteWaitGroupLeak(tr *trace.Trace, bug bugs.Bug) error {
	log.Info("Start rewriting trace for waitgroup leak...")

	wait := bug.TraceElement1[0]

	if len(bug.TraceElement2) == 0 {
		return errors.New("no possible partner to move. Cannot rewrite trace")
	}

	tr.ShiftConcurrentOrAfterToAfter(wait)

	tr.AddTraceElementReplay(wait.GetTPre()+1, helper.ExitCodeLeakWG)

	nrAdd, nrDone := tr.GetNrAddDoneBeforeTime(wait.GetID(), wait.GetTSort())

	if nrAdd != nrDone {
		return errors.New("the wait group is not balanced. Cannot rewrite trace")
	}

	return nil
}

// ================== Cond ====================

// Rewrite a trace where a leaking cond was found.
//
// Parameter:
//   - tr *trace.Trace: The trace to rewrite
//   - bug Bug: The bug to create a trace for
//
// Returns:
//   - error: An error if the trace could not be created
func rewriteCondLeak(tr *trace.Trace, bug bugs.Bug) error {
	log.Info("Start rewriting trace for cond leak...")

	couldRewrite := false

	wait := bug.TraceElement1[0]

	res := tr.GetConcurrentWaitGroups(wait)

	// possible signals to release the wait
	if len(res["signal"]) > 0 {
		couldRewrite = true

		wait.SetT(wait.GetTPre())

		// move the signal after the wait
		tr.ShiftConcurrentOrAfterToAfter(wait)
	}

	// possible broadcasts to release the wait
	for _, broad := range res["broadcast"] {
		couldRewrite = true
		tr.ShiftConcurrentToBefore(broad)
	}

	wait.SetT(wait.GetTPre())

	if len(bug.TraceElement2) == 0 {
		tr.AddTraceElementReplay(wait.GetTPre()+1, helper.ExitCodeLeakCond)
	} else {
		tr.AddTraceElementReplay(wait.GetTPre()+1, helper.ExitCodeLeakCond)
	}

	if couldRewrite {
		return nil
	}

	return errors.New("Could not rewrite trace for cond leak")
}
