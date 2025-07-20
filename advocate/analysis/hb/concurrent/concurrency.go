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
	"advocate/analysis/data"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/cssts"
	"advocate/analysis/hb/pog"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/log"
	"time"
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
// testGetHB: VC: 1m12.514176153s/12.104044486s; OG: 1m59.067111952s/3.901911481s; ST1: 395.051728ms/861.348032ms
// testGetConcurrent: VC: 9m7.457251619s/50.972716941s; OG: 982.252538ms/672.449538ms; ST1: 1.084918573s/1.430036096s, ST2: 14m19.141553283s/4.613264804s
func GetConcurrent(elem trace.Element, all, sameElem, weak bool) []trace.Element {
	// b := vc.GetConcurrent(elem, all, sameElem, weak)
	g := pog.GetConcurrent(elem, all, sameElem, weak)
	// c1 := cssts.GetConcurrentAllPairs(elem, all, sameElem, weak)
	// c2 := cssts.GetConcurrent(elem, all, sameElem, weak)

	testGetHB(elem, all)
	// testGetConcurrent(elem, all, sameElem, weak)

	elem.SetNumberConcurrent(len(g), weak)
	return g
}

func testGetHB(elem trace.Element, all bool) {

	start := time.Now()
	for _, t1 := range data.MainTrace.GetTraces() {
		if len(t1) == 0 {
			continue
		}
		e1 := t1[len(t1)/2]
		for _, t2 := range data.MainTrace.GetTraces() {
			if len(t2) == 0 {
				continue
			}
			e2 := t2[len(t2)/2]
			clock.GetHappensBefore(e1.GetVC(), e2.GetVC())
		}
	}
	dur_b := time.Since(start)

	start = time.Now()
	for _, t1 := range data.MainTrace.GetTraces() {
		if len(t1) == 0 {
			continue
		}
		e1 := t1[len(t1)/2]
		for _, t2 := range data.MainTrace.GetTraces() {
			if len(t2) == 0 {
				continue
			}
			e2 := t2[len(t2)/2]
			clock.GetHappensBefore(e1.GetWVc(), e2.GetWVc())
		}
	}
	dur_b_weak := time.Since(start)

	start = time.Now()
	for _, t1 := range data.MainTrace.GetTraces() {
		if len(t1) == 0 {
			continue
		}
		e1 := t1[len(t1)/2]
		for _, t2 := range data.MainTrace.GetTraces() {
			if len(t2) == 0 {
				continue
			}
			e2 := t2[len(t2)/2]
			pog.GetHappensBefore(e1, e2, false)
		}
	}
	dur_g := time.Since(start)

	start = time.Now()
	for _, t1 := range data.MainTrace.GetTraces() {
		if len(t1) == 0 {
			continue
		}
		e1 := t1[len(t1)/2]
		for _, t2 := range data.MainTrace.GetTraces() {
			if len(t2) == 0 {
				continue
			}
			e2 := t2[len(t2)/2]
			pog.GetHappensBefore(e1, e2, true)
		}
	}
	dur_g_weak := time.Since(start)

	start = time.Now()
	for _, t1 := range data.MainTrace.GetTraces() {
		if len(t1) == 0 {
			continue
		}
		e1 := t1[len(t1)/2]
		for _, t2 := range data.MainTrace.GetTraces() {
			if len(t2) == 0 {
				continue
			}
			e2 := t2[len(t2)/2]
			cssts.GetHappensBefore(e1, e2, false)
		}
	}
	dur_c1 := time.Since(start)
	for _, t1 := range data.MainTrace.GetTraces() {
		if len(t1) == 0 {
			continue
		}
		e1 := t1[len(t1)/2]
		for _, t2 := range data.MainTrace.GetTraces() {
			if len(t2) == 0 {
				continue
			}
			e2 := t2[len(t2)/2]
			cssts.GetHappensBefore(e1, e2, true)
		}
	}
	dur_c1_weak := time.Since(start)

	log.Importantf("VC: %s/%s; OG: %s/%s; ST1: %s/%s", dur_b.String(),
		dur_b_weak.String(), dur_g.String(), dur_g_weak.String(), dur_c1.String(), dur_c1_weak.String())
}

func testGetConcurrent(elem trace.Element, all, sameElem, weak bool) {
	// b := vc.GetConcurrent(elem, all, sameElem, weak)
	// g := pog.GetConcurrent(elem, all, sameElem, weak)
	// c1 := cssts.GetConcurrentAllPairs(elem, all, sameElem, weak)
	// c2 := cssts.GetConcurrent(elem, all, sameElem, weak)

	// log.Importantf("VC: %d; OG: %d; ST1: %d, ST2: %d", len(b), len(g), len(c1), len(c2))

	start := time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		vc.GetConcurrent(elem, true, false, false)
	}
	dur_b := time.Since(start)

	start = time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		vc.GetConcurrent(elem, true, false, true)
	}
	dur_b_weak := time.Since(start)

	start = time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		pog.GetConcurrent(elem, true, false, false)
	}
	dur_g := time.Since(start)

	start = time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		pog.GetConcurrent(elem, true, false, true)
	}
	dur_g_weak := time.Since(start)

	start = time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		cssts.GetConcurrentAllPairs(elem, true, false, false)
	}
	dur_c1 := time.Since(start)

	start = time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		cssts.GetConcurrentAllPairs(elem, true, false, true)
	}
	dur_c1_weak := time.Since(start)

	start = time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		cssts.GetConcurrent(elem, true, false, false)
	}
	dur_c2 := time.Since(start)

	start = time.Now()
	for _, trace := range data.MainTrace.GetTraces() {
		if len(trace) == 0 {
			continue
		}
		elem := trace[len(trace)/2]
		cssts.GetConcurrent(elem, true, false, true)
	}
	dur_c2_weak := time.Since(start)

	log.Importantf("VC: %s/%s; OG: %s/%s; ST1: %s/%s, ST2: %s/%s", dur_b.String(),
		dur_b_weak.String(), dur_g.String(), dur_g_weak.String(), dur_c1.String(), dur_c1_weak.String(), dur_c2.String(), dur_c2_weak.String())
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
