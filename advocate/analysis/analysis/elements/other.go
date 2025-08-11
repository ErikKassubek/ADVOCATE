// Copyright (c) 2025 Erik Kassubek
//
// File: vcOther.go
// Brief: Function for happens before info for
//   elements that do not change, but only store the vc
//
// Author: Erik Kassubek
// Created: 2025-04-26
//
// License: BSD-3-Clause

package elements

import (
	"advocate/analysis/data"
	"advocate/analysis/hb/hbcalc"
	"advocate/trace"
)

// AnalyzeNew store the vector clock of the element
// Parameter:
//   - n *trace.TraceElementNew: the new trace element
func AnalyzeNew(n *trace.ElementNew) {
	hbcalc.UpdateHBNew(n)

	data.NewChan[n.GetID()] = n.GetFile()
}

// AnalyzeRoutineEnd store the vector clock of the element
// Parameter:
//   - re *trace.TraceElementRoutineEnd: the new trace element
func AnalyzeRoutineEnd(re *trace.ElementRoutineEnd) {
	hbcalc.UpdateHBRoutineEnd(re)
}
