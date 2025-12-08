// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the pog for atomics
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

// UpdateHBAtomic update the pog for an atomic operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateHBAtomic(graph *PoGraph, at *trace.ElementAtomic) {
	switch at.GetType(true) {
	case trace.AtomicLoad, trace.AtomicSwap, trace.AtomicCompAndSwap:
		Read(graph, at, true)
	case trace.AtomicStore, trace.AtomicAdd, trace.AtomicAnd, trace.AtomicOr:
		// pog does not add an edge for write
	default:
		err := "Unknown operation: " + at.ToString()
		log.Error(err)
	}
}

// Read calculates the new vector clock for a read operation and update cv
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
func Read(graph *PoGraph, at *trace.ElementAtomic, sync bool) {
	id := at.GetObjId()

	if sync && baseA.LastAtomicWriter[id] != nil {
		if graph != nil {
			graph.AddEdge(at, baseA.LastAtomicWriter[id])
		} else {
			AddEdge(at, baseA.LastAtomicWriter[id], false)
		}
	}
}
