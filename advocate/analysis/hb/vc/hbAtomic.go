// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for atomics
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

// UpdateHBAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateHBAtomic(at *trace.ElementAtomic) {
	routine := at.GetRoutine()

	at.SetVc(CurrentVC[routine])
	at.SetWVc(CurrentWVC[routine])

	switch at.GetType(true) {
	case trace.AtomicLoad:
		Read(at, true, routine)
	case trace.AtomicStore, trace.AtomicAdd, trace.AtomicAnd, trace.AtomicOr:
		Write(at, routine)
	case trace.AtomicSwap, trace.AtomicCompAndSwap:
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
	id := at.GetObjId()

	if sync && baseA.LastAtomicWriter[id] != nil {
		CurrentVC[routine].Sync(baseA.LastAtomicWriter[id].GetVC())
		CurrentWVC[routine].Sync(baseA.LastAtomicWriter[id].GetWVC())
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
