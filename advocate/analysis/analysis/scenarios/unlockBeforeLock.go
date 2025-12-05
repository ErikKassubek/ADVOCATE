// Copyright (c) 2024 Erik Kassubek
//
// File: analysisUnlockBeforeLock.go
// Brief: Analysis for unlock of not locked mutex
//
// Author: Erik Kassubek
// Created: 2024-09-23
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
	"advocate/utils/types"
)

// CheckForUnlockBeforeLockLock collects all locks for the analysis
//
// Parameter:
//   - mu *TraceElementMutex: the trace mutex element
func CheckForUnlockBeforeLockLock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	id := mu.GetObjId()

	if _, ok := baseA.AllLocks[id]; !ok {
		baseA.AllLocks[id] = make([]trace.Element, 0)
	}

	baseA.AllLocks[id] = append(baseA.AllLocks[id], mu)
}

// CheckForUnlockBeforeLockUnlock collects all unlocks for the analysis
//
// Parameter:
//   - mu *TraceElementMutex: the trace mutex element
func CheckForUnlockBeforeLockUnlock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	id := mu.GetObjId()

	if _, ok := baseA.AllLocks[id]; !ok {
		baseA.AllUnlocks[id] = make([]trace.Element, 0)
	}

	baseA.AllUnlocks[id] = append(baseA.AllUnlocks[id], mu)
}

// CheckForUnlockBeforeLock check if we can get a unlock of a not locked mutex
// For each done operation, build a bipartite st graph.
// Use the Ford-Fulkerson algorithm to find the maximum flow.
// If the maximum flow is smaller than the number of unlock operations, a unlock before lock is possible.
func CheckForUnlockBeforeLock() {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	for id := range baseA.AllUnlocks { // for all mutex ids
		// if a lock and the corresponding unlock is always in the same routine, this cannot happen
		if trace.SameRoutine(baseA.AllLocks[id], baseA.AllUnlocks[id]) {
			continue
		}

		graph := buildResidualGraph(baseA.AllLocks[id], baseA.AllUnlocks[id])

		maxFlow, graph, err := calculateMaxFlow(graph)
		if err != nil {
			log.Error("Could not check for unlock before lock: ", err)
		}

		nrUnlock := len(baseA.AllUnlocks)

		locks := make([]trace.Element, 0)
		unlocks := make([]trace.Element, 0)

		if maxFlow < nrUnlock {
			for _, l := range baseA.AllLocks[id] {
				if !types.Contains(graph[&drain], l) {
					locks = append(locks, l)
				}
			}

			for _, u := range graph[&source] {
				unlocks = append(unlocks, u)
			}

			locksSorted := make([]trace.Element, 0)
			unlockSorted := make([]trace.Element, 0)

			for i := 0; i < len(locks); {
				removed := false
				for j := 0; j < len(unlocks); {
					if clock.GetHappensBefore(locks[i].GetVC(), unlocks[j].GetVC()) == hb.Concurrent {
						locksSorted = append(locksSorted, locks[i])
						unlockSorted = append(unlockSorted, unlocks[i])
						locks = append(locks[:i], locks[i+1:]...)
						unlocks = append(unlocks[:j], unlocks[j+1:]...)
						removed = true
						break
					} else {
						j++
					}
				}
				if !removed {
					i++
				}
			}

			args1 := []results.ResultElem{} // unlocks
			args2 := []results.ResultElem{} // locks

			for _, u := range unlockSorted {
				if u.GetTID() == "\n" {
					continue
				}
				file, line, tPre, err := trace.InfoFromTID(u.GetTID())
				if err != nil {
					log.Error(err.Error())
					continue
				}

				args1 = append(args1, results.TraceElementResult{
					RoutineID: u.GetRoutine(),
					ObjID:     id,
					TPre:      tPre,
					ObjType:   u.GetType(true),
					File:      file,
					Line:      line,
				})
			}

			for _, l := range locksSorted {
				if l.GetTID() == "\n" {
					continue
				}
				file, line, tPre, err := trace.InfoFromTID(l.GetTID())
				if err != nil {
					log.Error(err.Error())
					continue
				}

				args2 = append(args2, results.TraceElementResult{
					RoutineID: l.GetRoutine(),
					ObjID:     id,
					TPre:      tPre,
					ObjType:   l.GetType(true),
					File:      file,
					Line:      line,
				})
			}

			results.Result(results.CRITICAL, helper.PUnlockBeforeLock, "unlock",
				args1, "lock", args2)
		}
	}
}
