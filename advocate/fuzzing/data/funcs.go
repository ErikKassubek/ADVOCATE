// Copyright (c) 2025 Erik Kassubek
//
// File: funcs.go
// Brief: Function on data
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package data

// Add a mutation to the queue. If a maximum number of mutation runs in set,
// only add the mutation if it does not exceed this max number
//
// Parameter:
//   - mut mutation: the mutation to add
func AddMutToQueue(mut Mutation) {
	if MaxNumberRuns == -1 || NumberFuzzingRuns+len(MutationQueue) <= MaxNumberRuns {
		MutationQueue = append(MutationQueue, mut)
	}
}
