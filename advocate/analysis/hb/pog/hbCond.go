// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the cssts for conds
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/baseA"
	"advocate/trace"
)

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(graph *PoGraph, co *trace.ElementCond) {
	switch co.GetType(true) {
	case trace.CondWait:
		// wait does not add any edge
	case trace.CondSignal:
		CondSignal(graph, co)
	case trace.CondBroadcast:
		CondBroadcast(graph, co)
	}
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - co *TraceElementCond: The trace element
func CondSignal(graph *PoGraph, co *trace.ElementCond) {
	id := co.GetObjId()

	if len(baseA.CurrentlyWaiting[id]) != 0 {
		tWait := baseA.CurrentlyWaiting[id][0]
		if graph != nil {
			graph.AddEdge(co, tWait)
		} else {
			AddEdge(co, tWait, false)
		}
	}
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - co *TraceElementCond: The trace element
func CondBroadcast(graph *PoGraph, co *trace.ElementCond) {
	id := co.GetObjId()
	for _, wait := range baseA.CurrentlyWaiting[id] {
		if graph != nil {
			graph.AddEdge(co, wait)
		} else {
			AddEdge(co, wait, false)
		}
	}
}
