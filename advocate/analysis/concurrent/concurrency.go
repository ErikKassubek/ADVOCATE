// Copyright (c) 2025 Erik Kassubek
//
// File: concurrency.go
// Brief: Functions to find concurrent elements
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package concurrent

import (
	"advocate/analysis/concurrent/cssts"
	"advocate/trace"
)

// GetConcurrent returns all concurrent elements for an element
//
// Parameters:
//   - elem trace.Element: the element to find the concurrent elements for
//
// Returns:
//   - []trace.Element: concurrent elements to elem
func GetConcurrent(elem trace.Element) []trace.Element {
	// elem := GetConcurrentBruteForce(elem, true)
	// elem := GetConcurrentPartialOrderGraph(elem, true)
	concurrent := cssts.GetConcurrentCSST(elem, true)
	elem.SetNumberConcurrent(len(concurrent))
	return concurrent
}

// GetNumberConcurrent returns the number of elements that are concurrent to the elem
//
// Parameters:
//   - elem trace.Element
//
// Returns:
//   - int: number of elements that are concurrent to the element
func GetNumberConcurrent(elem trace.Element) int {
	m := elem.GetNumberConcurrent()
	if m != -1 {
		return m
	}

	m = len(GetConcurrent(elem))
	elem.SetNumberConcurrent(m)
	return m
}
