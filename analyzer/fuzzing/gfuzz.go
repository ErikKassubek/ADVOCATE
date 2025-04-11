// Copyright (c) 2025 Erik Kassubek
//
// File: gfuzz.go
// Brief: Main function to run gfuzz
//
// Author: Erik Kassubek
// Created: 2025-03-22
//
// License: BSD-3-Clause

package fuzzing

import "analyzer/utils"

/*
* Create new mutations for GFuzz if the previous run was interesting
 */
func createGFuzzMut() {
	// add new mutations based on GFuzz select
	if isInterestingSelect() {
		numberMut := numberMutations()
		flipProb := getFlipProbability()
		numMutAdd := createMutationsGFuzz(numberMut, flipProb)
		utils.LogInfof("Add %d select mutations to queue", numMutAdd)
	} else {
		utils.LogInfo("Add 0 select mutations to queue")
	}
}
