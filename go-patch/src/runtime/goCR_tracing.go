// GOCP-FILE_START

// File: goCR_trace.go
// Brief: Functionality for tracing
//
// Created: 2024-04-10
//
// License: BSD-3-Clause

package runtime

var finishTracingFunc func()

var tracingStartNano int64

// InitTracing enables the collection of the trace
//
// Parameter:
//   - finishFuzzing func(): function injection for the gocr.FinishFuzzing function
func InitTracing(finishFuzzing func()) {
	goCRTracingDisabled = false
	finishTracingFunc = finishFuzzing
	setCurrentRoutineToActive()

	if tracingStartNano == 0 {
		tracingStartNano = nanotime()
	}
}

// DisableTracing disables the trace recording
func DisableTracing() {
	goCRTracingDisabled = true
}

// IsTracingEnabled returns whether tracing is enabled
//
// Returns:
//   - true if enabled, false otherwise
func IsTracingEnabled() bool {
	return !goCRTracingDisabled
}
