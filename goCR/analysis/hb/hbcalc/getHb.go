// Copyright (c) 2025 Erik Kassubek
//
// File: getHb.go
// Brief: Get the HB info between two elements
//
// Author: Erik Kassubek
// Created: 2025-07-21
//
// License: BSD-3-Clause

package hbcalc

import (
	"goCR/analysis/hb"
	"goCR/analysis/hb/vc"
	"goCR/trace"
)

// GetHappensBefore returns the happens before relation between two operations
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - weak bool: get based on weak happens before
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(e1, e2 trace.Element, weak bool) hb.HappensBefore {
	return vc.GetHappensBefore(e1, e2, weak)
}
