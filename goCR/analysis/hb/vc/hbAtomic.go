//
// File: hbAtomic.go
// Brief: Update the vc for atomics
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package vc

import (
	"goCR/analysis/data"
	"goCR/trace"
)

// UpdateHBAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateHBAtomic(at *trace.ElementAtomic) {
	routine := at.GetRoutine()

	at.SetVc(CurrentVC[routine])
	at.SetWVc(CurrentWVC[routine])

	switch at.GetOpA() {
	case trace.LoadOp:
		Read(at, true, routine)
	case trace.StoreOp, trace.AddOp, trace.AndOp, trace.OrOp:
		Write(at, routine)
	case trace.SwapOp, trace.CompSwapOp:
		Swap(at, true, routine)
	default:

	}
}

// Write calculates the new vector clock for a write operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - routine int: the routine of at
func Write(at *trace.ElementAtomic, routine int) {
	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// Read calculates the new vector clock for a read operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
//   - routine int: the routine of at
func Read(at *trace.ElementAtomic, sync bool, routine int) {
	id := at.GetID()

	if sync && data.LastAtomicWriter[id] != nil {
		CurrentVC[routine].Sync(data.LastAtomicWriter[id].GetVC())
	}
}

// Swap calculate the new vector clock for a swap operation and update cv. A swap
// operation is a read and a write.
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
//   - routine int: the routine of at
func Swap(at *trace.ElementAtomic, sync bool, routine int) {
	Read(at, sync, routine)
	Write(at, routine)
}
