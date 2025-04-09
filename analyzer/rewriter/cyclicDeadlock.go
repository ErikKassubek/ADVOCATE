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
	"analyzer/analysis"
	"analyzer/bugs"
	"analyzer/clock"
	"errors"
	"fmt"
)

func rewriteCyclicDeadlock(trace *analysis.Trace, bug bugs.Bug) error {
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
	analysis.ShortenTrace(lastTime, true)
	for _, elem := range bug.TraceElement2 {
		analysis.ShortenRoutine(elem.GetRoutine(), elem.GetTSort())
	}

	var locksetElements []analysis.TraceElement

	// Find the lockset elements
	for i, elem := range bug.TraceElement2 {
		// This is one is guranteed to be in the lockset of elem
		prevElement := bug.TraceElement2[(i+len(bug.TraceElement2)-1)%len(bug.TraceElement2)]
		for j := len(analysis.GetRoutineTrace(elem.GetRoutine())) - 1; j >= 0; j-- {
			locksetElement := analysis.GetRoutineTrace(elem.GetRoutine())[j]
			if locksetElement.GetID() != prevElement.GetID() {
				continue
			}
			if !locksetElement.(*analysis.TraceElementMutex).IsLock() {
				continue
			}
			locksetElements = append(locksetElements, locksetElement)
			break
		}
	}

	// If there are any unlocks in the remaining traces, try to ensure that those can happen before we run into the deadlock!
	for _, relevantRoutineElem := range bug.TraceElement2 {
		routine := relevantRoutineElem.GetRoutine()                // Iterate through all relevant routines
		for _, unlock := range analysis.GetRoutineTrace(routine) { // Iterate through all remaining elements in the routine
			switch unlock := unlock.(type) {
			case *analysis.TraceElementMutex:
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
							var concurrentStartElem analysis.TraceElement = nil
							for _, possibleStart := range analysis.GetRoutineTrace(lockElem.GetRoutine()) {
								if clock.GetHappensBefore(possibleStart.GetVCWmHB(), (*unlock).GetVCWmHB()) == clock.Concurrent {
									// fmt.Println("Concurrent to", possibleStart.GetTID(), possibleStart.GetTPre(), possibleStart.GetTPost(), possibleStart.GetRoutine(), possibleStart.GetID())
									concurrentStartElem = possibleStart
									break
								}
							}

							if concurrentStartElem == nil {
								fmt.Println("Could not find concurrent element for Routine", lockElem.GetRoutine(), "so we cannot move it behind unlock", unlock.GetID(), "in Routine", unlock.GetRoutine())
								break
							}

							routineEndElem := analysis.GetRoutineTrace(lockElem.GetRoutine())[len(analysis.GetRoutineTrace(lockElem.GetRoutine()))-1]
							analysis.ShiftRoutine(lockElem.GetRoutine(), concurrentStartElem.GetTPre(), ((*unlock).GetTSort()-concurrentStartElem.GetTSort())+1)
							if routineEndElem.GetTPost() > lastTime {
								lastTime = routineEndElem.GetTPost()
							}
							analysis.ShiftConcurrentOrAfterToAfter(unlock)
						}
					}
				}
			}
		}
	}

	analysis.AddTraceElementReplay(lastTime+1, exitCodeCyclic)

	// fmt.Println("Rewritten Trace:")
	// analysis.PrintTrace()

	for _, elem := range bug.TraceElement2 {
		fmt.Println("Deadlocking Element: ", elem.GetRoutine(), "M", elem.GetTPre(), elem.GetTPost(), elem.GetID())
	}

	return nil
}

func findLastTime(bugElements []analysis.TraceElement) int {
	lastTime := -1

	for _, e := range bugElements {
		if lastTime == -1 || e.GetTSort() > lastTime {
			lastTime = e.GetTSort()
		}
	}
	return lastTime
}
