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
	"analyzer/analysis"
	"analyzer/clock"
)

var (
	schedulingChains []chain
	currentChain     chain
	lastRoutine      = -1
)

type chain struct {
	elems []analysis.TraceElement
}

func newChain() chain {
	elems := make([]analysis.TraceElement, 0)
	return chain{elems}
}

/*
 * Traverse all elements in the trace in order of execution.
 * From this build the scheduling chains. A scheduling chain
 * is always the segment of maximum length, such that
 * to neighboring elements in the chain are neighbors in the global trace
 * and two neighboring elements in the chain are not in the same routine
 */
func addElemToChain(elem analysis.TraceElement) {
	routine := elem.GetRoutine()

	// add elem if the last routine is different from the routine of the elem
	// if the current routine is empty, lastRoutine is -1 and this is always true
	if lastRoutine != routine {
		currentChain.add(elem.Copy())
	} else {
		// if the routine is the same as the last routine, we need to start a new
		// chain. In this case, store the current chain as a scheduling chains
		// and start a new routine with the current element
		schedulingChains = append(schedulingChains, currentChain)
		currentChain = newChain()
		currentChain.add(elem.Copy())
	}

	lastRoutine = routine
}

func (ch *chain) add(elem analysis.TraceElement) {
	if elem == nil {
		return
	}

	ch.elems = append(ch.elems, elem)
}

func (ch *chain) replace(index int, elem analysis.TraceElement) {
	if elem == nil {
		return
	}

	if index < 0 || index >= len(ch.elems) {
		return
	}
	ch.elems[index] = elem
}

func (ch *chain) contains(elem analysis.TraceElement) bool {
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

func (ch *chain) removeHead() {
	ch.elems = ch.elems[1:]
}

func (ch *chain) removeTail() {
	ch.elems = ch.elems[:len(ch.elems)-1]
}

func (ch *chain) lastElem() analysis.TraceElement {
	return ch.elems[len(ch.elems)-1]
}

func (ch *chain) swap(i, j int) {
	if i >= 0 && i < len(ch.elems) && j >= 0 && j < len(ch.elems) {
		ch.elems[i], ch.elems[j] = ch.elems[j], ch.elems[i]
	}
}

func (ch *chain) copy() chain {
	newElems := make([]analysis.TraceElement, len(ch.elems))

	for i, elem := range ch.elems {
		newElems[i] = elem
	}

	newChain := chain{
		elems: newElems,
	}
	return newChain
}

func (ch *chain) len() int {
	return len(ch.elems)
}

func (ch *chain) toString() string {
	res := ""
	for _, e := range ch.elems {
		res += e.ToString()
	}
	return res
}

/*
 * Check if a chain is valid.
 * A chain is valid if it isn't violation the HB relation
 * If the analyzer did not run and therefore did not calculate the HB relation,
 * the function will always return true
 * Since HB relations are transitive, it is enough to check neighboring elements
 * Returns:
 * 	(bool): True if the mutation is valid, false otherwise
 */
func (ch *chain) isValid() bool {
	if !analysis.HBWasCalc() {
		return true
	}

	for i := range ch.len() - 1 {
		hb := clock.GetHappensBefore(ch.elems[i].GetVC(), ch.elems[i+1].GetVC())
		if hb == clock.After {
			return false
		}
	}

	return true
}
