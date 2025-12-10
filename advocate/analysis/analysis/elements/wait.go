// Copyright (c) 2024 Erik Kassubek
//
// File: vcWait.go
// Brief: Update functions for happens before info for wait group operations
//        Some function start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package elements

import (
	"advocate/analysis/analysis/scenarios"
	"advocate/analysis/baseA"
	"advocate/analysis/hb/hbcalc"
	"advocate/fuzzing/baseF"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
)

// AnalyzeWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func AnalyzeWait(wa *trace.ElementWait) {
	hbcalc.UpdateHBWait(wa)

	switch wa.GetType(true) {
	case trace.WaitAdd, trace.WaitDone:
		baseA.LastChangeWG[wa.GetObjId()] = wa

		if baseA.AnalysisCasesMap[flags.DoneBeforeAdd] || baseF.FuzzingModeGoCRHBPlus {
			scenarios.CheckForDoneBeforeAddChange(wa)
		}
	case trace.WaitWait:
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		log.Error(err)
	}
}
