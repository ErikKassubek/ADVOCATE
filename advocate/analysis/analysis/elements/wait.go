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
	"advocate/analysis/data"
	"advocate/analysis/hb/hbCalc"
	"advocate/trace"
	"advocate/utils/log"
)

// AnalyzeWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func AnalyzeWait(wa *trace.ElementWait) {
	hbCalc.UpdateHBWait(wa)

	switch wa.GetOpW() {
	case trace.ChangeOp:
		data.LastChangeWG[wa.GetID()] = wa

		if data.AnalysisCasesMap[data.DoneBeforeAdd] {
			scenarios.CheckForDoneBeforeAddChange(wa)
		}
	case trace.WaitOp:
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		log.Error(err)
	}
}
