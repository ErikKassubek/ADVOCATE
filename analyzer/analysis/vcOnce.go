// Copyright (c) 2024 Erik Kassubek
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

// Create a new oSuc if needed
//
// Parameter:
//   - index int: The id of the atomic variable
//   - nRout int: The number of routines in the trace
func newOSuc(index int, nRout int) {
	if _, ok := oSuc[index]; !ok {
		oSuc[index] = clock.NewVectorClock(nRout)
	}
}

// DoSuc updates and calculates the vector clocks given a successful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoSuc(on *TraceElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newOSuc(on.id, currentVC[on.routine].GetSize())
	oSuc[on.id] = currentVC[on.routine].Copy()

	currentVC[on.routine].Inc(on.routine)
	currentWVC[on.routine].Inc(on.routine)
}

// DoFail updates and calculates the vector clocks given a unsuccessful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoFail(on *TraceElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newOSuc(on.id, currentVC[on.routine].GetSize())

	currentVC[on.routine].Sync(oSuc[on.id])
	currentVC[on.routine].Inc(on.routine)
	currentWVC[on.routine].Inc(on.routine)
}
