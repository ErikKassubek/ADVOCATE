// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data for GoPie
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package gopie

import (
	"advocate/trace"
	"advocate/utils/helper"
)

var (
	// store all created mutations to avoid doubling
	allGoPieMutations = make(map[string]struct{})

	numberWrittenGoPieMuts = 0
	maxGoPieScore          = 0

	// for each mutation file, store the file number and the chain
	chainFiles = make(map[int]Chain)

	// number of different starting points for chains in GoPie (in the original: cfg.MaxWorker)
	maxSCStart = helper.GoPieSCStart

	SchedulingChains []Chain
	CurrentChain     Chain
	LastRoutine      = -1

	// GoPie relations
	counterCPOP1 = 0
	counterCPOP2 = 0
	rel1         = make(map[trace.Element]map[trace.Element]struct{})
	rel2         = make(map[trace.Element]map[trace.Element]struct{})

	ElemsByID = make(map[int][]trace.Element) // id -> chan/sel/mutex elem
)

func ClearData() {
	rel1 = make(map[trace.Element]map[trace.Element]struct{})
	rel2 = make(map[trace.Element]map[trace.Element]struct{})
	counterCPOP1 = 0
	counterCPOP2 = 0
	ElemsByID = make(map[int][]trace.Element)
	numberWrittenGoPieMuts = 0
	maxGoPieScore = 0
}
