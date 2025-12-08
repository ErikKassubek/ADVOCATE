// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the pog for wait group
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/baseA"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateHBWait update the pog for a wait group operation
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - wa *trace.TraceElementWait: the wait group operation
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func UpdateHBWait(graph *PoGraph, wa *trace.ElementWait, recorded bool) {
	switch wa.GetOpW() {
	case trace.WaitAdd, trace.WaitDone:
		Change(graph, wa)
	case trace.WaitWait:
		Wait(graph, wa, recorded)
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		log.Error(err)
	}
}

// Change updates the pog for an add or done operation and update cv
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - wa *TraceElementWait: The trace element
func Change(graph *PoGraph, wa *trace.ElementWait) {
	id := wa.GetObjId()

	lw := baseA.LastChangeWG[id]
	if lw != nil {
		if graph != nil {
			graph.AddEdge(lw, wa)
		} else {
			AddEdge(lw, wa, false)
		}
	}
	baseA.LastChangeWG[id] = wa
}

// Wait updates the pog for a wait operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - wa *TraceElementWait: The trace element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func Wait(graph *PoGraph, wa *trace.ElementWait, recorded bool) {
	if recorded && wa.GetTPost() == 0 {
		return
	}

	id := wa.GetObjId()

	lc := baseA.LastChangeWG[id]
	if lc != nil {
		if graph != nil {
			graph.AddEdge(lc, wa)
		} else {
			AddEdge(lc, wa, false)
		}
	}
}
