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

import "math"

func numberMutations() int {
	score := calculateScore()
	maxScore = math.Max(score, maxScore)
	res := 0.0
	if maxScore == 0 {
		res = 5.0 * score
	} else {
		res = math.Ceil(5.0 * score / maxScore)
	}
	return int(res)
}

/*
 * Calculate the score of the given run
 * score = sum log2 countChOpPair + 10 * createCh + 10 * closeCh + 10 * sum maxChBufFull
 */
func calculateScore() float64 {
	const fact1 = 10.0
	const fact2 = 10.0

	res := 0.0

	// number of communications per communication pair (countChOpPair)
	for _, pair := range pairInfoTrace {
		res += math.Log2(float64(pair.com))
	}

	// number of channels created (createCh)
	res += fact1 * float64(len(channelInfoTrace))

	// number of close (closeCh)
	res += fact1 * float64(numberClose)

	// maximum buffer size for each chan (maxChBufFull)
	bufFullSum := 0.0
	for _, ch := range channelInfoFile {
		bufFullSum += float64(ch.maxQCount)
	}
	res += fact1 * bufFullSum

	if useHBInfoFuzzing {
		res += fact2 * float64(numberSelectCasesWithPartner)
	}

	return res
}
