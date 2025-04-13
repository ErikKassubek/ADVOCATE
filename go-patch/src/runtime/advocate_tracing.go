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

/*
 * InitTracing enables the collection of the trace
 */
func InitTracing(finishFuzzing func()) {
	advocateTracingDisabled = false
	finishTracingFunc = finishFuzzing
}

/*
 * DisableTracing disables the collection of the trace
 */
func DisableTracing() {
	advocateTracingDisabled = true
}

/*
 * IsTracingEnabled returns wether tracing is enabled
 * Return:
 * 	true if enabled, false otherwise
 */
func IsTracingEnabled() bool {
	return !advocateTracingDisabled
}
