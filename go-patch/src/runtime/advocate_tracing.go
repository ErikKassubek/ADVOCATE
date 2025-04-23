// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace.go
// Brief: Functionality for tracing
//
// Author: Erik Kassubek
// Created: 2024-04-10
//
// License: BSD-3-Clause

package runtime

var finishTracingFunc func()

var tracingStartNano int64

// InitTracing enables the collection of the trace
//
// Parameter:
//   - finishFuzzing func(): function injection for the advocate.FinishFuzzing function
func InitTracing(finishFuzzing func()) {
	advocateTracingDisabled = false
	finishTracingFunc = finishFuzzing

	if tracingStartNano == 0 {
		tracingStartNano = nanotime()
	}
}

// DisableTracing disables the trace recording
func DisableTracing() {
	advocateTracingDisabled = true
}

// IsTracingEnabled returns whether tracing is enabled
//
// Returns:
//   - true if enabled, false otherwise
func IsTracingEnabled() bool {
	return !advocateTracingDisabled
}
