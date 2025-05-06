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
	"advocate/clock"
	"advocate/timer"
	"advocate/trace"
	"advocate/utils"
)

// UpdateVCAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateVCAtomic(at *trace.TraceElementAtomic) {

	routine := at.GetRoutine()

	at.SetVc(currentVC[routine])
	at.SetWVc(currentWVC[routine])

	switch at.GetOpA() {
	case trace.LoadOp:
		Read(at, true)
	case trace.StoreOp, trace.AddOp, trace.AndOp, trace.OrOp:
		Write(at)
	case trace.SwapOp, trace.CompSwapOp:
		Swap(at, true)
	default:
		err := "Unknown operation: " + at.ToString()
		utils.LogError(err)
	}
}

// Store and update the vector clock of the element if the IgnoreCriticalSections
// tag has been set
func UpdateVCAtomicAlt(at *trace.TraceElementAtomic) {
	at.SetVc(currentVC[at.GetRoutine()])

	switch at.GetOpA() {
	case trace.LoadOp:
		Read(at, false)
	case trace.StoreOp, trace.AddOp, trace.AndOp, trace.OrOp:
		Write(at)
	case trace.SwapOp, trace.CompSwapOp:
		Swap(at, false)
	default:
		err := "Unknown operation: " + at.ToString()
		utils.LogError(err)
	}
}

// Create a new lw if needed
//
// Parameter:
//   - index int: The id of the atomic variable
//   - nRout int: The number of routines in the trace
func newLw(index int, nRout int) {
	if _, ok := lw[index]; !ok {
		lw[index] = clock.NewVectorClock(nRout)
	}
}

// Write calculates the new vector clock for a write operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
func Write(at *trace.TraceElementAtomic) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := at.GetID()
	routine := at.GetRoutine()

	newLw(id, currentVC[id].GetSize())
	lw[id] = currentVC[routine].Copy()

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}

// Read calculates the new vector clock for a read operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
func Read(at *trace.TraceElementAtomic, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := at.GetID()
	routine := at.GetRoutine()

	newLw(id, currentVC[id].GetSize())
	if sync {
		currentVC[routine].Sync(lw[id])
	}

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}

// Swap calculate the new vector clock for a swap operation and update cv. A swap
// operation is a read and a write.
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
func Swap(at *trace.TraceElementAtomic, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	Read(at, sync)
	Write(at)
}
