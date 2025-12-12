// Copyright (c) 2025 Erik Kassubek
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
	for _, t := range processedTraces[origID] {
		if areEquivalent(&t1, &t) {
			return true
		}
	}

	processedTraces[origID] = append(processedTraces[origID], t1)

	return false
}

// AddOrig adds an actually executed trace to processed traces. Must be run
// before running HasEquivalence with the given id
func AddOrig(t TraceEq, id int) {
	t.BuildCanonicalSignature()

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
	if t1.signature == "" {
		t1.BuildCanonicalSignature()
	}

	if t2.signature == "" {
		t1.BuildCanonicalSignature()
	}

	return t1.signature == t2.signature
}
