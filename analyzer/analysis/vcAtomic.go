// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcAtomic.go
// Brief: Update for vector clocks from atomic operations
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

// vector clocks for last write times
var lw map[int]clock.VectorClock = make(map[int]clock.VectorClock)

/*
 * Create a new lw if needed
 * Args:
 *   index (int): The id of the atomic variable
 *   nRout (int): The number of routines in the trace
 */
func newLw(index int, nRout int) {
	if _, ok := lw[index]; !ok {
		lw[index] = clock.NewVectorClock(nRout)
	}
}

/*
 * Calculate the new vector clock for a write operation and update cv
 * Args:
 *   at (*TraceElementAtomic): The trace element
 *   vc (*map[int]VectorClock): The vector clocks
 */
func Write(at *TraceElementAtomic, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newLw(at.id, vc[at.id].GetSize())
	lw[at.id] = vc[at.routine].Copy()
	vc[at.routine] = vc[at.routine].Inc(at.routine)
}

/*
 * Calculate the new vector clock for a read operation and update cv
 * Args:
 *   at (*TraceElementAtomic): The trace element
 *   numberOfRoutines (int): The number of routines in the trace
 *   vc (map[int]VectorClock): The vector clocks
 *   sync bool: sync reader with last writer
 */
func Read(at *TraceElementAtomic, vc map[int]clock.VectorClock, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newLw(at.id, vc[at.id].GetSize())
	if sync {
		vc[at.routine] = vc[at.routine].Sync(lw[at.id])
	}
	vc[at.routine] = vc[at.routine].Inc(at.routine)
}

/*
 * Calculate the new vector clock for a swap operation and update cv. A swap
 * operation is a read and a write.
 * Args:
 *   at (*TraceElementAtomic): The trace element
 *   numberOfRoutines (int): The number of routines in the trace
 *   cv (map[int]VectorClock): The vector clocks
 *   sync bool: sync reader with last writer
 */
func Swap(at *TraceElementAtomic, cv map[int]clock.VectorClock, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	Read(at, cv, sync)
	Write(at, cv)
}
