// Copyright (c) 2024 Erik Kassubek
//
// File: vcOnce.go
// Brief: Update functions for happens before info for once operations
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package elements

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb/hbcalc"
	"advocate/trace"
)

// AnalyzeOnce update the hb info of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func AnalyzeOnce(on *trace.ElementOnce) {
	hbcalc.UpdateHBOnce(on)

	if on.GetSuc() {
		id := on.GetObjId()
		baseA.OSuc[id] = on
	}
}
