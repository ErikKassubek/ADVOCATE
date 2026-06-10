// Copyright (c) 2026
//
// File: consts.go
// Brief: Consts for fuzzing
//
// Created: 2025-07-03
//
// License: BSD-3-Clause

package baseF

// Possible values for fuzzing mode
const (
	Default = ""         // not set
	GFuzz   = "GFuzz"    // only GFuzz
	GFuzzHB = "GFuzzHB"  // GFuzz with use of hb info
	GoPie   = "GoPie"    // only goPie
	GoCR    = "GoCRNoHb" // improved goPie without predictive analysis
	GoCRHB  = "GoCR"     // improved goPie with predictive analysis
	Guided  = "Guided"   // hb guided fuzzing
)

// Possible mut types
const (
	MutSelType  = 0
	MutPiType   = 1
	MutFlowType = 2
)
