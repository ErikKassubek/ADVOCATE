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
	"advocate/analysis/data"
	"advocate/trace"
)

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	switch co.GetOpC() {
	case trace.WaitCondOp:
		// wait does not add any edge
	case trace.SignalOp:
		CondSignal(co)
	case trace.BroadcastOp:
		CondBroadcast(co)
	}
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondSignal(co *trace.ElementCond) {
	id := co.GetID()

	if len(data.CurrentlyWaiting[id]) != 0 {
		tWait := data.CurrentlyWaiting[id][0]
		AddEdge(co, tWait, false)
	}
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondBroadcast(co *trace.ElementCond) {
	id := co.GetID()
	for _, wait := range data.CurrentlyWaiting[id] {
		AddEdge(co, wait, false)
	}
}
