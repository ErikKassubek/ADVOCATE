//
// File: hbAtomic.go
// Brief: Update the pog for atomics
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package pog

import (
	"goCR/analysis/data"
	"goCR/trace"
	"goCR/utils/log"
)

// UpdateHBAtomic update the pog for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
//   - alt bool: Store and update the vector clock of the element if the IgnoreCriticalSections tag has been set
func UpdateHBAtomic(at *trace.ElementAtomic, alt bool) {
	switch at.GetOpA() {
	case trace.LoadOp, trace.SwapOp, trace.CompSwapOp:
		Read(at, true)
	case trace.StoreOp, trace.AddOp, trace.AndOp, trace.OrOp:
		// pog does not add an edge for write
	default:
		err := "Unknown operation: " + at.ToString()
		log.Error(err)
	}
}

// Read calculates the new vector clock for a read operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
func Read(at *trace.ElementAtomic, sync bool) {
	id := at.GetID()

	if sync && data.LastAtomicWriter[id] != nil {
		AddEdge(at, data.LastAtomicWriter[id], false)
	}
}
