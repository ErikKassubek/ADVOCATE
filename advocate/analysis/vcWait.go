// Copyright (c) 2024 Erik Kassubek
//
// File: vcWait.go
// Brief: Update functions of vector groups for wait group operations
//        Some function start analysis functions
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

// Create a new wg if needed
//
// Parameter:
//   - index int: The id of the wait group
//   - nRout int: The number of routines in the trace
func newWg(index int, nRout int) {
	if _, ok := lastChangeWG[index]; !ok {
		lastChangeWG[index] = clock.NewVectorClock(nRout)
	}
}

// UpdateVCWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func UpdateVCWait(wa *trace.TraceElementWait) {
	routine := wa.GetRoutine()
	wa.SetVc(currentVC[routine])
	wa.SetWVc(currentWVC[routine])

	switch wa.GetOpW() {
	case trace.ChangeOp:
		Change(wa)
	case trace.WaitOp:
		Wait(wa)
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		utils.LogError(err)
	}
}

// Change calculate the new vector clock for a add or done operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Change(wa *trace.TraceElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := wa.GetID()
	routine := wa.GetRoutine()

	newWg(id, currentVC[routine].GetSize())
	lastChangeWG[id].Sync(currentVC[routine])

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["doneBeforeAdd"] {
		checkForDoneBeforeAddChange(wa)
	}
}

// Wait calculates the new vector clock for a wait operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Wait(wa *trace.TraceElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := wa.GetID()
	routine := wa.GetRoutine()

	newWg(id, currentVC[routine].GetSize())

	if wa.GetTPost() != 0 {
		currentVC[routine].Sync(lastChangeWG[id])
	}

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}
