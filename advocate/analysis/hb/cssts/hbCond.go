// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the cssts for conds
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package cssts

import (
	"advocate/analysis/baseA"
	"advocate/trace"
)

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	switch co.GetType(true) {
	case trace.CondWait:
		// wait does not add any edge
	case trace.CondSignal:
		CondSignal(co)
	case trace.CondBroadcast:
		CondBroadcast(co)
	}
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondSignal(co *trace.ElementCond) {
	id := co.GetObjId()

	if len(baseA.CurrentlyWaiting[id]) != 0 {
		tWait := baseA.CurrentlyWaiting[id][0]
		AddEdge(co, tWait, false)
	}
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondBroadcast(co *trace.ElementCond) {
	id := co.GetObjId()
	for _, wait := range baseA.CurrentlyWaiting[id] {
		AddEdge(co, wait, false)
	}
}
