// Copyright (c) 2025 Erik Kassubek
//
// File: atomic.go
// Brief: Update the data for an atomic element
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"gocdr/analysis/hb/a_hbcalc"
	"gocdr/trace"
)

// AnalyzeAtomic update the hb info for an atomic event
//
// Parameter:
//   - at *trace.ElementAtomic: the element
func AnalyzeAtomic(at *trace.ElementAtomic) {
	a_hbcalc.UpdateHBAtomic(at)
}
