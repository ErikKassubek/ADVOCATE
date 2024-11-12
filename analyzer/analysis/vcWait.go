// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcWait.go
// Brief: Update functions of vector groups for wait group operations
//        Some function start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	timemeasurement "analyzer/timeMeasurement"
)

// vector clock for each wait group
var lastChangeWG map[int]clock.VectorClock = make(map[int]clock.VectorClock)

/*
 * Create a new wg if needed
 * Args:
 *   index (int): The id of the wait group
 *   nRout (int): The number of routines in the trace
 */
func newWg(index int, nRout int) {
	if _, ok := lastChangeWG[index]; !ok {
		lastChangeWG[index] = clock.NewVectorClock(nRout)
	}
}

/*
 * Calculate the new vector clock for a add or done operation and update cv
 * Args:
 *   wa (*TraceElementWait): The trace element
 *   vc (map[int]VectorClock): The vector clocks
 */
func Change(wa *TraceElementWait, vc map[int]clock.VectorClock) {
	newWg(wa.id, vc[wa.id].GetSize())
	lastChangeWG[wa.id] = lastChangeWG[wa.id].Sync(vc[wa.routine])
	vc[wa.routine] = vc[wa.routine].Inc(wa.routine)

	if analysisCases["doneBeforeAdd"] {
		timemeasurement.Start("panic")
		checkForDoneBeforeAddChange(wa)
		timemeasurement.End("panic")
	}
}

/*
 * Calculate the new vector clock for a wait operation and update cv
 * Args:
 *   wa (*TraceElementWait): The trace element
 *   vc (*map[int]VectorClock): The vector clocks
 */
func Wait(wa *TraceElementWait, vc map[int]clock.VectorClock) {
	newWg(wa.id, vc[wa.id].GetSize())
	if wa.tPost != 0 {
		vc[wa.routine] = vc[wa.routine].Sync(lastChangeWG[wa.id])
		vc[wa.routine] = vc[wa.routine].Inc(wa.routine)
	}
}
