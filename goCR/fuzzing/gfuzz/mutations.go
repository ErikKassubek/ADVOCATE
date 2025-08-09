//
// File: mutations.go
// Brief: Create the mutations
//
// Created: 2024-12-03
//
// License: BSD-3-Clause

package gfuzz

import (
	"fmt"
	"goCR/fuzzing/data"
	"sort"
)

// createMutationsGFuzz creates the new mutations for a trace based on GFuzz
//
// Parameter:
//   - numberMutation int: number of mutation to create
//   - flipChance float64: probability that for a given select the preferred case is changed
//
// Returns:
//   - int: number of added mutations
func createMutationsGFuzz(numberMutations int, flipChance float64) int {
	numberMutAdded := 0

	numberSkip := 0
	for i := 0; i < numberMutations; i++ {
		mut := createMutation(flipChance)

		id := getIDFromMut(mut)
		if id == "" {
			continue
		}

		if num, _ := data.AllMutations[id]; num < maxRunPerMut {
			mut := data.Mutation{MutType: data.MutSelType, MutSel: mut}
			data.AddMutToQueue(mut)
			data.AllMutations[id]++
			numberMutAdded++
			numberSkip = 0
		} else {
			// redraw mutation that has already been executed to often
			// break if no new valid mut has been created for a while
			numberSkip++
			if numberSkip > len(mut) {
				break
			}
			i--
		}
	}

	return numberMutAdded
}

// createMutation creates one new mutation
//
// Parameter:
//   - flipChance float64: probability that a select changes its preferred case
//
// Returns:
//   - map[string][]fuzzingSelect: the new mutation
func createMutation(flipChance float64) map[string][]data.FuzzingSelect {
	res := make(map[string][]data.FuzzingSelect)

	for key, listSel := range SelectInfoTrace {
		res[key] = make([]data.FuzzingSelect, 0)
		for _, sel := range listSel {
			res[key] = append(res[key], sel.GetCopyRandom(sel.ContainsDefault, flipChance))
		}
	}

	return res
}

// Get a unique string id for a given mutation
//
// Parameter:
//   - mut map[string][]fuzzingSelect: mutation
//
// Returns:
//   - string: id
func getIDFromMut(mut map[string][]data.FuzzingSelect) string {
	keys := make([]string, 0, len(mut))
	for key := range mut {
		keys = append(keys, key)
	}

	// Sort the keys alphabetically
	sort.Strings(keys)

	id := ""

	// Iterate over the sorted keys
	for _, key := range keys {
		id += (key + "-")
		for _, sel := range mut[key] {
			id += fmt.Sprintf("%d", sel.ChosenCase)
		}
	}

	return id
}
