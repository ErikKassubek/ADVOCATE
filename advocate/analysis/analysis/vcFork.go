// Copyright (c) 2024 Erik Kassubek
//
// File: vcFork.go
// Brief: Update function for vector clocks from forks (creation of new routine)
//
// Author: Erik Kassubek
// Created: 2023-07-26
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/trace"
	"advocate/utils/timer"
)

// UpdateVCFork update and calculate the vector clock of the element
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func UpdateVCFork(fo *trace.TraceElementFork) {
	routine := fo.GetRoutine()

	fo.SetVc(currentVC[routine])
	fo.SetWVc(currentWVC[routine])

	Fork(fo)
}

// Fork updates the vector clocks given a fork operation
//
// Parameter:
//   - oldRout int: The id of the old routine
//   - newRout int: The id of the new routine
func Fork(fo *trace.TraceElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	oldRout := fo.GetRoutine()
	newRout := fo.GetID()

	currentVC[newRout] = currentVC[oldRout].Copy()
	currentVC[oldRout].Inc(oldRout)
	currentVC[newRout].Inc(newRout)

	currentWVC[newRout] = currentWVC[oldRout].Copy()
	currentWVC[oldRout].Inc(oldRout)
	currentWVC[newRout].Inc(newRout)

	forkOperations[fo.GetID()] = fo
}
