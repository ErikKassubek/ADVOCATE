// Copyright (c) 2025 Erik Kassubek
//
// File: consts.go
// Brief: Consts for fuzzing
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package data

// Possible values for fuzzing mode
const (
	GFuzz       = "GFuzz"       // only GFuzz
	GFuzzHB     = "GFuzzHB"     // GFuzz with use of hb info
	GFuzzHBFlow = "GFuzzHBFlow" // GFuzz with use of hb info and flow mutation
	Flow        = "Flow"        // only flow mutation
	GoPie       = "GoPie"       // only goPie
	GoPiePlus   = "GoPie+"      // improved goPie without HB
	GoPieHB     = "GoPieHB"     // goPie with HB relation
)

const (
	MutSelType  = 0
	MutPiType   = 1
	MutFlowType = 2
)
