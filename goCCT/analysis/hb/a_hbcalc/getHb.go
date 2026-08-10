// Copyright (c) 2025 Erik Kassubek
//
// File: getHb.go
// Brief: Get the HB info between two elements
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_hbcalc

import (
	"gocct/analysis/a_hb"
	"gocct/analysis/hb/a_cssts"
	"gocct/analysis/hb/a_pog"
	"gocct/analysis/hb/a_vc"
	"gocct/trace"
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
func GetHappensBefore(e1, e2 trace.Element, weak bool) a_hb.HappensBefore {
	if CalcVC {
		return a_vc.GetHappensBefore(e1, e2, weak)
	}

	if CalcPog {
		return a_pog.GetHappensBefore(e1, e2, weak)
	}

	if CalcCssts {
		return a_cssts.GetHappensBefore(e1, e2, weak)
	}

	return a_hb.None
}
