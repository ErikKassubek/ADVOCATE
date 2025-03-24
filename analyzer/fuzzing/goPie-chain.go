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

import "analyzer/analysis"

var schedulingChains []chain
var currentChain chain = newChain()
var lastRoutine int = -1

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
		currentChain.add(elem)
	} else {
		// if the routine is the same as the last routine, we need to start a new
		// chain. In this case, store the current chain as a scheduling chains
		// and start a new routine with the current element
		schedulingChains = append(schedulingChains, currentChain)
		currentChain = newChain()
		currentChain.add(elem)
	}

	lastRoutine = routine
}

func (ch *chain) add(elem analysis.TraceElement) {
	ch.elems = append(ch.elems, elem)
}

func (ch *chain) replace(index int, elem analysis.TraceElement) {
	if index < 0 || index >= len(ch.elems) {
		return
	}
	ch.elems[index] = elem
}

func (ch *chain) contains(elem analysis.TraceElement) bool {
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
	if len(ch.elems) == 0 {
		return nil
	}
	return ch.elems[len(ch.elems)-1]
}

func (ch *chain) swap(i, j int) {
	if i >= 0 && i < len(ch.elems) && j >= 0 && j < len(ch.elems) {
		ch.elems[i], ch.elems[j] = ch.elems[j], ch.elems[i]
	}
}

func (ch *chain) copy() chain {
	newElems := make([]analysis.TraceElement, len(ch.elems))

	for _, elem := range ch.elems {
		newElems = append(newElems, elem)
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
