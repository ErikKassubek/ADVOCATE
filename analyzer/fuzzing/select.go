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
	"fmt"
	"sort"

	"math/rand"
)

var (
	allSelects    = make(map[string][]fuzzingSelect) // id -> []fuzzingSelects
	numberSelects = 0
)

/*
 * Struct to handle the selects for fuzzing
 *   t: tpost of the select execution, used for order
 *   chosenCase (int): id of the chosen case, -1 for default
 *   numberCases (int): number of cases not including default
 *   containsDefault (bool): true if contains default case, otherwise false
 */
type fuzzingSelect struct {
	id              string
	t               int
	chosenCase      int
	numberCases     int
	containsDefault bool
}

func addFuzzingSelect(id string, t int, chosenCase int, numberCases int, containsDefault bool) {
	fs := fuzzingSelect{
		id:              id,
		t:               t,
		chosenCase:      chosenCase,
		numberCases:     numberCases,
		containsDefault: containsDefault,
	}

	allSelects[id] = append(allSelects[id], fs)
	numberSelects++
}

func sortSelects() {
	for key := range allSelects {
		sort.Slice(allSelects[key], func(i, j int) bool {
			return allSelects[key][i].t < allSelects[key][j].t
		})
	}
}

func (fs fuzzingSelect) toString() string {
	return fmt.Sprintf("%s;%d;%d", fs.id, fs.chosenCase, fs.numberCases)
}

/*
 * Get a copy of fs with a randomly selected case id
 * Args:
 *   def (bool): if true, default is a possible value, if false it is not
 *   flipChange (bool): probability that a select case is chosen randomly. Otherwise the chosen case is kept
 * Return:
 *   (int): the chosen case ID
 */
func (fs fuzzingSelect) getCopyRandom(def bool, flipChance float64) fuzzingSelect {
	// do only flip with certain chance
	if rand.Float64() > flipChance {
		return fuzzingSelect{id: fs.id, t: fs.t, chosenCase: fs.chosenCase, numberCases: fs.numberCases, containsDefault: fs.containsDefault}
	}

	// if at most one case and no default (should not happen), or only default select the same case again
	if (!def && fs.numberCases <= 1) || (def && fs.numberCases == 0) {
		return fuzzingSelect{id: fs.id, t: fs.t, chosenCase: fs.chosenCase, numberCases: fs.numberCases, containsDefault: fs.containsDefault}
	}

	// if def == false -> rand between 0 and fs.numberCases - 1
	// otherwise rand between -1 and fs.numberCases - 1
	start := 0
	if def && fs.containsDefault {
		start = -1
	}

	chosenCase := rand.Intn(fs.numberCases-start-1) + start

	// do not select the same number again
	if chosenCase >= fs.chosenCase {
		chosenCase++
	}

	return fuzzingSelect{id: fs.id, t: fs.t, chosenCase: chosenCase, numberCases: fs.numberCases, containsDefault: fs.containsDefault}
}

func (fs fuzzingSelect) isEqual(fs2 fuzzingSelect) bool {
	return fs.id == fs2.id && fs.chosenCase == fs2.chosenCase && fs.numberCases == fs2.numberCases && fs.containsDefault == fs2.containsDefault
}
