//
// File: consts.go
// Brief: Consts for fuzzing
//
// Created: 2025-07-03
//
// License: BSD-3-Clause

package data

// Possible values for fuzzing mode
const (
	GoPie = "GoPie" // only goPie
	GFuzz = "GFuzz" // only GFuzz
	GoCR  = "GoCR"  // improved goPie without predictive analysis
)

// Possible mut types
const (
	MutSelType = 0
	MutPiType  = 1
)
