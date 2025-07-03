// Copyright (c) 2024 Erik Kassubek
//
// File: vcAtomic.go
// Brief: Update for vector clocks from atomic operations
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// UpdateVCAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateVCAtomic(at *trace.ElementAtomic) {

	routine := at.GetRoutine()

	at.SetVc(data.CurrentVC[routine])
	at.SetWVc(data.CurrentWVC[routine])

	switch at.GetOpA() {
	case trace.LoadOp:
		Read(at, true)
	case trace.StoreOp, trace.AddOp, trace.AndOp, trace.OrOp:
		Write(at)
	case trace.SwapOp, trace.CompSwapOp:
		Swap(at, true)
	default:
		err := "Unknown operation: " + at.ToString()
		log.Error(err)
	}
}

// Store and update the vector clock of the element if the IgnoreCriticalSections
// tag has been set
func UpdateVCAtomicAlt(at *trace.ElementAtomic) {
	at.SetVc(data.CurrentVC[at.GetRoutine()])

	switch at.GetOpA() {
	case trace.LoadOp:
		Read(at, false)
	case trace.StoreOp, trace.AddOp, trace.AndOp, trace.OrOp:
		Write(at)
	case trace.SwapOp, trace.CompSwapOp:
		Swap(at, false)
	default:
		err := "Unknown operation: " + at.ToString()
		log.Error(err)
	}
}

// Write calculates the new vector clock for a write operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
func Write(at *trace.ElementAtomic) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := at.GetID()
	routine := at.GetRoutine()

	data.Lw[id] = data.CurrentVC[routine].Copy()

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}

// Read calculates the new vector clock for a read operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
func Read(at *trace.ElementAtomic, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := at.GetID()
	routine := at.GetRoutine()

	data.NewLW(id, data.CurrentVC[routine].GetSize())
	if sync {
		data.CurrentVC[routine].Sync(data.Lw[id])
	}

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}

// Swap calculate the new vector clock for a swap operation and update cv. A swap
// operation is a read and a write.
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
func Swap(at *trace.ElementAtomic, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	Read(at, sync)
	Write(at)
}
