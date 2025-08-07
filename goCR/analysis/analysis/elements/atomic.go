//
// File: atomic.go
// Brief: Update the data for an atomic element
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package elements

import (
	"goCR/analysis/hb/hbcalc"
	"goCR/trace"
)

// AnalyzeAtomic update the hb info for an atomic event
//
// Parameter:
//   - at *trace.ElementAtomic: the element
//   - alt bool: if ignoreCriticalSections is set
func AnalyzeAtomic(at *trace.ElementAtomic, alt bool) {
	hbcalc.UpdateHBAtomic(at, alt)
}
