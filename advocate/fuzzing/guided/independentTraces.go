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

import "advocate/trace"

// IndependentTracesMin checks if the two given min traces are independent
//
// Parameter:
//   - t1: trace.TraceMin: first trace
//   - t2: trace.TraceMin: second trace
//
// Returns:
//   - bool: true if they are independent (we only need to run one of them), false otherwise
func independentTracesMin(t1, t2 trace.TraceMin) bool {
	// TODO: implement
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
