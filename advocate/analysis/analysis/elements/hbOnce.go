// Copyright (c) 2024 Erik Kassubek
//
// File: vcOnce.go
// Brief: Update functions for happens before info for once operations
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package elements

import (
	"advocate/analysis/data"
	"advocate/analysis/hb/cssts"
	"advocate/analysis/hb/pog"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/timer"
)

// TODO: do we need the oSuc

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(on *trace.ElementOnce) {
	routine := on.GetRoutine()
	on.SetVc(vc.CurrentVC[routine])
	on.SetWVc(vc.CurrentVC[routine])

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

	vc.CurrentVC[routine].Inc(routine)
	vc.CurrentWVC[routine].Inc(routine)
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

	vc.CurrentVC[routine].Sync(suc.GetVC())
	vc.CurrentVC[routine].Inc(routine)
	vc.CurrentWVC[routine].Inc(routine)

	pog.AddEdge(suc, on, false)
	cssts.AddEdge(suc, on, false)
}
