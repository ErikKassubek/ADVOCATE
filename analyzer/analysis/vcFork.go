// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcFork.go
// Brief: Update function for vector clocks from forks (creation of new routine)
//
// Author: Erik Kassubek
// Created: 2023-07-26
//
// License: BSD-3-Clause

package analysis

import "analyzer/clock"

/*
 * Update the vector clocks given a fork operation
 * Args:
 *   oldRout (int): The id of the old routine
 *   newRout (int): The id of the new routine
 *   vcHb (map[int]VectorClock): The current hb vector clocks
 *   vcMhb (map[int]VectorClock): The current mhb vector clocks
 */
func Fork(fo *TraceElementFork, vcHb map[int]clock.VectorClock, vcMhb map[int]clock.VectorClock) {
	oldRout := fo.routine
	newRout := fo.id

	vcHb[newRout] = vcHb[oldRout].Copy()
	vcHb[oldRout] = vcHb[oldRout].Inc(oldRout)
	vcHb[newRout] = vcHb[newRout].Inc(newRout)

	vcMhb[newRout] = vcMhb[oldRout].Copy()
	vcMhb[oldRout] = vcMhb[oldRout].Inc(oldRout)
	vcMhb[newRout] = vcMhb[newRout].Inc(newRout)

	allForks[fo.id] = fo
}
