// Copyright (c) 2024 Erik Kassubek
//
// File: hbFork.go
// Brief: Update function for happens before info for forks (creation of new routine)
//
// Author: Erik Kassubek
// Created: 2023-07-26
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/concurrent/cssts"
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/timer"
)

// UpdateHBFork update and calculate happens before information for fork operations
// It only calculates the VC and csst, not the pog, which is included in the
// edge creation of elements in the same routine
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func UpdateHBFork(fo *trace.ElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	routine := fo.GetRoutine()

	fo.SetVc(data.CurrentVC[routine])
	fo.SetWVc(data.CurrentWVC[routine])

	oldRout := fo.GetRoutine()
	newRout := fo.GetID()

	data.CurrentVC[newRout] = data.CurrentVC[oldRout].Copy()
	data.CurrentVC[oldRout].Inc(oldRout)
	data.CurrentVC[newRout].Inc(newRout)

	data.CurrentWVC[newRout] = data.CurrentWVC[oldRout].Copy()
	data.CurrentWVC[oldRout].Inc(oldRout)
	data.CurrentWVC[newRout].Inc(newRout)

	data.ForkOperations[fo.GetID()] = fo

	cssts.AddEdgeFork(fo)
}
