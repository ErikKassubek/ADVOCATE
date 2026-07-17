// Copyright (c) 2024 Erik Kassubek
//
// File: vcOnce.go
// Brief: Update functions for happens before info for once operations
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package a_elements

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_hbcalc"
	"advocate/trace"
)

// AnalyzeOnce update the hb info of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func AnalyzeOnce(on *trace.ElementOnce) {
	a_hbcalc.UpdateHBOnce(on)

	if on.GetSuc() {
		id := on.ObjID()
		a_base.OSuc[id] = on
	}
}
