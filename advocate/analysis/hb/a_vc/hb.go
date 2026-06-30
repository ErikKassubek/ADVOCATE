// Copyright (c) 2025 Erik Kassubek
//
// File: hb.go
// Brief: Get the happens before info
//
// Author: Erik Kassubek
// Created: 2025-07-21
//
// License: BSD-3-Clause

package a_vc

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
)

// GetHappensBefore returns the happens before relation between two operations given there
// vc
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - weak bool: get based on weak happens before
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(t1, t2 trace.Element, weak bool) a_hb.HappensBefore {
	if weak {
		return a_clock.GetHappensBefore(t1.GetWVC(), t2.GetWVC())
	}
	return a_clock.GetHappensBefore(t1.GetVC(), t2.GetVC())
}
