// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for once
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package vc

import (
	"advocate/analysis/baseA"
	"advocate/trace"
)

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(on *trace.ElementOnce) {
	routine := on.GetRoutine()

	on.SetVc(CurrentVC[routine])
	on.SetWVc(CurrentVC[routine])

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
	routine := on.GetRoutine()

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// DoFail updates and calculates the vector clocks given a unsuccessful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoFail(on *trace.ElementOnce) {
	id := on.GetID()
	routine := on.GetRoutine()

	suc := baseA.OSuc[id]

	if suc != nil {
		CurrentVC[routine].Sync(suc.GetVC())
	}
	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}
