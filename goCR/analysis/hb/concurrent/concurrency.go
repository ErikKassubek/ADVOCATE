//
// File: concurrency.go
// Brief: Functions to find concurrent elements
//
// Created: 2025-07-03
//
// License: BSD-3-Clause

package concurrent

import (
	"goCR/analysis/data"
	"goCR/analysis/hb/clock"
	"goCR/analysis/hb/vc"
	"goCR/trace"
	"goCR/utils/log"
)

// GetConcurrent returns all concurrent elements for an element
//
// Parameters:
//   - elem trace.Element: the element to find the concurrent elements for
//   - all bool: if true, return all concurrent elements, otherwise only the first
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - sameType bool: only count values on the same type (no effect if same element is true)
//   - weak bool: get based on weak happens before
//
// Returns:
//   - []trace.Element: concurrent elements to elem
//
// testGetHB: VC: 1m12.514176153s/12.104044486s; OG: 1m59.067111952s/3.901911481s; ST1: 395.051728ms/861.348032ms
// testGetConcurrent: VC: 9m7.457251619s/50.972716941s; OG: 982.252538ms/672.449538ms; ST1: 1.084918573s/1.430036096s, ST2: 14m19.141553283s/4.613264804s
func GetConcurrent(elem trace.Element, all, sameElem, sameType, weak bool) []trace.Element {
	// if elem.GetNumberConcurrent(weak, sameElem) != -1 {
	// 	return elem.GetConcurrent(weak, sameElem)
	// }

	b := vc.GetConcurrent(elem, all, sameElem, sameType, weak)
	// b := pog.GetConcurrent(elem, all, sameElem, weak)
	// b := cssts.GetConcurrentAllPairs(elem, all, sameElem, weak)
	// b := cssts.GetConcurrent(elem, all, sameElem, weak)

	elem.SetNumberConcurrent(len(b), weak, sameElem)

	// if all concurrent are selected, filter out sameElement concurrent and set
	// as well
	if !sameElem && !sameType {
		res := make([]trace.Element, 0)

		for _, e := range b {
			switch a := e.(type) {
			case *trace.ElementSelect:
				for _, c := range a.GetCases() {
					if elem.GetID() == c.GetID() {
						res = append(res, e)
					}
				}
			default:
				if e.GetID() == elem.GetID() {
					res = append(res, e)
				}
			}

		}
		elem.SetNumberConcurrent(len(res), weak, true)
	}
	return b
}

// GetNumberConcurrent returns the number of elements that are concurrent to the elem
//
// Parameters:
//   - elem trace.Element
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - sameType bool: only count values on the same type (no effect if same element is true)
//   - weak bool: get based on weak happens before
//
// Returns:
//   - int: number of elements that are concurrent to the element
func GetNumberConcurrent(elem trace.Element, sameElem, sameType, weak bool) int {
	m := elem.GetNumberConcurrent(weak, sameElem)
	if m != -1 {
		return m
	}

	n := GetConcurrent(elem, true, sameElem, sameType, weak)
	return len(n)
}

// IsConcurrent returns if two elements are concurrent.
// The function assumes, that the vcs have been calculated
//
// Parameter:
//   - elem1: trace.Element: the first element
//   - elem2: trace.Element: the second element
//
// Returns:
//   - bool: true if the elements are concurrent, false otherwise
func IsConcurrent(elem1, elem2 trace.Element) bool {
	if !data.HBWasCalc() {
		log.Error("Cannot check for concurrency: VCs have not been calculated")
		return false
	}

	return clock.IsConcurrent(elem1.GetVC(), elem2.GetVC())
}
