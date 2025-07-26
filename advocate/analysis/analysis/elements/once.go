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
	"advocate/analysis/data"
	"advocate/analysis/hb/hbCalc"
	"advocate/trace"
)

// AnalyzeOnce update the hb info of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func AnalyzeOnce(on *trace.ElementOnce) {
	hbCalc.UpdateHBOnce(on)

	if on.GetSuc() {
		id := on.GetID()
		data.OSuc[id] = on
	}
}
