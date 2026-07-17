// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for conds
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package a_vc

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
)

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	routine := co.Routine()
	co.Vc(a_clock.Strong, CurrentVC[routine])
	co.Vc(a_clock.Weak, CurrentWVC[routine])

	switch co.Type(true) {
	case trace.CondWait:
		CondWait(co)
	case trace.CondSignal:
		CondSignal(co)
	case trace.CondBroadcast:
		CondBroadcast(co)
	}
}

// CondWait updates and calculates the vector clocks given a wait operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondWait(co *trace.ElementCond) {
	routine := co.Routine()

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondSignal(co *trace.ElementCond) {
	id := co.ObjID()
	routine := co.Routine()

	if len(a_base.CurrentlyWaiting[id]) != 0 {
		tWait := a_base.CurrentlyWaiting[id][0]
		CurrentVC[tWait.Routine()].Sync(CurrentVC[routine])
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondBroadcast(co *trace.ElementCond) {
	id := co.ObjID()
	routine := co.Routine()

	for _, wait := range a_base.CurrentlyWaiting[id] {
		CurrentVC[wait.Routine()].Sync(CurrentVC[routine])
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}
