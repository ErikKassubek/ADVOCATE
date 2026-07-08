// Copyright (c) 2026 Erik Kassubek
//
// File: atomic.go
// Brief: Update the data for an atomic element
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package a_elements

import (
	"advocate/analysis/hb/a_hbcalc"
	"advocate/trace"
)

// AnalyzeAtomic update the hb info for an atomic event
//
// Parameter:
//   - at *trace.ElementAtomic: the element
func AnalyzeAtomic(at *trace.ElementAtomic) {
	a_hbcalc.UpdateHBAtomic(at)
}
