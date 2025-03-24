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

func rewriteCyclicDeadlock(bug bugs.Bug) error {
	firstTime := -1
	lastTime := -1

	if len(bug.TraceElement2) == 0 {
		return errors.New("No trace elements in bug")
	}

	for _, elem := range bug.TraceElement2 {
		// get the first and last mutex operation in the cycle
		time := elem.GetTPre()
		if firstTime == -1 || time < firstTime {
			firstTime = time
		}
		if lastTime == -1 || time > lastTime {
			lastTime = time
		}
	}

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
			res := analysis.ShiftRoutine(elem1.GetRoutine(), elem1.GetTPre(), elem2.GetTPre()-elem1.GetTPre()+1)

			if res {
				found = true
			}
		}

		if !found {
			break
		}
	}

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

	// add start and end signal
	analysis.AddTraceElementReplay(lastTime+1, exitCodeCyclic)

	return nil
}
