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
	"analyzer/clock"
	"analyzer/timer"
)

// Create a new lw if needed
//
// Parameter:
//   - index (int): The id of the atomic variable
//   - nRout (int): The number of routines in the trace
func newLw(index int, nRout int) {
	if _, ok := lw[index]; !ok {
		lw[index] = clock.NewVectorClock(nRout)
	}
}

// Calculate the new vector clock for a write operation and update cv
//
// Parameter:
//   - at (*TraceElementAtomic): The trace element
func Write(at *TraceElementAtomic) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newLw(at.id, currentVC[at.id].GetSize())
	lw[at.id] = currentVC[at.routine].Copy()

	currentVC[at.routine].Inc(at.routine)
	currentWVC[at.routine].Inc(at.routine)
}

// Calculate the new vector clock for a read operation and update cv
//
// Parameter:
//   - at (*TraceElementAtomic): The trace element
//   - numberOfRoutines (int): The number of routines in the trace
//   - sync bool: sync reader with last writer
func Read(at *TraceElementAtomic, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	newLw(at.id, currentVC[at.id].GetSize())
	if sync {
		currentVC[at.routine].Sync(lw[at.id])
	}

	currentVC[at.routine].Inc(at.routine)
	currentWVC[at.routine].Inc(at.routine)
}

// Calculate the new vector clock for a swap operation and update cv. A swap
// operation is a read and a write.
//
// Parameter:
//   - at (*TraceElementAtomic): The trace element
//   - numberOfRoutines (int): The number of routines in the trace
//   - sync bool: sync reader with last writer
func Swap(at *TraceElementAtomic, sync bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	Read(at, sync)
	Write(at)
}
