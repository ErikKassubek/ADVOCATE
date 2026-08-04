// Copyright (c) 2025 Erik Kassubek
//
// File: vcOther.go
// Brief: Function for happens before info for
//   elements that do not change, but only store the vc
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_hbcalc"
	"advocate/trace"
)

// AnalyzeNew store the vector clock of the element
// Parameter:
//   - n *trace.TraceElementAlloc: the new trace element
func AnalyzeNew(n *trace.ElementAlloc) {
	a_hbcalc.UpdateHBNew(n)

	a_base.NewChan[n.ObjID()] = n.File()
}

// AnalyzeRoutineEnd store the vector clock of the element
// Parameter:
//   - re *trace.TraceElementRoutineEnd: the new trace element
func AnalyzeRoutineEnd(re *trace.ElementRoutineEnd) {
	a_hbcalc.UpdateHBRoutineEnd(re)
}
