//
// File: vcWait.go
// Brief: Update functions for happens before info for wait group operations
//        Some function start analysis functions
//
// Created: 2023-07-25
//
// License: BSD-3-Clause

package elements

import (
	"goCR/analysis/data"
	"goCR/analysis/hb/hbcalc"
	"goCR/trace"
	"goCR/utils/log"
)

// AnalyzeWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func AnalyzeWait(wa *trace.ElementWait) {
	hbcalc.UpdateHBWait(wa)

	switch wa.GetOpW() {
	case trace.ChangeOp:
		data.LastChangeWG[wa.GetID()] = wa
	case trace.WaitOp:
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		log.Error(err)
	}
}
