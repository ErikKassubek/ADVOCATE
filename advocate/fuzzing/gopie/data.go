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
	"advocate/utils/settings.go"
	"advocate/utils/types"
)

// Data for goPie fuzzing
var (
	// store all created mutations to avoid doubling
	allGoPieMutations = make(map[string]struct{})

	numberWrittenGoPieMuts = 0
	maxGoPieScore          = 0

	// for each mutation file, store the file number and the chain
	chainFiles = make(map[int]Chain)

	// number of different starting points for chains in GoPie (in the original: cfg.MaxWorker)
	maxSCStart = settings.GoPieSCStart

	SchedulingChains []Chain
	CurrentChain     Chain
	LastRoutine      = -1

	// GoPie relations
	counterCPOP1 = 0
	counterCPOP2 = 0
	rel1         = make(map[trace.Element]map[trace.Element]struct{})
	rel2         = make(map[trace.Element]map[trace.Element]struct{})

	ElemsByID = make(map[int][]trace.Element) // id -> chan/sel/mutex elem

	usedStartPos = make([]types.Pair[trace.Element, trace.Element], 0)

	NumberInvalidMuts = 0
	NumberDoubleMuts  = 0
	NumberTotalMuts   = 0
)

// ClearDataRun the data for one run in GoPie
func ClearDataRun() {
	rel1 = make(map[trace.Element]map[trace.Element]struct{})
	rel2 = make(map[trace.Element]map[trace.Element]struct{})
	counterCPOP1 = 0
	counterCPOP2 = 0
	ElemsByID = make(map[int][]trace.Element)
}

// ClearData deletes all the GoPie data
func ClearData() {
	ClearDataRun()
	maxGoPieScore = 0
	usedStartPos = make([]types.Pair[trace.Element, trace.Element], 0)

	NumberInvalidMuts = 0
	NumberTotalMuts = 0
	NumberDoubleMuts = 0
}
