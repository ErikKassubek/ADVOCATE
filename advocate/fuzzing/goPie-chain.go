// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-sc.go
// Brief: scheduling Chains for GoPie
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package fuzzing

import (
	"advocate/analysis/clock"
	"advocate/analysis/data"
	"advocate/trace"
	"fmt"
)

var (
	schedulingChains []chain
	currentChain     chain
	lastRoutine      = -1
)

// Representation of a scheduling chain
// A chain is an ordered list of adjacent element from the trace,
// where two neighboring elements must be from different routines
type chain struct {
	elems []trace.TraceElement
}

// Create a new, empty chain
//
// Returns: chain: the new chain
func newChain() chain {
	elems := make([]trace.TraceElement, 0)
	return chain{elems}
}

// func addElemToChain(elem trace.TraceElement) {
// 	routine := elem.GetRoutine()

// 	// if the element is already in the chain, it is not added again
// 	if currentChain.contains(elem) {
// 		return
// 	}

// 	// add elem if the last routine is different from the routine of the elem
// 	// if the current routine is empty, lastRoutine is -1 and this is always true
// 	if lastRoutine != routine {
// 		currentChain.add(elem.Copy())
// 	} else {
// 		// if the routine is the same as the last routine, we need to start a new
// 		// chain. In this case, store the current chain as a scheduling chains
// 		// and start a new routine with the current element
// 		if currentChain.len() > 1 {
// 			schedulingChains = append(schedulingChains, currentChain)
// 		}
// 		currentChain = newChain()
// 		currentChain.add(elem.Copy())
// 	}

// 	lastRoutine = routine
// }

// randomChain returns a chain consisting of a
// pair of operations (only of channel, select or mutex)
// that are in a rel2 relation
//
// Returns:
//   - the chain, or an empty chain if pair exists
func randomChain() chain {
	res := newChain()

	for elem1, rel := range rel2 {
		for elem2 := range rel {
			res.add(elem1)
			res.add(elem2)
			return res
		}
	}

	return res
}

// Add a new element to the chain
//
// Parameter:
//   - elem analysis.TraceElement: Element to add
func (ch *chain) add(elem trace.TraceElement) {
	if elem == nil {
		return
	}

	ch.elems = append(ch.elems, elem)
}

// replace replaces the element at a given index in a chain with another element
//
// Parameter:
//   - index int: index to change at
//   - elem analysis.TraceElement: element to set at index
func (ch *chain) replace(index int, elem trace.TraceElement) {
	if elem == nil {
		return
	}

	if index < 0 || index >= len(ch.elems) {
		return
	}
	ch.elems[index] = elem
}

// Returns if the chain contains a specific element
//
// Parameter:
//   - elem analysis.TraceElement: the element to check for
//
// Returns:
//   - bool: true if the chain contains elem, false otherwise
func (ch *chain) contains(elem trace.TraceElement) bool {
	if elem == nil {
		return false
	}

	for _, c := range ch.elems {
		if elem.IsEqual(c) {
			return true
		}
	}

	return false
}

// Remove the first element from the chain
func (ch *chain) removeHead() {
	ch.elems = ch.elems[1:]
}

// Remove the last element from the chain
func (ch *chain) removeTail() {
	ch.elems = ch.elems[:len(ch.elems)-1]
}

// Return the first element of a chain
//
// Returns:
//   - analysis.TraceElement: the first element in the chain, or nil if chain is empty
func (ch *chain) firstElement() trace.TraceElement {
	if ch.len() == 0 {
		return nil
	}
	return ch.elems[0]
}

// Return the last element of a chain
//
// Returns:
//   - analysis.TraceElement: the last element in the chain, or nil if chain is empty
func (ch *chain) lastElem() trace.TraceElement {
	if ch.len() == 0 {
		return nil
	}
	return ch.elems[len(ch.elems)-1]
}

// Swap the two elements in the chain given by the indexes.
// If at least on index is not in the chain, nothing is done
//
// Parameter:
//   - i int: index of the first element
//   - j int: index of the second element
func (ch *chain) swap(i, j int) {
	if i >= 0 && i < len(ch.elems) && j >= 0 && j < len(ch.elems) {
		ch.elems[i], ch.elems[j] = ch.elems[j], ch.elems[i]
	}
}

// Create a copy of the chain
//
// Returns:
//   - chain: a copy of the chain
func (ch *chain) copy() chain {
	newElems := make([]trace.TraceElement, len(ch.elems))

	for i, elem := range ch.elems {
		newElems[i] = elem
	}

	newChain := chain{
		elems: newElems,
	}
	return newChain
}

// Get the number of elements in a scheduling chain
//
// Returns:
//   - the number of elements in the chain
func (ch *chain) len() int {
	return len(ch.elems)
}

// Get a string representation of a scheduling chain
//
// Returns:
//   - A string representation of the chain
func (ch *chain) toString() string {
	res := ""
	for _, e := range ch.elems {
		res += fmt.Sprintf("%d:%s&", e.GetRoutine(), e.GetPos())
	}
	return res
}

// Check if a chain is valid.
// A chain is valid if it isn't violation the HB relation
// If the analyzer did not run and therefore did not calculate the HB relation,
// the function will always return true
// Since HB relations are transitive, it is enough to check neighboring elements
//
// Returns:
//   - bool: True if the mutation is valid, false otherwise
func (ch *chain) isValid() bool {
	if !data.HBWasCalc() {
		return true
	}

	for i := range ch.len() - 1 {
		hb := clock.GetHappensBefore(ch.elems[i].GetWVc(), ch.elems[i+1].GetWVc())
		if hb == clock.After {
			return false
		}
	}

	return true
}
