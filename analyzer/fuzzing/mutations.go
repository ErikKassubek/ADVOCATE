// Copyright (c) 2024 Erik Kassubek
//
// File: mutations.go
// Brief: Create the mutations
//
// Author: Erik Kassubek
// Created: 2024-12-03
//
// License: BSD-3-Clause

package fuzzing

func createMutations(numberMutations int, flipChance float64) {
	for i := 0; i < numberMutations; i++ {
		mut := getMutation(flipChance)

		if isNewMutation(mut) {
			mutationQueue = append(mutationQueue, mut)
			allMutations = append(allMutations, mut)
		}
	}
}

func getMutation(flipChance float64) map[string][]fuzzingSelect {
	res := make(map[string][]fuzzingSelect)

	for key, listSel := range allSelects {
		res[key] = make([]fuzzingSelect, 0)
		for _, sel := range listSel {
			res[key] = append(res[key], sel.getCopyRandom(false, flipChance))
		}
	}

	return res
}

func isNewMutation(mut map[string][]fuzzingSelect) bool {
	for _, mut2 := range allMutations {
		if areMutEqual(mut, mut2) {
			return false
		}
	}
	return true
}

func areMutEqual(mut1, mut2 map[string][]fuzzingSelect) bool {
	// different amount of keys
	if len(mut1) != len(mut2) {
		return false
	}

	for key, slice1 := range mut1 {
		slice2, exists := mut2[key]
		// key in mut1 is not in mut2
		if !exists {
			return false
		}

		// slice1 and slice 2 are not identical, order must be the same
		if len(slice1) != len(slice2) {
			return false
		}

		for index, sel := range slice1 {
			if !sel.isEqual(slice2[index]) {
				return false
			}
		}
	}

	return true
}
