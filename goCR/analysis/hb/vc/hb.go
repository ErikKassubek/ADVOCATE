//
// File: hb.go
// Brief: Get the happens before info
//
// Created: 2025-07-21
//
// License: BSD-3-Clause

package vc

import (
	"goCR/analysis/hb"
	"goCR/analysis/hb/clock"
	"goCR/trace"
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
func GetHappensBefore(t1, t2 trace.Element, weak bool) hb.HappensBefore {
	if weak {
		return clock.GetHappensBefore(t1.GetWVC(), t2.GetWVC())
	}
	return clock.GetHappensBefore(t1.GetVC(), t2.GetVC())
}
