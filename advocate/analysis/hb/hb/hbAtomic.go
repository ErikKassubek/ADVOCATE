// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update hb info for atomics
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package hb

import (
	"advocate/analysis/hb/cssts"
	"advocate/analysis/hb/pog"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/timer"
)

// UpdateHBAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
//   - alt bool: update if the ignoreCriticalSections tag has been set
func UpdateHBAtomic(at *trace.ElementAtomic, alt bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if calcVC {
		vc.UpdateHBAtomic(at, alt)
	}

	if calcPog {
		pog.UpdateHBAtomic(at, alt)
	}

	if calcCssts {
		cssts.UpdateHBAtomic(at, alt)
	}
}
