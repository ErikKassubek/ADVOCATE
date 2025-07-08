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
	"advocate/analysis/concurrent/cssts"
	"advocate/analysis/concurrent/pog"
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/timer"
)

// TODO: do we need the oSuc

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

	data.OSuc[id] = on

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

	suc := data.OSuc[id]

	data.CurrentVC[routine].Sync(suc.GetVC())
	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)

	pog.AddEdge(suc, on, false)
	cssts.AddEdge(suc, on, false)
}
