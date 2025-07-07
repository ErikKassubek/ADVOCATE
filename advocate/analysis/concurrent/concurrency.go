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
	"advocate/trace"
)

// GetConcurrent returns all concurrent elements for an element
//
// Parameters:
//   - elem trace.Element: the element to find the concurrent elements for
//   - all bool: if true, return all concurrent elements, otherwise only the first
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//
// Returns:
//   - []trace.Element: concurrent elements to elem
//
// VC: 44.747463124s; OG: 619.517929ms; ST1: 1.26254268s, ST2: 4.226630035s
func GetConcurrent(elem trace.Element, all, sameElem bool) []trace.Element {
	// b := GetConcurrentVC(elem, all, sameElem)
	g := GetConcurrentPartialOrderGraph(elem, all, sameElem)
	// c1 := cssts.GetConcurrentCSSTAllPairs(elem, all, sameElem)
	// c2 := cssts.GetConcurrentCSST(elem, all, sameElem)

	// log.Importantf("VC: %d; OG: %d; ST1: %d, ST2: %d", len(b), len(g), len(c1), len(c2))
	// // elem.SetNumberConcurrent(len(e))

	// log.Important("S1")
	// start := time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	GetConcurrentVC(elem, true, false)
	// }
	// dur_b := time.Since(start)

	// log.Important("S2")
	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	GetConcurrentPartialOrderGraph(elem, true, false)
	// }
	// dur_g := time.Since(start)

	// log.Important("S3")
	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	cssts.GetConcurrentCSSTAllPairs(elem, true, false)
	// }
	// dur_c1 := time.Since(start)

	// log.Important("S4")
	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	cssts.GetConcurrentCSST(elem, true, false)
	// }
	// dur_c2 := time.Since(start)
	// log.Importantf("VC: %s; OG: %s; ST1: %s, ST2: %s", dur_b.String(), dur_g.String(), dur_c1.String(), dur_c2.String())

	// return make([]trace.Element, 0)
	return g
}

// GetNumberConcurrent returns the number of elements that are concurrent to the elem
//
// Parameters:
//   - elem trace.Element
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//
// Returns:
//   - int: number of elements that are concurrent to the element
func GetNumberConcurrent(elem trace.Element, sameElem bool) int {
	m := elem.GetNumberConcurrent()
	if m != -1 {
		return m
	}

	m = len(GetConcurrent(elem, true, sameElem))
	elem.SetNumberConcurrent(m)
	return m
}

// Function to filter out element which do not correspond to valid operations, e.g.
// end of a routine
//
// Parameter:
//   - elem trace.Element: the element to test
//
// Returns:
//   - bool: true if the element is valid, false otherwise
func valid(elem trace.Element) bool {
	t := elem.GetObjType(false)
	return !(t == trace.ObjectTypeReplay || t == trace.ObjectTypeNew || t == trace.ObjectTypeRoutineEnd)
}
