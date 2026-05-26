// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_trace.go
// Brief: Functionality for tracing
//
// Author: Erik Kassubek
// Created: 2024-04-10
//
// License: BSD-3-Clause

package runtime

var finishTracingFunc func()
var writeTraceToFileFunc func(routine int, fromRuntime bool) bool

var tracingStartNano int64

// InitTracing enables the collection of the trace
//
// Parameter:
//   - finishFuzzing func(): function injection for the gocdr.FinishFuzzing function
//   - wrwriteToTraceFile func(r int, f bool) bool: function injection for writing to trace files
func InitTracing(finishFuzzing func(), writeToTraceFile func(r int, f bool) bool) {
	gocdrTracingDisabled = false
	finishTracingFunc = finishFuzzing
	writeTraceToFileFunc = writeToTraceFile
	setCurrentRoutineToActive()

	if tracingStartNano == 0 {
		tracingStartNano = nanotime()
	}
}

// DisableTracing disables the trace recording
func DisableTracing() {
	gocdrTracingDisabled = true
}

// IsTracingEnabled returns whether tracing is enabled
//
// Returns:
//   - true if enabled, false otherwise
func IsTracingEnabled() bool {
	return !gocdrTracingDisabled
}
