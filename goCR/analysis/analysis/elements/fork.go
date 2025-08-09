//
// File: hbFork.go
// Brief: Update function for happens before info for forks (creation of new routine)
//
// Created: 2023-07-26
//
// License: BSD-3-Clause

package elements

import (
	"goCR/analysis/data"
	"goCR/analysis/hb/hbcalc"
	"goCR/trace"
	"goCR/utils/timer"
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

	hbcalc.UpdateHBFork(fo)

	// store fork operations for each routine
	data.ForkOperations[fo.GetID()] = fo
}
