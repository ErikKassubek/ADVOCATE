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
	"advocate/analysis/concurrent/pog"
	"advocate/trace"
)

// GetConcurrent returns all concurrent elements for an element
//
// Parameters:
//   - elem trace.Element: the element to find the concurrent elements for
//   - all bool: if true, return all concurrent elements, otherwise only the first
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: get based on weak happens before
//
// Returns:
//   - []trace.Element: concurrent elements to elem
//
// 9m11.105803818s/48.577234333s; OG: 910.626721ms/644.247894ms; ST1: 1.075209986s/1.770912088s, ST2: 13m57.841537023s/3.984323305s
func GetConcurrent(elem trace.Element, all, sameElem, weak bool) []trace.Element {
	// b := vc.GetConcurrent(elem, all, sameElem, weak)
	g := pog.GetConcurrent(elem, all, sameElem, weak)
	// c1 := cssts.GetConcurrentAllPairs(elem, all, sameElem, weak)
	// c2 := cssts.GetConcurrent(elem, all, sameElem, weak)

	// log.Importantf("VC: %d; OG: %d; ST1: %d, ST2: %d", len(b), len(g), len(c1), len(c2))

	// start := time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	vc.GetConcurrent(elem, true, false, false)
	// }
	// dur_b := time.Since(start)

	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	vc.GetConcurrent(elem, true, false, true)
	// }
	// dur_b_weak := time.Since(start)

	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	pog.GetConcurrent(elem, true, false, false)
	// }
	// dur_g := time.Since(start)

	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	pog.GetConcurrent(elem, true, false, true)
	// }
	// dur_g_weak := time.Since(start)

	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	cssts.GetConcurrentAllPairs(elem, true, false, false)
	// }
	// dur_c1 := time.Since(start)

	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	cssts.GetConcurrentAllPairs(elem, true, false, true)
	// }
	// dur_c1_weak := time.Since(start)

	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	cssts.GetConcurrent(elem, true, weak, false)
	// }
	// dur_c2 := time.Since(start)

	// start = time.Now()
	// for _, trace := range data.MainTrace.GetTraces() {
	// 	if len(trace) == 0 {
	// 		continue
	// 	}
	// 	elem := trace[len(trace)/2]
	// 	cssts.GetConcurrent(elem, true, weak, true)
	// }
	// dur_c2_weak := time.Since(start)

	// log.Importantf("VC: %s/%s; OG: %s/%s; ST1: %s/%s, ST2: %s/%s", dur_b.String(),
	// 	dur_b_weak.String(), dur_g.String(), dur_g_weak.String(), dur_c1.String(), dur_c1_weak.String(), dur_c2.String(), dur_c2_weak.String())

	elem.SetNumberConcurrent(len(g), weak)
	return g
}

// GetNumberConcurrent returns the number of elements that are concurrent to the elem
//
// Parameters:
//   - elem trace.Element
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: get based on weak happens before
//
// Returns:
//   - int: number of elements that are concurrent to the element
func GetNumberConcurrent(elem trace.Element, sameElem, weak bool) int {
	m := elem.GetNumberConcurrent(weak)
	if m != -1 {
		return m
	}

	m = len(GetConcurrent(elem, true, sameElem, weak))
	elem.SetNumberConcurrent(m, weak)
	return m
}
