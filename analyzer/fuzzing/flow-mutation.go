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

import "analyzer/analysis"

// if true, a new mutation run is created for each flow mutations,
// if false, all flow mutations are collected into one mutations run
const oneMutPerDelay = true

func createMutationsFlow() int {
	numberMutAdded := 0

	delay := make([](*[]analysis.ConcurrentEntry), 4)

	// once, mutex, send, recv
	delay[0], delay[1], delay[2], delay[3] = analysis.GetConcurrentInfoForFuzzing()

	// add mutations
	mutFlow := make(map[string]int)
	for i := 0; i < 4; i++ {
		for _, on := range *delay[i] {
			// limit number of mutations created by this
			if numberMutAdded > maxFlowMut {
				return numberMutAdded
			}

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

			if oneMutPerDelay {
				mutFlow = make(map[string]int)
			}
			mutFlow[pos] = on.Counter

			if _, ok := alreadyDelayedElems[pos]; !ok {
				alreadyDelayedElems[pos] = make([]int, 0)
			}
			alreadyDelayedElems[pos] = append(alreadyDelayedElems[pos], on.Counter)

			// if one mut per change, comment this in
			if oneMutPerDelay {
				mut := mutation{mutSel: selectInfoTrace, mutFlow: mutFlow}
				mutationQueue = append(mutationQueue, mut)
				numberMutAdded++
			}
		}
	}

	if oneMutPerDelay && len(mutFlow) != 0 {
		mut := mutation{mutSel: selectInfoTrace, mutFlow: mutFlow}
		mutationQueue = append(mutationQueue, mut)
		numberMutAdded++
	}

	return numberMutAdded
}
