// Copyright (c) 2025 Erik Kassubek
//
// File: independentTraces.go
// Brief: Some traces are independent, meaning if one trace does not
//    contain a concurrency bug (panic or deadlock), the other cannot
//    contain one either. This file contains functions to check if two operations
//    are independent.
//
// Author: Erik Kassubek
// Created: 2025-10-13
//
// License: BSD-3-Clause

package guided

import (
	"advocate/trace"
)

// IndependentTracesMin checks if the two given min traces are independent
//
// Parameter:
//   - t1: trace.TraceMin: first trace
//   - t2: trace.TraceMin: second trace
//
// Returns:
//   - bool: true if they are independent (we only need to run one of them), false otherwise
func independentTracesMin(t1, t2 trace.TraceMin) bool {
	var shorter, longer trace.TraceMin

	if t1.Len() < t2.Len() {
		shorter = t1
		longer = t2
	} else {
		shorter = t2
		longer = t1
	}

	for i := 0; i < longer.Len()-shorter.Len(); i++ {
		sub := longer.CloneSub(i, i+shorter.Len())
		if reachable(sub, shorter, make(map[string]bool)) {
			return true
		}
	}

	return true
}

func reachable(curr, target trace.TraceMin, memo map[string]bool) bool {
	if curr.IsEqual(&target) {
		return true
	}

	key := curr.Key()

	if v, ok := memo[key]; ok {
		return v
	}

	for i := 0; i < curr.Len()-1; i++ {
		if ok, _ := areIndependent(curr.Get(i), curr.Get(i+1)); ok {
			next := curr.Clone()
			next.Flip(i, i+1)

			if reachable(next, target, memo) {
				memo[key] = true
				return true
			}
		}
	}

	memo[key] = false
	return false
}

// independentTraceMin checks if there is an independent trace in processedTraces
//
// Parameter:
//   - t1 trace.TraceMin: the trace to test
//
// Returns:
//   - true if there is an independent trace (we do not need to rerun t1), false otherwise
func independentTraceMin(t1 trace.TraceMin) bool {
	for _, t := range processedTraces {
		if independentTracesMin(t1, t) {
			return true
		}
	}

	return false
}
