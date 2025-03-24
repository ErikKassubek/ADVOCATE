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

import (
	"fmt"
	"sort"
)

func createMutationsGFuzz(numberMutations int, flipChance float64) int {
	numberMutAdded := 0

	for i := 0; i < numberMutations; i++ {
		mut := createMutation(flipChance)

		id := getIdFromMut(mut)
		if id == "" {
			continue
		}

		if num, _ := allMutations[id]; num < maxRunPerMut {
			muta := mutation{mutType: mutSelType, mutSel: mut}
			mutationQueue = append(mutationQueue, muta)
			allMutations[id]++
			numberMutAdded++
		}
	}

	return numberMutAdded
}

func createMutation(flipChance float64) map[string][]fuzzingSelect {
	res := make(map[string][]fuzzingSelect)

	for key, listSel := range selectInfoTrace {
		res[key] = make([]fuzzingSelect, 0)
		for _, sel := range listSel {
			res[key] = append(res[key], sel.getCopyRandom(true, flipChance))
		}
	}

	return res
}

func popMutation() mutation {
	var mut mutation
	mut, mutationQueue = mutationQueue[0], mutationQueue[1:]
	return mut
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

/*
 * Get a unique string id for a given mutation
 * Args:
 * 	mut map[string][]fuzzingSelect: mutation
 * Returns:
 * 	string: id
 */
func getIdFromMut(mut map[string][]fuzzingSelect) string {
	keys := make([]string, 0, len(mut))
	for key := range mut {
		keys = append(keys, key)
	}

	// Sort the keys alphabetically
	sort.Strings(keys)

	id := ""

	// Iterate over the sorted keys
	for _, key := range keys {
		id := key + "-"
		for _, sel := range mut[key] {
			id += fmt.Sprintf("%d", sel.chosenCase)
		}
	}

	return id
}
