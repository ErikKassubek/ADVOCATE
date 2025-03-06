// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcOnce.go
// Brief: Update functions of vector clocks for once operations
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

// TODO: do we need the oSuc

/*
 * Create a new oSuc if needed
 * Args:
 *   index (int): The id of the atomic variable
 *   nRout (int): The number of routines in the trace
 */
func newOSuc(index int, nRout int) {
	if _, ok := oSuc[index]; !ok {
		oSuc[index] = clock.NewVectorClock(nRout)
	}
}

/*
 * Update and calculate the vector clocks given a successful do operation
 * Args:
 *   on (*TraceElementOnce): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 */
func DoSuc(on *TraceElementOnce, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newOSuc(on.id, vc[on.id].GetSize())
	oSuc[on.id] = vc[on.routine].Copy()
	vc[on.routine] = vc[on.routine].Inc(on.routine)
}

/*
 * Update and calculate the vector clocks given a unsuccessful do operation
 * Args:
 *   on (*TraceElementOnce): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 */
func DoFail(on *TraceElementOnce, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newOSuc(on.id, vc[on.id].GetSize())
	vc[on.routine] = vc[on.routine].Sync(oSuc[on.id])
	vc[on.routine] = vc[on.routine].Inc(on.routine)
}
