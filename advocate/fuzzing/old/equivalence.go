// Copyright (c) 2026 Erik Kassubek
//
// File: equivalence.go
// Brief: Functions to check if two mutations/traces are equivalent for the
//   purpose of fuzzing
//
// Author: Erik Kassubek
// Created: 2025-12-04
//
// License: BSD-3-Clause

package equivalence

var processedTraces = make(map[int][]TraceEq)

// HasEquivalent checks if there is an independent trace in processedTraces
//
// Parameter:
//   - t1 trace.TraceMin: the trace to test
//   - origID int: id of actually run execution t1 is based on
//
// Returns:
//   - true if there is an independent trace (we do not need to rerun t1), false otherwise
func HasEquivalent(t1 TraceEq, origID int) bool {
	if t1.illFormedBug {
		return false
	}

	for _, t := range processedTraces[origID] {
		if areEquivalent(&t, &t1) {
			return true
		}
	}

	processedTraces[origID] = append(processedTraces[origID], t1)

	return false
}

// AddOrig adds an actually executed trace to processed traces. Must be run
// before running HasEquivalence with the given id
func AddOrig(t TraceEq, id int) {
	processedTraces[id] = make([]TraceEq, 0)
	processedTraces[id] = append(processedTraces[id], t)
}

// areEquivalent takes two traces and determines, if they
// are equivalent.
//
// Parameter:
//   - t1: TraceEq: trace1 1
//   - t2: TraceEq: trace2 2
//
// Returns:
//   - bool: true if the traces are equivalent
func areEquivalent(t1, t2 *TraceEq) bool {
	// t1 should be longer
	if len(t1.trace) < len(t2.trace) {
		t1, t2 = t2, t1
	}

	// Build a set of IDs for t1
	ids1 := make(map[int]bool)
	for _, e := range t1.trace {
		ids1[e.GetID()] = true
	}

	// Check that every element of t2 exists in t1
	shared := make(map[int]bool)
	for _, e := range t2.trace {
		id := e.GetID()
		if !ids1[id] {
			// Found an element in t2 not in t1 -> t2 is not subset
			return false
		}
		shared[id] = true
	}

	signature1 := t1.BuildCanonicalSignature(shared, true)
	signature2 := t2.BuildCanonicalSignature(shared, len(t1.trace) == len(t2.trace))

	return signature1 == signature2
}
