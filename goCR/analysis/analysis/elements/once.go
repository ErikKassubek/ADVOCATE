//
// File: vcOnce.go
// Brief: Update functions for happens before info for once operations
//
// Created: 2023-07-25
//
// License: BSD-3-Clause

package elements

import (
	"goCR/analysis/data"
	"goCR/analysis/hb/hbcalc"
	"goCR/trace"
)

// AnalyzeOnce update the hb info of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func AnalyzeOnce(on *trace.ElementOnce) {
	hbcalc.UpdateHBOnce(on)

	if on.GetSuc() {
		id := on.GetID()
		data.OSuc[id] = on
	}
}
