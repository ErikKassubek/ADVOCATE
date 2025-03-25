// Copyrigth (c) 2024 Erik Kassubek
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
	"errors"
	"fmt"
	"slices"
)

/*
 * Given a cyclic deadlock, rewrite the trace to make the bug occur. The trace is rewritten as follows:
 * We already get this (ordered) cycle from the analysis (the cycle is ordered in
 * such a way, that the edges inside a routine always go down). We now have to
 * reorder in such a way, that for edges from a to b, where a and b are in different
 * routines, b is run before a. We do this by shifting the timer of all b back,
 * until it is greater as a.
 *
 * For the example we therefor get the the following:
 * ~~~
 *   T1         T2          T3
 * lock(m)
 * unlock(m)
 * lock(m)
 *            lock(n)
 * lock(n)
 * unlock(m)
 * unlock(n)
 *                        lock(o)
 *            lock(o)     lock(m)
 *            unlock(o)   unlock(m)
 *            unlock(n)   unlock(o)
 * ~~~
 *
 * If this can lead to operations having the same time stamp. In this case,
 * we decide arbitrarily, which operation is executed first. (In practice
 * we set the same timestamp in the rewritten trace and the replay mechanism
 * will then select one of them arbitrarily).
 * If this is done for all edges, we remove all unlock operations, which
 * do not have a lock operation in the circle behind them in the same routine.
 * After that, we add the start and end marker before the first, and after the
 * last lock operation in the cycle.
 * Therefore the final rewritten trace will be
 * ~~~
 *   T1         T2          T3
 * start()
 * lock(m)
 * unlock(m)
 * lock(m)
 *            lock(n)
 * lock(n)
 *                        lock(o)
 *            lock(o)     lock(m)
 * end()
 */

func rewriteCyclicDeadlock_old(bug bugs.Bug) error {
	if len(bug.TraceElement2) == 0 {
		return errors.New("No trace elements in bug")
	}

	if len(bug.TraceElement2) < 2 {
		return errors.New("At least 2 trace elements are needed for a deadlock")
	}

	fmt.Println("Original trace:")
	analysis.PrintTrace([]string{}, true)

	lastTime := -1

	for _, e := range bug.TraceElement2 {
		if e.GetTSort() > lastTime {
			lastTime = e.GetTSort()
		}
	}

	fmt.Println("Last time:", lastTime)

	// remove tail after lastTime
	analysis.ShortenTrace(lastTime, true)

	routinesInCycle := make(map[int]struct{})

	maxIterations := 100 // prevent infinite loop
	for iter := 0; iter < maxIterations; iter++ {
		found := false
		// for all edges in the cycle shift the routine so that the next element is before the current element
		for i := 0; i < len(bug.TraceElement2); i++ {
			routinesInCycle[bug.TraceElement2[i].GetRoutine()] = struct{}{}

			j := (i + 1) % len(bug.TraceElement2)

			elem1 := bug.TraceElement2[i]
			elem2 := bug.TraceElement2[j]

			if elem1.GetRoutine() == elem2.GetRoutine() {
				continue
			}

			// shift the routine of elem1 so that elem 2 is before elem1
			res := analysis.ShiftRoutine(elem1.GetRoutine(), elem1.GetTPre(), elem2.GetTPre()-elem1.GetTPre()+2)
			if res {
				found = true
			}
		}

		if !found {
			fmt.Println("Needed", iter, "iterations")
			break
		}
	}

	// Remove trailing unlocks in relevant routines
	currentTrace := analysis.GetTraces()
	lastTime = -1

	for routine := range routinesInCycle {
		found := false
		for i := len(currentTrace[routine]) - 1; i >= 0; i-- {
			elem := currentTrace[routine][i]
			switch elem := elem.(type) {
			case *analysis.TraceElementMutex:
				if (*elem).IsLock() {
					analysis.ShortenRoutineIndex(routine, i, true)
					if lastTime == -1 || (*elem).GetTSort() > lastTime {
						lastTime = (*elem).GetTSort()
					}
					found = true
				}
			}
			if found {
				break
			}
		}
	}

	// Get the last lock in the cycle that will be sucessfull
	slices.SortFunc(bug.TraceElement2, func(a, b analysis.TraceElement) int {
		return a.GetTPre() - b.GetTPre()
	})

	lastSuccessfullLockTime := bug.TraceElement2[len(bug.TraceElement2)-2].GetTPre()

	fmt.Println("Prev last time:", lastSuccessfullLockTime)

	// add end signal at the point the replay should get stuck
	analysis.AddTraceElementReplay(lastSuccessfullLockTime+1, exitCodeCyclic, lastSuccessfullLockTime)

	analysis.PrintTrace([]string{}, false)

	return nil
}

func rewriteCyclicDeadlock(bug bugs.Bug) error {
	if len(bug.TraceElement2) == 0 {
		return errors.New("No trace elements in bug")
	}

	if len(bug.TraceElement2) < 2 {
		return errors.New("At least 2 trace elements are needed for a deadlock")
	}

	fmt.Println("Original trace:")
	analysis.PrintTrace([]string{}, true)

	lastTime := findLastTime(bug.TraceElement2)

	fmt.Println("Last time:", lastTime)

	// remove tail after lastTime and the last lock
	analysis.ShortenTrace(lastTime, true)
	for _, elem := range bug.TraceElement2 {
		analysis.ShortenRoutine(elem.GetRoutine(), elem.GetTSort()+1)
	}

	var locksetElements []analysis.TraceElement
	currentTrace := analysis.GetTraces()

	// Find the lockset elements and move their sections behind the unrelated parts of the trace
	for i, elem := range bug.TraceElement2 {
		// This is one is guranteed to be in the ls of elem
		prevElement := bug.TraceElement2[(i+len(bug.TraceElement2)-1)%len(bug.TraceElement2)]
		for j := len(currentTrace[elem.GetRoutine()]) - 1; j >= 0; j-- {
			locksetElement := currentTrace[elem.GetRoutine()][j]
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

	for _, elem := range slices.Backward(locksetElements) {
		analysis.ShiftRoutine(elem.GetRoutine(), elem.GetTPre(), lastTime-elem.GetTPre())
		lastTime = findLastTime(bug.TraceElement2)
	}

	// Find the largest time a lock needed to aquire so we can space them accordingly
	largestLockTime := 10
	for _, e := range bug.TraceElement2 {
		if e.GetTPost()-e.GetTPre() > largestLockTime {
			largestLockTime = e.GetTPost() - e.GetTPre()
		}
	}

	for i := 0; i < len(bug.TraceElement2); i++ {
		elem1 := bug.TraceElement2[i]

		analysis.ShiftRoutine(elem1.GetRoutine(), elem1.GetTPre(), lastTime-elem1.GetTPre()+largestLockTime*i)
	}

	lastSuccessfullLock := bug.TraceElement2[len(bug.TraceElement2)-2]
	analysis.AddTraceElementReplay(lastSuccessfullLock.GetTPre()+1, exitCodeCyclic, lastSuccessfullLock.GetTPre())

	analysis.PrintTrace([]string{}, true)

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
