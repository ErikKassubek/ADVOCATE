// Copyright (c) 2024 Erik Kassubek
//
// File: select.go
// Brief: File for the selects for fuzzing
//
// Author: Erik Kassubek
// Created: 2024-12-04
//
// License: BSD-3-Clause

package fuzzing

import (
	"advocate/trace"
	"math/rand"
	"sort"
)

// Struct to handle the selects for fuzzing
//
//   - id string: replay id
//   - t int: tPost of the select execution, used for order
//   - chosenCase int: id of the chosen case, -1 for default
//   - numberCases int: number of cases not including default
//   - containsDefault bool: true if contains default case, otherwise false
//   - casiWithPos[]int: list of casi with possible partner
type fuzzingSelect struct {
	id              string
	t               int
	chosenCase      int
	numberCases     int
	containsDefault bool
	casiWithPos     []int
}

// Add a select to selectInfoTrace
//
// Parameter:
//   - e *trace.TraceElementSelect: the select trace element to add
func addFuzzingSelect(e *trace.TraceElementSelect) {
	fs := fuzzingSelect{
		id:              e.GetReplayID(),
		t:               e.GetTPost(),
		chosenCase:      e.GetChosenIndex(),
		numberCases:     len(e.GetCases()),
		containsDefault: e.GetContainsDefault(),
		casiWithPos:     e.GetCasiWithPosPartner(),
	}

	selectInfoTrace[fs.id] = append(selectInfoTrace[fs.id], fs)
	numberSelects++
}

// Sort the list of occurrences of each select by the time value
func sortSelects() {
	for key := range selectInfoTrace {
		sort.Slice(selectInfoTrace[key], func(i, j int) bool {
			return selectInfoTrace[key][i].t < selectInfoTrace[key][j].t
		})
	}
}

// Get a copy of fs with a randomly selected case id.
//
// Parameter:
//   - def bool: if true, default is a possible value, if false it is not
//   - flipChange bool: probability that a select case is chosen randomly. Otherwise the chosen case is kept
//
// Returns:
//   - int: the chosen case ID
func (fs fuzzingSelect) getCopyRandom(def bool, flipChance float64) fuzzingSelect {
	// do only flip with certain chance
	if rand.Float64() > flipChance {
		return fuzzingSelect{id: fs.id, t: fs.t, chosenCase: fs.chosenCase, numberCases: fs.numberCases, containsDefault: fs.containsDefault}
	}

	// if at most one case and no default (should not happen), or only default select the same case again
	if (!def && fs.numberCases <= 1) || (def && fs.numberCases == 0) {
		return fuzzingSelect{id: fs.id, t: fs.t, chosenCase: fs.chosenCase, numberCases: fs.numberCases, containsDefault: fs.containsDefault}
	}

	prefCase := fs.chooseRandomCase(def)

	return fuzzingSelect{id: fs.id, t: fs.t, chosenCase: prefCase, numberCases: fs.numberCases, containsDefault: fs.containsDefault}
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
func (fs fuzzingSelect) chooseRandomCase(def bool) int {
	// Determine the starting number based on includeZero
	start := 0
	if def {
		start = -1
	}

	// Create a weight map for the probabilities
	weights := make(map[int]int)

	// Assign weights to each number
	for i := start; i < fs.numberCases; i++ {
		weights[i] = 1 // Default weight
	}

	// Increase weights for numbers in fs.casiWithPos
	for _, num := range fs.casiWithPos {
		if num >= start && num < fs.numberCases && num != fs.chosenCase {
			weights[num] *= factorCaseWithPartner
		}
	}

	// Generate a cumulative weight array
	cumulativeWeights := []int{}
	numbers := []int{} // Keep track of the corresponding numbers
	totalWeight := 0

	for i := start; i < fs.numberCases; i++ {
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
