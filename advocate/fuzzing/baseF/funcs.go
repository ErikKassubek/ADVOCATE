// Copyright (c) 2025 Erik Kassubek
//
// File: funcs.go
// Brief: Function on data
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package baseF

import (
	"advocate/trace"
)

// AddMutToQueue adds a mutation to the queue. If a maximum number of mutation runs in set,
// only add the mutation if it does not exceed this max number
//
// Parameter:
//   - mut mutation: the mutation to add
//   - front bool: if true, add as next mutation, otherwise as last in queue
//   - force bool: if true, the mutation is always added, otherwise only if it does not exceed the max number of runs
//
// Returns:
//   - bool: true, if the mutation was added, false otherwise
func AddMutToQueue(mut Mutation, front, force bool) bool {
	if force || MaxNumberRuns == -1 || NumberFuzzingRuns+len(MutationQueue) <= MaxNumberRuns {
		if front {
			MutationQueue = append([]Mutation{mut}, MutationQueue...)
		} else {
			MutationQueue = append(MutationQueue, mut)
		}
		return true
	}
	return false
}

// IgnoreFuzzing checks if an element should be ignored for fuzzing
// For the creation of mutations we ignore all elements that do not directly
// correspond to relevant operations. Those are new, replay, routineEnd
//
// Parameter:
//   - elem *trace.TraceElementFork: The element to check
//   - ignoreNew bool: if true, new elem is ignored elem, otherwise not
//
// Returns:
//   - True if the element is of one of those types, false otherwise
func IgnoreFuzzing(elem trace.Element, ignoreNew bool) bool {
	t := elem.GetType(false)
	return (ignoreNew && t == trace.New) || t == trace.Replay || t == trace.End
}
