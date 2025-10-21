// Copyright (c) 2024 Erik Kassubek
//
// File: data.go
// Brief: File to define and contain the fuzzing data
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package baseF

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
	AllMutations          = make(map[string]int)
	FuzzingModeGFuzz      = false
	FuzzingModeGoPie      = false
	FuzzingModeGoCRHBPlus = false
	FuzzingModeFlow       = false
	FuzzingModeGuided     = false
	FuzzingHbAnalysis     = true

	NumberOfPreviousRuns = 0

	UseHBInfoFuzzing = true

	FinishIfBugFound = false

	NumberWrittenMutations = 0
	// for each mutation file, store the file number and the chain
	ChainFiles = make(map[int]Chain)
)

// ClearDataFull resets the fuzzing data that is unique for each test but used for each fuzzing
// run of a test
func ClearDataFull() {
	results.Reset()

	NumberOfPreviousRuns = 0
}
