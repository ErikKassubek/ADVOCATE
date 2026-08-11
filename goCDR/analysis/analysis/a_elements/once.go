// Copyright (c) 2024 Erik Kassubek
//
// File: vcOnce.go
// Brief: Update functions for happens before info for once operations
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"gocdr/analysis/a_base"
	"gocdr/analysis/hb/a_hbcalc"
	"gocdr/trace"
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
