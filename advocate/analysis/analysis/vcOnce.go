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
	"advocate/analysis/clock"
	"advocate/trace"
	"advocate/utils/timer"
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

// UpdateVCOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateVCOnce(on *trace.TraceElementOnce) {
	routine := on.GetRoutine()
	on.SetVc(currentVC[routine])
	on.SetWVc(currentVC[routine])

	if on.GetSuc() {
		DoSuc(on)
	} else {
		DoFail(on)
	}
}

// DoSuc updates and calculates the vector clocks given a successful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoSuc(on *trace.TraceElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := on.GetID()
	routine := on.GetRoutine()

	newOSuc(id, currentVC[routine].GetSize())
	oSuc[id] = currentVC[routine].Copy()

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}

// DoFail updates and calculates the vector clocks given a unsuccessful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoFail(on *trace.TraceElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := on.GetID()
	routine := on.GetRoutine()

	newOSuc(id, currentVC[routine].GetSize())

	currentVC[routine].Sync(oSuc[id])
	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}
