// Copyright (c) 2025 Erik Kassubek
//
// File: flow-mutation.go
// Brief: Add mutations based on flow
//
// Author: Erik Kassubek
// Created: 2025-02-24
//
// License: BSD-3-Clause

package fuzzing

import (
	"analyzer/analysis"
)

var alreadyDelayedElems = make(map[string][]int)

// TODO: add channel mutations
func createMutationsFlow() int {
	numberMutAdded := 0

	elemsToDelay := analysis.GetConcurrentInfoForFuzzing()

	// add mutations
	for _, on := range *elemsToDelay {
		pos := on.Elem.GetPos()
		if counts, ok := alreadyDelayedElems[pos]; ok {
			found := false
			for _, count := range counts {
				if count == on.Counter {
					found = true
					break
				}
			}
			if found {
				continue
			}
		}

		mutFlow := make(map[string]int)
		mutFlow[pos] = on.Counter

		if _, ok := alreadyDelayedElems[pos]; !ok {
			alreadyDelayedElems[pos] = make([]int, 0)
		}
		alreadyDelayedElems[pos] = append(alreadyDelayedElems[pos], on.Counter)

		mut := mutation{mutSel: selectInfoTrace, mutFlow: mutFlow}
		mutationQueue = append(mutationQueue, mut)
		numberMutAdded++
	}
	return numberMutAdded
}
