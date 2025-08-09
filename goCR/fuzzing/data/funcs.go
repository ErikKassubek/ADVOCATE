//
// File: funcs.go
// Brief: Function on data
//
// Created: 2025-07-03
//
// License: BSD-3-Clause

package data

import "goCR/trace"

// AddMutToQueue adds a mutation to the queue. If a maximum number of mutation runs in set,
// only add the mutation if it does not exceed this max number
//
// Parameter:
//   - mut mutation: the mutation to add
func AddMutToQueue(mut Mutation) {
	if MaxNumberRuns == -1 || NumberFuzzingRuns+len(MutationQueue) <= MaxNumberRuns {
		MutationQueue = append(MutationQueue, mut)
	}
}

// CanBeAddedToChain decides if an element can be added to a scheduling chain
// For GoPie without improvements (!useHBInfoFuzzing) those are only mutex and channel (incl. select)
// With improvements those are all not ignored fuzzing elements
//
// Parameter:
//   - elem analysis.TraceElement: Element to check
//
// Returns:
//   - true if it can be added to a scheduling chain, false otherwise
func CanBeAddedToChain(elem trace.Element) bool {
	t := elem.GetObjType(false)
	if FuzzingMode == GoPie {
		// for standard GoPie, only mutex, channel and select operations are considered
		return t == trace.ObjectTypeMutex || t == trace.ObjectTypeChannel || t == trace.ObjectTypeSelect
	}

	return t != trace.ObjectTypeAtomic && !IgnoreFuzzing(elem, true)
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
	t := elem.GetObjType(false)
	return (ignoreNew && t == trace.ObjectTypeNew) || t == trace.ObjectTypeReplay || t == trace.ObjectTypeRoutineEnd
}
