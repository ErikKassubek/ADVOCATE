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
	"advocate/analysis/clock"
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
func GetConcurrentBruteForce(elem trace.TraceElement, all bool) []trace.TraceElement {
	if !data.HBWasCalc() {
		log.Error("Cannot find concurrent elements: VCs have not been calculated")
		return make([]trace.TraceElement, 0)
	}

	res := make([]trace.TraceElement, 0)
	for rout, trace := range data.MainTrace.GetTraces() {
		if rout == elem.GetRoutine() {
			continue
		}

		for _, tElem := range trace {
			if clock.GetHappensBefore(elem.GetWVc(), tElem.GetWVc()) == clock.Concurrent {
				res = append(res, tElem)
				if !all {
					return res
				}
			}
		}
	}

	return res
}
