// Copyright (c) 2024 Erik Kassubek
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
	"analyzer/timer"
)

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
 */
func Change(wa *TraceElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newWg(wa.id, currentVC[wa.id].GetSize())
	lastChangeWG[wa.id].Sync(currentVC[wa.routine])

	currentVC[wa.routine].Inc(wa.routine)
	currentWVC[wa.routine].Inc(wa.routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["doneBeforeAdd"] {
		checkForDoneBeforeAddChange(wa)
	}
}

/*
 * Calculate the new vector clock for a wait operation and update cv
 * Args:
 *   wa (*TraceElementWait): The trace element
 */
func Wait(wa *TraceElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newWg(wa.id, currentVC[wa.id].GetSize())

	if wa.tPost != 0 {
		currentVC[wa.routine].Sync(lastChangeWG[wa.id])
	}

	currentVC[wa.routine].Inc(wa.routine)
	currentWVC[wa.routine].Inc(wa.routine)
}
