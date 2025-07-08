// Copyright (c) 2024 Erik Kassubek
//
// File: vcOnce.go
// Brief: Update functions for happens before info for once operations
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/concurrent/clock"
	"advocate/analysis/data"
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
	if _, ok := data.OSuc[index]; !ok {
		data.OSuc[index] = clock.NewVectorClock(nRout)
	}
}

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(on *trace.ElementOnce) {
	routine := on.GetRoutine()
	on.SetVc(data.CurrentVC[routine])
	on.SetWVc(data.CurrentVC[routine])

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
func DoSuc(on *trace.ElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := on.GetID()
	routine := on.GetRoutine()

	newOSuc(id, data.CurrentVC[routine].GetSize())
	data.OSuc[id] = data.CurrentVC[routine].Copy()

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}

// DoFail updates and calculates the vector clocks given a unsuccessful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoFail(on *trace.ElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := on.GetID()
	routine := on.GetRoutine()

	newOSuc(id, data.CurrentVC[routine].GetSize())

	data.CurrentVC[routine].Sync(data.OSuc[id])
	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}
