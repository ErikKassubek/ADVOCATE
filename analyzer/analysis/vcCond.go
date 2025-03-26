// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcCond.go
// Brief: Update functions for vector clocks from conditional variables operations
//
// Author: Erik Kassubek
// Created: 2024-01-09
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/timer"
)

/*
 * Update and calculate the vector clocks given a wait operation
 * Args:
 *   co (*TraceElementCond): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 */
func CondWait(co *TraceElementCond, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if co.tPost != 0 { // not leak
		if _, ok := currentlyWaiting[co.id]; !ok {
			currentlyWaiting[co.id] = make([]int, 0)
		}
		currentlyWaiting[co.id] = append(currentlyWaiting[co.id], co.routine)
	}
	vc[co.routine].Inc(co.routine)
}

/*
 * Update and calculate the vector clocks given a signal operation
 * Args:
 *   co (*TraceElementCond): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 */
func CondSignal(co *TraceElementCond, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if len(currentlyWaiting[co.id]) != 0 {
		tWait := currentlyWaiting[co.id][0]
		currentlyWaiting[co.id] = currentlyWaiting[co.id][1:]
		vc[tWait] = vc[tWait].Sync(vc[co.routine])
	}
	vc[co.routine].Inc(co.routine)
}

/*
 * Update and calculate the vector clocks given a broadcast operation
 * Args:
 *   co (*TraceElementCond): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 */
func CondBroadcast(co *TraceElementCond, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	for _, wait := range currentlyWaiting[co.id] {
		vc[wait] = vc[wait].Sync(vc[co.routine])
	}
	currentlyWaiting[co.id] = make([]int, 0)

	vc[co.routine].Inc(co.routine)
}
