// Copyright (c) 2024 Erik Kassubek
//
// File: /rewriter/mixedDeadlock.go
// Brief: Main functions to rewrite the trace
//
// Author: Erik Kassubek
// Created: 2023-11-30
//
// License: BSD-3-Clause

package rewriter

import (
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/bugs"
	"advocate/trace"
	"errors"
	"fmt"
)

// rewriteMixedDeadlock rewrites the trace to confirm a P06 mixed deadlock by
// reversing the HB order of the two critical lock acquires on the same mutex.
//
// Parameters:
//   - tr *trace.Trace: The trace to rewrite
//   - bug bugs.Bug: The mixed deadlock bug report
//
// Returns:
//   - error: An error if the rewrite fails
func rewriteMixedDeadlock(tr *trace.Trace, bug bugs.Bug, exitCode int) error {
	if len(bug.TraceElement1) == 0 && len(bug.TraceElement2) == 0 {
		return errors.New("mixed deadlock rewrite: no bug elements")
	}

	// Find acquire pair (shared lock)
	acqE, acqF, err := pickAcquirePair(tr, bug)
	if err != nil {
		return err
	}

	if acqE.GetID() != acqF.GetID() {
		return fmt.Errorf("acquires are on different mutex IDs (%d vs %d)",
			acqE.GetID(), acqF.GetID())
	}

	// Compute timeline frontier
	lastTime := acqE.GetTSort()
	if acqF.GetTSort() > lastTime {
		lastTime = acqF.GetTSort()
	}
	for _, e := range append(bug.TraceElement1, bug.TraceElement2...) {
		if e.GetTSort() > lastTime {
			lastTime = e.GetTSort()
		}
	}

	// Shorten trace
	tr.ShortenTrace(lastTime, true)
	tr.ShortenRoutine(acqE.GetRoutine(), acqE.GetTSort())
	tr.ShortenRoutine(acqF.GetRoutine(), acqF.GetTSort())

	// Early unlocks
	lockElems := []trace.Element{acqE, acqF}
	ensureEarlyUnlocks(tr, lockElems, &lastTime)

	// Reverse HB order of acquires
	switch clock.GetHappensBefore(acqE.GetWVC(), acqF.GetWVC()) {
	case hb.Before:
		if err := advanceRoutineUntilBefore(tr, acqF, acqE); err != nil {
			return fmt.Errorf("cannot make receiver acquire before sender: %w", err)
		}
	case hb.After:
		if err := advanceRoutineUntilBefore(tr, acqE, acqF); err != nil {
			return fmt.Errorf("cannot make sender acquire before receiver: %w", err)
		}
	default:
		// concurrent: minimal swap for determinism
		if acqE.GetTSort() < acqF.GetTSort() {
			_ = advanceRoutineUntilBefore(tr, acqF, acqE)
		} else {
			_ = advanceRoutineUntilBefore(tr, acqE, acqF)
		}
	}

	// Handle Close–Receive variant
	if len(bug.TraceElement1) > 0 && len(bug.TraceElement2) > 0 {
		if chClose, ok1 := bug.TraceElement1[0].(*trace.ElementChannel); ok1 &&
			chClose.GetType(true) == trace.ChannelClose {
			if recv, ok2 := bug.TraceElement2[0].(*trace.ElementChannel); ok2 &&
				recv.GetType(true) == trace.ChannelRecv {
				// close happens after recv to provoke blocking state
				if chClose.GetTSort() < recv.GetTSort() {
					_ = advanceRoutineUntilBefore(tr, recv, chClose)
				}
			}
		}
	}

	// Replay marker (42 = confirmed MD / A10)
	tr.AddTraceElementReplay(lastTime+1, exitCode)

	fmt.Printf("[rewriteMixedDeadlock] reversed acquire order for mutex %d: %d<->%d\n",
		acqE.GetID(), acqE.GetRoutine(), acqF.GetRoutine())

	return nil
}

