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

import "advocate/trace"

var processedTraces = make(map[int][]trace.TraceMin, 0)

// HasEquivalent checks if there is an independent trace in processedTraces
//
// Parameter:
//   - t1 trace.TraceMin: the trace to test
//   - origID int: id of actually run execution t1 is based on
//
// Returns:
//   - true if there is an independent trace (we do not need to rerun t1), false otherwise
func HasEquivalent(t1 trace.TraceMin, origID int) bool {
	if len(t1.Props()) == 0 {
		t1.SetProps(precompute(&t1))
	}

	for _, t := range processedTraces[origID] {
		if areEquivalent(&t1, &t) {
			return true
		}
	}

	processedTraces[origID] = append(processedTraces[origID], t1)

	return false
}

// areEquivalent checks if two traces are equivalent
//
// Parameter:
//   - t1 *trace.TraceMin: trace 1
//   - t2 *trace.TraceMin: trace 2
//
// Returns:
//   - bool: true if the traces are equivalent, false otherwise
func areEquivalent(t1, t2 *trace.TraceMin) bool {
	p1 := t1.Props()
	p2 := t2.Props()

	if len(p1) != len(p2) { // TODO: is this correct?
		return false
	}

	for id, val1 := range p1 {
		val2, ok := p2[id]
		if !ok { // TODO: same as above, is this correct?
			return false
		}

		if !val1.IsEqual(val2) {
			return false
		}
	}

	return true
}

// AddOrig adds an actually executed trace to processed traces. Must be run
// before running HasEquivalence with the given id
func AddOrig(t trace.TraceMin, id int) {
	t.SetProps(precompute(&t))

	processedTraces[id] = make([]trace.TraceMin, 0)
	processedTraces[id] = append(processedTraces[id], t)
}

// IndependentTracesMin checks if the two given min traces are independent
//
// Parameter:
//   - t1: trace.TraceMin: first trace
//   - t2: trace.TraceMin: second trace
//
// Returns:
//   - bool: true if they are independent (we only need to run one of them), false otherwise
// func areEquivalent(t1, t2 trace.TraceMin) bool {
// 	var shorter, longer trace.TraceMin

// 	if t1.Len() < t2.Len() {
// 		shorter = t1
// 		longer = t2
// 	} else {
// 		shorter = t2
// 		longer = t1
// 	}

// 	for i := 0; i < longer.Len()-shorter.Len(); i++ {
// 		sub := longer.CloneSub(i, i+shorter.Len())
// 		if reachable(sub, shorter, make(map[string]bool)) {
// 			return true
// 		}
// 	}

// 	return true
// }

// func reachable(curr, target trace.TraceMin, memo map[string]bool) bool {
// 	if curr.IsEqual(&target) {
// 		return true
// 	}

// 	key := curr.Key()

// 	if v, ok := memo[key]; ok {
// 		return v
// 	}

// 	for i := 0; i < curr.Len()-1; i++ {
// 		if ok, _ := areEquivalent(curr.Get(i), curr.Get(i+1)); ok {
// 			next := curr.Clone()
// 			next.Flip(i, i+1)

// 			if reachable(next, target, memo) {
// 				memo[key] = true
// 				return true
// 			}
// 		}
// 	}

// 	memo[key] = false
// 	return false
// }
