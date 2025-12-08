// Copyright (c) 2025 Erik Kassubek
//
// File: graph.go
// Brief: Create the hb graph for a mutation
//
// Author: Erik Kassubek
// Created: 2025-12-08
//
// License: BSD-3-Clause

package equivalence

import (
	"advocate/analysis/hb/pog"
	"advocate/trace"
)

func (this *TraceEq) BuildPOG() {
	graph := pog.NewPoGraph()

	// TODO: fork
	// TODO: precompute all required values, e.g. last writer, comm partner, ...

	for _, elem := range this.trace {
		switch e := elem.(type) {
		case *trace.ElementAtomic:
			pog.UpdateHBAtomic(&graph, e)
		case *trace.ElementChannel:
			pog.UpdateHBChannel(&graph, e, false)
		case *trace.ElementSelect:
			pog.UpdateHBSelect(&graph, e, false)
		case *trace.ElementCond:
			pog.UpdateHBCond(&graph, e)
		case *trace.ElementMutex:
			pog.UpdateHBMutex(&graph, e, false)
		case *trace.ElementOnce:
			pog.UpdateHBOnce(&graph, e)
		case *trace.ElementWait:
			pog.UpdateHBWait(&graph, e, false)
		}
	}
}
