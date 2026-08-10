// Copyright (c) 2024 Erik Kassubek
//
// File: hbFork.go
// Brief: Update function for happens before info for forks (creation of new routine)
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"gocct/analysis/a_base"
	"gocct/analysis/hb/a_hbcalc"
	"gocct/trace"
	"gocct/utils/timer"
)

// AnalyzeFork update and calculate happens before information for fork operations
// It only calculates the VC and csst, not the pog, which is included in the
// edge creation of elements in the same routine
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func AnalyzeFork(fo *trace.ElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	a_hbcalc.UpdateHBFork(fo)

	// store fork operations for each routine
	a_base.ForkOperations[fo.ObjID()] = fo
}
