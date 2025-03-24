// Copyright (c) 2024 Erik Kassubek
//
// File: analysisUnlockBeforeLock.go
// Brief: Analysis for unlock of not locked mutex
//
// Author: Erik Kassubek
// Created: 2024-09-23
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/results"
	"analyzer/timer"
	"analyzer/utils"
	"errors"
	"fmt"
)

/*
 * Collect all locks for the analysis
 * Args:
 *    mu *TraceElementMutex: the trace mutex element
 */
func checkForUnlockBeforeLockLock(mu *TraceElementMutex) {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	if _, ok := allLocks[mu.id]; !ok {
		allLocks[mu.id] = make([]TraceElement, 0)
	}

	allLocks[mu.id] = append(allLocks[mu.id], mu)
}

/*
 * Collect all unlocks for the analysis
 * Args:
 *    mu *TraceElementMutex: the trace mutex element
 */
func checkForUnlockBeforeLockUnlock(mu *TraceElementMutex) {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	if _, ok := allLocks[mu.id]; !ok {
		allUnlocks[mu.id] = make([]TraceElement, 0)
	}

	allUnlocks[mu.id] = append(allUnlocks[mu.id], mu)
}

/*
 * Check if we can get a unlock of a not locked mutex
 * For each done operation, build a bipartite st graph.
 * Use the Ford-Fulkerson algorithm to find the maximum flow.
 * If the maximum flow is smaller than the number of unlock operations, a unlock before lock is possible.
 */
func checkForUnlockBeforeLock() {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	for id := range allUnlocks { // for all mutex ids
		// if a lock and the corresponding unlock is always in the same routine, this cannot happen
		if sameRoutine(allLocks[id], allUnlocks[id]) {
			continue
		}

		graph := buildResidualGraph(allLocks[id], allUnlocks[id])

		maxFlow, graph, err := calculateMaxFlow(graph)
		if err != nil {
			fmt.Println("Could not check for unlock before lock: ", err)
		}

		nrUnlock := len(allUnlocks)

		locks := make([]TraceElement, 0)
		unlocks := make([]TraceElement, 0)

		if maxFlow < nrUnlock {
			for _, l := range allLocks[id] {
				if !utils.ContainsString(graph["t"], l.GetTID()) {
					locks = append(locks, l)
				}
			}

			for _, u := range graph["s"] {
				unlockTId, err := getUnlockElemFromTID(id, u)
				if err != nil {
					utils.LogError(err.Error())
				} else {
					unlocks = append(unlocks, unlockTId)
				}
			}

			locksSorted := make([]TraceElement, 0)
			unlockSorted := make([]TraceElement, 0)

			for i := 0; i < len(locks); {
				removed := false
				for j := 0; j < len(unlocks); {
					if clock.GetHappensBefore(locks[i].GetVC(), unlocks[j].GetVC()) == clock.Concurrent {
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
				file, line, tPre, err := infoFromTID(u.GetTID())
				if err != nil {
					utils.LogError(err.Error())
					continue
				}

				args1 = append(args1, results.TraceElementResult{
					RoutineID: u.GetRoutine(),
					ObjID:     id,
					TPre:      tPre,
					ObjType:   u.GetObjType(true),
					File:      file,
					Line:      line,
				})
			}

			for _, l := range locksSorted {
				if l.GetTID() == "\n" {
					continue
				}
				file, line, tPre, err := infoFromTID(l.GetTID())
				if err != nil {
					utils.LogError(err.Error())
					continue
				}

				args2 = append(args2, results.TraceElementResult{
					RoutineID: l.GetRoutine(),
					ObjID:     id,
					TPre:      tPre,
					ObjType:   l.GetObjType(true),
					File:      file,
					Line:      line,
				})
			}

			results.Result(results.CRITICAL, results.PUnlockBeforeLock, "unlock",
				args1, "lock", args2)
		}
	}
}

func getUnlockElemFromTID(id int, tID string) (TraceElement, error) {
	for _, u := range allUnlocks[id] {
		if u.GetTID() == tID {
			return u, nil
		}
	}
	return nil, errors.New("Could not find unlock operation with tID " + tID)
}
