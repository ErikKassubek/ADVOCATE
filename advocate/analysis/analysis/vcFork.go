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
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/timer"
)

// UpdateVCFork update and calculate the vector clock of the element
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func UpdateVCFork(fo *trace.ElementFork) {
	routine := fo.GetRoutine()

	fo.SetVc(data.CurrentVC[routine])
	fo.SetWVc(data.CurrentWVC[routine])

	Fork(fo)
}

// Fork updates the vector clocks given a fork operation
//
// Parameter:
//   - oldRout int: The id of the old routine
//   - newRout int: The id of the new routine
func Fork(fo *trace.ElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	oldRout := fo.GetRoutine()
	newRout := fo.GetID()

	data.CurrentVC[newRout] = data.CurrentVC[oldRout].Copy()
	data.CurrentVC[oldRout].Inc(oldRout)
	data.CurrentVC[newRout].Inc(newRout)

	data.CurrentWVC[newRout] = data.CurrentWVC[oldRout].Copy()
	data.CurrentWVC[oldRout].Inc(oldRout)
	data.CurrentWVC[newRout].Inc(newRout)

	data.ForkOperations[fo.GetID()] = fo
}
