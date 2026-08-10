// Copyright (c) 2024 Erik Kassubek
//
// File: vcWait.go
// Brief: Update functions for happens before info for wait group operations
//        Some function start analysis functions
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"gocct/analysis/a_base"
	"gocct/analysis/analysis/a_scenarios"
	"gocct/analysis/hb/a_hbcalc"
	"gocct/fuzzing/f_base"
	"gocct/trace"
	"gocct/utils/flags"
	"gocct/utils/log"
)

// AnalyzeWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func AnalyzeWait(wa *trace.ElementWait) {
	a_hbcalc.UpdateHBWait(wa)

	switch wa.Type(true) {
	case trace.WaitAdd, trace.WaitDone:
		a_base.LastChangeWG[wa.ObjID()] = wa

		if a_base.AnalysisCasesMap[flags.DoneBeforeAdd] || f_base.FuzzingModeGoCRHBPlus {
			a_scenarios.CheckForDoneBeforeAddChange(wa)
		}
	case trace.WaitWait:
	default:
		err := "Unknown operation on wait group: " + wa.String()
		log.Error(err)
	}
}
