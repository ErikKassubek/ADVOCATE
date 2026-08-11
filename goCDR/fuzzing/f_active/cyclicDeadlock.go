// Copyright (c) 2024 Erik Kassubek
//
// File: cyclicDeadlock.go
// Brief: Rewrite trace for cyclic deadlocks
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_active

import (
	"errors"
	"fmt"
	"gocdr/analysis/a_hb"
	"gocdr/analysis/hb/a_clock"
	"gocdr/trace"
	"gocdr/utils/helper"
	"gocdr/utils/log"
	"gocdr/utils/results/bugs"
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

	lastTime := findLastTime(bug.TraceElement2)

	// remove tail after lastTime and the last lock
	tr.ShortenTrace(lastTime, true)
	for _, elem := range bug.TraceElement2 {
		tr.ShortenRoutine(elem.Routine(), elem.T(trace.Sorting))
	}

	var locksetElements []trace.Element

	// Find the lockset elements
	for i, elem := range bug.TraceElement2 {
		// This is one is guranteed to be in the lockset of elem
		prevElement := bug.TraceElement2[(i+len(bug.TraceElement2)-1)%len(bug.TraceElement2)]
		for j := tr.GetRoutineTrace(elem.Routine()).Len() - 1; j >= 0; j-- {
			locksetElement := tr.GetRoutineTrace(elem.Routine()).At(j)
			if locksetElement.ObjID() != prevElement.ObjID() {
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
		routine := relevantRoutineElem.Routine()                     // Iterate through all relevant routines
		for _, unlock := range tr.GetRoutineTrace(routine).Elems() { // Iterate through all remaining elements in the routine
			switch unlock := unlock.(type) {
			case *trace.ElementMutex:
				if !(*unlock).IsLock() { // Find Unlock elements
					// Check if the unlocked mutex is in the locksets of the deadlock cycle
					for _, lockElem := range locksetElements {
						// If yes, make sure the unlock happens before the final lock attempts!
						if (*unlock).ObjID() == lockElem.ObjID() {
							// Do nothing if the unlock already happens before the lockset element
							if (*unlock).T(trace.Request) < lockElem.T(trace.Request) {
								break
							}

							// Move the as much of the routine of the deadlocking element as possible behind this unlock!
							var concurrentStartElem trace.Element = nil
							for _, possibleStart := range tr.GetRoutineTrace(lockElem.Routine()).Elems() {
								if a_clock.GetHappensBefore(possibleStart.GetVC(a_clock.Weak), (*unlock).GetVC(a_clock.Weak)) == a_hb.Concurrent {
									concurrentStartElem = possibleStart
									break
								}
							}

							if concurrentStartElem == nil {
								log.Info("Could not find concurrent element for Routine", lockElem.Routine(), "so we cannot move it behind unlock", unlock.ObjID(), "in Routine", unlock.Routine())
								break
							}

							routineEndElem := tr.GetRoutineTrace(lockElem.Routine()).Last()
							tr.ShiftRoutine(lockElem.Routine(), concurrentStartElem.T(trace.Request), ((*unlock).T(trace.Sorting)-concurrentStartElem.T(trace.Sorting))+1)
							if routineEndElem.T(trace.Commit) > lastTime {
								lastTime = routineEndElem.T(trace.Commit)
							}
							tr.ShiftConcurrentOrAfterToAfter(unlock)
						}
					}
				}
			}
		}
	}

	tr.AddTraceElementReplay(lastTime+1, helper.ExitCodeCyclic)

	for _, elem := range bug.TraceElement2 {
		fmt.Println("Deadlocking Element: ", elem.Routine(), "M", elem.T(trace.Request), elem.T(trace.Commit), elem.ObjID())
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
		if lastTime == -1 || e.T(trace.Sorting) > lastTime {
			lastTime = e.T(trace.Sorting)
		}
	}
	return lastTime
}
