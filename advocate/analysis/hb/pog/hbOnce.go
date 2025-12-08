// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the pog for once
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

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(graph *PoGraph, on *trace.ElementOnce) {
	// suc once does not create edge -> only not suc
	if !on.GetSuc() {
		suc := baseA.OSuc[on.GetObjId()]
		if graph != nil {
			graph.AddEdge(suc, on)
		} else {
			AddEdge(suc, on, false)
		}
	}
}
