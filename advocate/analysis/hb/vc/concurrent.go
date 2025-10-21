// Copyright (c) 2025 Erik Kassubek
//
// File: vc.go
// Brief: Data required for calculating the vector clocks
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package vc

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/helper"
	"advocate/trace"
	"advocate/utils/log"
)

// Current vector clocks
var (
	// current happens before vector clocks
	CurrentVC = make(map[int]*clock.VectorClock)

	// current must happens before vector clocks
	CurrentWVC = make(map[int]*clock.VectorClock)

	// vector clock for each buffer place in vector clock
	// the map key is the channel id. The slice is used for the buffer positions
	chanBuffer = make(map[int]([]baseA.BufferedVC))
	// the current buffer position
	chanBufferSize = make(map[int]int)
)

// InitVC initializes the current vector clocks
func InitVC() {
	chanBuffer = make(map[int][]baseA.BufferedVC)
	chanBufferSize = make(map[int]int)
	CurrentVC = make(map[int]*clock.VectorClock)
	CurrentWVC = make(map[int]*clock.VectorClock)

	noRoutine := baseA.MainTrace.GetNoRoutines()
	for i := 1; i <= noRoutine; i++ {
		CurrentVC[i] = clock.NewVectorClock(noRoutine)
		CurrentWVC[i] = clock.NewVectorClock(noRoutine)
	}
}

// GetConcurrent find a/all element(s) that are concurrent to a given element
// This function assumes that the vector clocks have already been calculated
// The function iterates over all elements, and compares the vector clocks
//
// Parameter:
//   - elem trace.TraceElement: the element to search for
//   - all bool: if true, find all concurrent elements, if false, find only one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - sameType bool: only count values on the same type (no effect if same element is true)
//   - weak bool: use the weak happens before relation
//
// Returns:
//   - []trace.Element: set of elements concurrent to elem
func GetConcurrent(elem trace.Element, all, sameElem, sameType, weak bool) []trace.Element {
	if !baseA.HBWasCalc() {
		log.Error("Cannot find concurrent elements: VCs have not been calculated")
		return make([]trace.Element, 0)
	}

	res := make([]trace.Element, 0)
	for rout, tr := range baseA.MainTrace.GetTraces() {
		if rout == elem.GetRoutine() {
			continue
		}

		for _, tElem := range tr {
			if tElem.GetTPost() == 0 {
				continue
			}

			if sameElem && elem.GetID() != tElem.GetID() {
				continue
			}

			elemType := elem.GetType(false)
			tElemType := tElem.GetType(false)

			if sameType && elemType != tElemType &&
				!((elemType == trace.Select && tElemType == trace.Channel) ||
					(elemType == trace.Channel && tElemType == trace.Select)) {
				continue
			}

			if !helper.Valid(tElem) {
				continue
			}

			if weak {
				if clock.IsConcurrent(elem.GetWVC(), tElem.GetWVC()) {
					res = append(res, tElem)
					if !all {
						return res
					}
				}
			} else {
				if clock.IsConcurrent(elem.GetVC(), tElem.GetVC()) {
					res = append(res, tElem)
					if !all {
						return res
					}
				}
			}
		}
	}

	return res
}
