// Copyright (c) 2025 Erik Kassubek
//
// File: compVC.go
// Brief: Function to find concurrent operations by directly comparing vector clocks
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package concurrent

import (
	"advocate/analysis/concurrent/clock"
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/log"
)

// For a given element, find a/all element(s) that are concurrent to it
// This function assumes that the vector clocks have already been calculated
// The function iterates over all elements, and compares the vector clocks
//
// Parameter:
//   - elem trace.TraceElement: the element to search for
//   - all bool: if true, find all concurrent elements, if false, find only one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//
// Returns:
//   - []trace.Element: set of elements concurrent to elem
func GetConcurrentVC(elem trace.Element, all, sameElem bool) []trace.Element {
	if !data.HBWasCalc() {
		log.Error("Cannot find concurrent elements: VCs have not been calculated")
		return make([]trace.Element, 0)
	}

	res := make([]trace.Element, 0)
	for rout, trace := range data.MainTrace.GetTraces() {
		if rout == elem.GetRoutine() {
			continue
		}

		for _, tElem := range trace {
			if sameElem && elem.GetID() != tElem.GetID() {
				continue
			}

			if !valid(tElem) {
				continue
			}

			if clock.IsConcurrent(elem.GetWVc(), tElem.GetWVc()) {
				res = append(res, tElem)
				if !all {
					return res
				}
			}
		}
	}

	return res
}
