// Copyright (c) 2024 Erik Kassubek
//
// File: score.go
// Brief: Functions to compute the score for fuzzing
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package fuzzing

import (
	"advocate/utils"
	"math"
)

// Calculate how many gFuzz mutations should be created for a given
// trace
//
// Returns:
//   - int: the number of mutations
func numberMutations() int {
	score := calculateScore()
	maxGFuzzScore = math.Max(score, maxGFuzzScore)

	return int(math.Ceil(5.0 * score / maxGFuzzScore))
}

// Calculate the score of the given run
func calculateScore() float64 {
	fact1 := utils.GFuzzW1
	fact2 := utils.GFuzzW2
	fact3 := utils.GFuzzW3
	fact4 := utils.GFuzzW4

	res := 0.0

	// number of communications per communication pair (countChOpPair)
	for _, pair := range pairInfoTrace {
		res += math.Log2(float64(pair.com))
	}

	// number of channels created (createCh)
	res += fact1 * float64(len(channelInfoTrace))

	// number of close (closeCh)
	res += fact2 * float64(numberClose)

	// maximum buffer size for each chan (maxChBufFull)
	bufFullSum := 0.0
	for _, ch := range channelInfoFile {
		bufFullSum += float64(ch.maxQCount)
	}
	res += fact3 * bufFullSum

	if useHBInfoFuzzing {
		// number of select cases with possible partner (both executed and not executed)
		res += fact4 * float64(numberSelectCasesWithPartner)
	}

	return res
}
