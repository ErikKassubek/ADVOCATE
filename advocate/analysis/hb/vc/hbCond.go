// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for conds
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

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	routine := co.GetRoutine()
	co.SetVc(CurrentVC[routine])
	co.SetWVc(CurrentWVC[routine])

	switch co.GetType(true) {
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
	routine := co.GetRoutine()

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondSignal(co *trace.ElementCond) {
	id := co.GetID()
	routine := co.GetRoutine()

	if len(baseA.CurrentlyWaiting[id]) != 0 {
		tWait := baseA.CurrentlyWaiting[id][0]
		CurrentVC[tWait.GetRoutine()].Sync(CurrentVC[routine])
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondBroadcast(co *trace.ElementCond) {
	id := co.GetID()
	routine := co.GetRoutine()

	for _, wait := range baseA.CurrentlyWaiting[id] {
		CurrentVC[wait.GetRoutine()].Sync(CurrentVC[routine])
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}