// pickAcquirePair returns the two lock-acquire elements participating in the mixed deadlock.
// It first tries to find them directly in the bug report; if not found, it derives
// them from the channel operations in the bug report.
//
// Parameters:
//   - tr *trace.Trace: The trace containing the events
//   - bug bugs.Bug: The mixed deadlock bug report
//
// Returns:
//   - *trace.ElementMutex: The first lock acquire element
//   - *trace.ElementMutex: The second lock acquire element
//   - error: An error if the acquires cannot be determined
func pickAcquirePair(tr *trace.Trace, bug bugs.Bug) (*trace.ElementMutex, *trace.ElementMutex, error) {
	// Try direct form first: two mutex acquires provided
	if len(bug.TraceElement2) >= 2 {
		a, aok := bug.TraceElement2[0].(*trace.ElementMutex)
		b, bok := bug.TraceElement2[1].(*trace.ElementMutex)
		if aok && bok && a.IsLock() && b.IsLock() {
			return a, b, nil
		}
	}

	// Else derive from channel ops (send/close in TraceElement1, recv in TraceElement2)
	if len(bug.TraceElement1) == 0 || len(bug.TraceElement2) == 0 {
		return nil, nil, errors.New("cannot derive acquires: channel ops missing in bug")
	}
	chE := bug.TraceElement1[0] // send or close
	chF := bug.TraceElement2[0] // recv

	// For each routine, find last lock-acquire before the channel op
	findPrevLock := func(anchor trace.Element) *trace.ElementMutex {
		r := anchor.GetRoutine()
		rt := tr.GetRoutineTrace(r)
		cutoff := anchor.GetTSort()
		for i := len(rt) - 1; i >= 0; i-- {
			if rt[i].GetTSort() >= cutoff {
				continue
			}
			if m, ok := rt[i].(*trace.ElementMutex); ok && m.IsLock() {
				return m
			}
		}
		return nil
	}

	a := findPrevLock(chE)
	b := findPrevLock(chF)
	if a == nil || b == nil {
		return nil, nil, errors.New("cannot derive acquires: no preceding lock found")
	}

	// If not the same mutex, find a common id by walking further back
	if a.GetID() != b.GetID() {
		return nil, nil, fmt.Errorf("derived acquires are on different mutex IDs (%d vs %d)", a.GetID(), b.GetID())
	}
	return a, b, nil
}

// ensureEarlyUnlocks modifies the trace to ensure that unlocks for the locks in lockset
// occur as early as possible, i.e., before any concurrent events in other routines.
//
// Parameters:
//   - tr *trace.Trace: The trace to modify
//   - lockset []trace.Element: The set of lock acquire elements whose unlocks to ensure early
//   - lastTime *int: Pointer to the last time in the trace; updated if needed
func ensureEarlyUnlocks(tr *trace.Trace, lockset []trace.Element, lastTime *int) {
	hasID := func(id int) bool {
		for _, e := range lockset {
			if e.GetID() == id {
				return true
			}
		}
		return false
	}

	for _, l := range lockset {
		r := l.GetRoutine()
		for _, e := range tr.GetRoutineTrace(r) {
			if mu, ok := e.(*trace.ElementMutex); ok && !mu.IsLock() && hasID(mu.GetID()) {
				// Find first element in the other routine that is concurrent to this unlock
				// and move that routine’s tail behind unlock; also push concurrent/after to after.
				var concurrentStart trace.Element
				for _, cand := range tr.GetRoutineTrace(l.GetRoutine()) {
					if clock.GetHappensBefore(cand.GetWVC(), mu.GetWVC()) == hb.Concurrent {
						concurrentStart = cand
						break
					}
				}
				if concurrentStart != nil {
					routineEnd := tr.GetRoutineTrace(l.GetRoutine())[len(tr.GetRoutineTrace(l.GetRoutine()))-1]
					tr.ShiftRoutine(l.GetRoutine(), concurrentStart.GetTSort(), (mu.GetTSort()-concurrentStart.GetTSort())+1)
					if routineEnd.GetTPost() > *lastTime {
						*lastTime = routineEnd.GetTPost()
					}
					tr.ShiftConcurrentOrAfterToAfter(mu)
				}
			}
		}
	}
}

// advanceRoutineUntilBefore moves the routine chunk containing 'who' so that
// 'who' occurs before 'anchor' in the trace timeline.
//
// Parameters:
//   - tr *trace.Trace: The trace to modify
//   - who trace.Element: The event to move
//   - anchor trace.Element: The event before which 'who' should occur
//
// Returns:
//   - error: An error if the operation fails
func advanceRoutineUntilBefore(tr *trace.Trace, who trace.Element, anchor trace.Element) error {
	if who.GetTSort() < anchor.GetTSort() {
		return nil
	}
	// find a concurrent start in 'who' routine relative to anchor
	var start trace.Element
	for _, cand := range tr.GetRoutineTrace(who.GetRoutine()) {
		if clock.GetHappensBefore(cand.GetWVC(), anchor.GetWVC()) == hb.Concurrent {
			start = cand
			break
		}
	}
	if start == nil {
		return fmt.Errorf("cannot find concurrent start in routine %d to advance before anchor", who.GetRoutine())
	}
	delta := (anchor.GetTSort() - start.GetTSort()) - 1
	if delta <= 0 {
		return fmt.Errorf("non-positive delta while advancing")
	}
	tr.ShiftRoutine(who.GetRoutine(), start.GetTSort(), delta)
	return nil
}
