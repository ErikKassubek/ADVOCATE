// Copyright (c) 2025 Erik Kassubek
//
// File: settings.go
// Brief: Settings for hb calculation
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package hb

var (
	calcVC    = false
	calcPog   = false
	calcCssts = false
)

// SetHBSettings sets which hb structure should be calculated
//
// Parameter:
//   - vc bool: calculate vector clocks
//   - pog bool: calculate partial order graph
//   - cssts bool: calculate Collective Sparse Segment Trees
func SetHbSettings(vc, pog, cssts bool) {
	calcVC = vc
	calcPog = pog
	calcCssts = cssts
}
