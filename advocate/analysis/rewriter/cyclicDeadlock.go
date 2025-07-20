// Copyright (c) 2024 Erik Kassubek
//
// File: cyclicDeadlock.go
// Brief: Rewrite trace for cyclic deadlocks
//
// Author: Erik Kassubek
// Created: 2024-04-05
//
// License: BSD-3-Clause

package rewriter

import (
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/bugs"
	"advocate/trace"
	"advocate/utils/helper"
	"errors"
	"fmt"
)

// rewriteCyclicDeadlock rewrites the trace in such a way, that it should
// trigger the cyclic/resource deadlock described in the bug
//
// Parameter:
//   - trace *trace.Trace: the trace to rewrite
//   - bug bugs.Bug: the bug that should be triggered by the rewrite
func rewriteCyclicDeadlock(tr *trace.Trace, bug bugs.Bug) error {
	if len(bug.TraceElement2) == 0 {
		return errors.New("no trace elements in bug")
	}

	if len(bug.TraceElement2) < 2 {
		return errors.New("at least 2 trace elements are needed for a deadlock")
	}

	// fmt.Println("Original trace:")
	// analysis.PrintTrace()

	lastTime := findLastTime(bug.TraceElement2)

	// fmt.Println("Last time:", lastTime)

	// remove tail after lastTime and the last lock
	tr.ShortenTrace(lastTime, true)
	for _, elem := range bug.TraceElement2 {
		tr.ShortenRoutine(elem.GetRoutine(), elem.GetTSort())
	}

	var locksetElements []trace.Element

	// Find the lockset elements
	for i, elem := range bug.TraceElement2 {
		// This is one is guranteed to be in the lockset of elem
		prevElement := bug.TraceElement2[(i+len(bug.TraceElement2)-1)%len(bug.TraceElement2)]
		for j := len(tr.GetRoutineTrace(elem.GetRoutine())) - 1; j >= 0; j-- {
			locksetElement := tr.GetRoutineTrace(elem.GetRoutine())[j]
			if locksetElement.GetID() != prevElement.GetID() {
				continue
			}
			if !locksetElement.(*trace.ElementMutex).IsLock() {
				continue
			}
			locksetElements = append(locksetElements, locksetElement)
			break
		}
	}

	// If there are any unlocks in the remaining traces, try to ensure that those can happen before we run into the deadlock!
	for _, relevantRoutineElem := range bug.TraceElement2 {
		routine := relevantRoutineElem.GetRoutine()          // Iterate through all relevant routines
		for _, unlock := range tr.GetRoutineTrace(routine) { // Iterate through all remaining elements in the routine
			switch unlock := unlock.(type) {
			case *trace.ElementMutex:
				if !(*unlock).IsLock() { // Find Unlock elements
					// Check if the unlocked mutex is in the locksets of the deadlock cycle
					for _, lockElem := range locksetElements {
						// If yes, make sure the unlock happens before the final lock attempts!
						if (*unlock).GetID() == lockElem.GetID() {
							// Do nothing if the unlock already happens before the lockset element
							if (*unlock).GetTPre() < lockElem.GetTPre() {
								break
							}

							// Move the as much of the routine of the deadlocking element as possible behind this unlock!
							var concurrentStartElem trace.Element = nil
							for _, possibleStart := range tr.GetRoutineTrace(lockElem.GetRoutine()) {
								if clock.GetHappensBefore(possibleStart.GetWVc(), (*unlock).GetWVc()) == hb.Concurrent {
									// fmt.Println("Concurrent to", possibleStart.GetTID(), possibleStart.GetTPre(), possibleStart.GetTPost(), possibleStart.GetRoutine(), possibleStart.GetID())
									concurrentStartElem = possibleStart
									break
								}
							}

							if concurrentStartElem == nil {
								fmt.Println("Could not find concurrent element for Routine", lockElem.GetRoutine(), "so we cannot move it behind unlock", unlock.GetID(), "in Routine", unlock.GetRoutine())
								break
							}

							routineEndElem := tr.GetRoutineTrace(lockElem.GetRoutine())[len(tr.GetRoutineTrace(lockElem.GetRoutine()))-1]
							tr.ShiftRoutine(lockElem.GetRoutine(), concurrentStartElem.GetTPre(), ((*unlock).GetTSort()-concurrentStartElem.GetTSort())+1)
							if routineEndElem.GetTPost() > lastTime {
								lastTime = routineEndElem.GetTPost()
							}
							tr.ShiftConcurrentOrAfterToAfter(unlock)
						}
					}
				}
			}
		}
	}

	tr.AddTraceElementReplay(lastTime+1, helper.ExitCodeCyclic)

	// fmt.Println("Rewritten Trace:")
	// analysis.PrintTrace()

	for _, elem := range bug.TraceElement2 {
		fmt.Println("Deadlocking Element: ", elem.GetRoutine(), "M", elem.GetTPre(), elem.GetTPost(), elem.GetID())
	}

	return nil
}

// findLastTime returns the latest time stamp from the bug elements
//
// Parameters:
//   - bugElements []trace.TraceElement: the bug element to search through
//
// Returns:
//   - int: the highest tPost from the bug elements
func findLastTime(bugElements []trace.Element) int {
	lastTime := -1

	for _, e := range bugElements {
		if lastTime == -1 || e.GetTSort() > lastTime {
			lastTime = e.GetTSort()
		}
	}
	return lastTime
}
