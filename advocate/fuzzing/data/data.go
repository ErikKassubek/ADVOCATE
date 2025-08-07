// Copyright (c) 2024 Erik Kassubek
//
// File: data.go
// Brief: File to define and contain the fuzzing data
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package data

import (
	"advocate/results/results"

	"time"
)

const factorCaseWithPartner = 3

// General data for fuzzing
var (
	MaxNumberRuns     = 100
	MaxTime           = 7 * time.Minute
	MaxTimeSet        = false
	NumberFuzzingRuns = 0
	MutationQueue     = make([]Mutation, 0)

	// count how often a specific mutation has been in the queue
	AllMutations           = make(map[string]int)
	FuzzingMode            = ""
	FuzzingModeGFuzz       = false
	FuzzingModeGoPie       = false
	FuzzingModeGoPieHBPlus = false
	FuzzingModeFlow        = false

	CancelTestIfBugFound = false

	NumberOfPreviousRuns = 0

	UseHBInfoFuzzing = true
	runFullAnalysis  = true

	FinishIfBugFound = false
)

// ClearDataFull resets the fuzzing data that is unique for each test but used for each fuzzing
// run of a test
func ClearDataFull() {
	results.Reset()

	NumberOfPreviousRuns = 0
}
