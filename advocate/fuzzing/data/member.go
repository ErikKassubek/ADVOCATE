// Copyright (c) 2025 Erik Kassubek
//
// File: member.go
// Brief: Member function on the types in fuzzing/data/types.go
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package data

import "math/rand"

// Get a copy of fs with a randomly selected case id.
//
// Parameter:
//   - def bool: if true, default is a possible value, if false it is not
//   - flipChange bool: probability that a select case is chosen randomly. Otherwise the chosen case is kept
//
// Returns:
//   - int: the chosen case ID
func (fs FuzzingSelect) GetCopyRandom(def bool, flipChance float64) FuzzingSelect {
	// do only flip with certain chance
	if rand.Float64() > flipChance {
		return FuzzingSelect{Id: fs.Id, T: fs.T, ChosenCase: fs.ChosenCase, NumberCases: fs.NumberCases, ContainsDefault: fs.ContainsDefault}
	}

	// if at most one case and no default (should not happen), or only default select the same case again
	if (!def && fs.NumberCases <= 1) || (def && fs.NumberCases == 0) {
		return FuzzingSelect{Id: fs.Id, T: fs.T, ChosenCase: fs.ChosenCase, NumberCases: fs.NumberCases, ContainsDefault: fs.ContainsDefault}
	}

	prefCase := fs.chooseRandomCase(def)

	return FuzzingSelect{Id: fs.Id, T: fs.T, ChosenCase: prefCase, NumberCases: fs.NumberCases, ContainsDefault: fs.ContainsDefault}
}

// Randomly select a case.
// The case is between 0 and fs.numberCases if def is false and between -1 and fs.numberCases otherwise
// fs.chosenCase is never chosen
// The values in fs.casiWithPos have a higher likelihood to be chosen by a factor factorCaseWithPartner (defined in fuzzing/data.go)
//
// Parameter
//   - def bool: true if the select contains a bool
//
// Returns:
//   - the chosen case id
func (fs FuzzingSelect) chooseRandomCase(def bool) int {
	// Determine the starting number based on includeZero
	start := 0
	if def {
		start = -1
	}

	// Create a weight map for the probabilities
	weights := make(map[int]int)

	// Assign weights to each number
	for i := start; i < fs.NumberCases; i++ {
		weights[i] = 1 // Default weight
	}

	// Increase weights for numbers in fs.casiWithPos
	for _, num := range fs.CasiWithPos {
		if num >= start && num < fs.NumberCases && num != fs.ChosenCase {
			weights[num] *= factorCaseWithPartner
		}
	}

	// Generate a cumulative weight array
	cumulativeWeights := []int{}
	numbers := []int{} // Keep track of the corresponding numbers
	totalWeight := 0

	for i := start; i < fs.NumberCases; i++ {
		if weight, exists := weights[i]; exists && weight > 0 {
			totalWeight += weight
			cumulativeWeights = append(cumulativeWeights, totalWeight)
			numbers = append(numbers, i)
		}
	}

	// Handle edge case where no valid number can be chosen
	if totalWeight == 0 {
		return 0
	}

	r := rand.Intn(totalWeight)

	// Find the number corresponding to the random weight
	for i, cw := range cumulativeWeights {
		if r < cw {
			return numbers[i]
		}
	}

	// Fallback (should never reach here)
	return 0
}
