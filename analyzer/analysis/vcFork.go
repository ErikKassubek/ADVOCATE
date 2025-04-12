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
	"analyzer/timer"
)

// Update the vector clocks given a fork operation
//
// Parameter:
//   - oldRout (int): The id of the old routine
//   - newRout (int): The id of the new routine
func Fork(fo *TraceElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	oldRout := fo.routine
	newRout := fo.id

	currentVC[newRout] = currentVC[oldRout].Copy()
	currentVC[oldRout].Inc(oldRout)
	currentVC[newRout].Inc(newRout)

	currentWVC[newRout] = currentWVC[oldRout].Copy()
	currentWVC[oldRout].Inc(oldRout)
	currentWVC[newRout].Inc(newRout)

	allForks[fo.id] = fo
}
