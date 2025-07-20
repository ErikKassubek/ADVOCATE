// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-sc.go
// Brief: scheduling Chains for GoPie
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package gopie

import (
	"advocate/analysis/data"
	anaData "advocate/analysis/data"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/concurrent"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/types"
	"fmt"
	"math"
)

// Chain is a representation of a scheduling Chain
// A Chain is an ordered list of adjacent element from the trace,
// where two neighboring elements must be from different routines
type Chain struct {
	Elems []trace.Element
}

// NewChain create a new, empty chain
//
// Returns: chain: the new chain
func NewChain() Chain {
	elems := make([]trace.Element, 0)
	return Chain{elems}
}

// startChain returns a chain consisting of a
// pair of operations that are in a rel2 relation
//
// Returns:
//   - the chain, or an empty chain if no pair exists
func startChain() Chain {
	res := NewChain()

	maxQuality := math.MinInt
	var elemMax types.Pair[trace.Element, trace.Element]

	for elem1, rel := range rel2 {
		qElem1 := quality(elem1)
		for elem2 := range rel {
			q := qElem1 + quality(elem2)
			if q > maxQuality {
				p := types.NewPair(elem1, elem2)
				if !helper.Contains(usedStartPos, p) {
					maxQuality = q
					elemMax = p
				}
			}

			return res
		}
	}

	if maxQuality > math.MaxInt {
		res.add(elemMax.X)
		res.add(elemMax.Y)
		usedStartPos = append(usedStartPos, elemMax)
	}

	return res
}

// quality calculates how fit for mutation an element is
// This is based on how many times was an operation called on the same element
// and how many concurrent operation has the operations
//
// Parameters:
//   - elem trace.Element: the element to check for
//
// Returns:
//   - the quality
func quality(elem trace.Element) int {
	numberOps, _ := data.GetOpsPerID(elem.GetID())
	numberConcurrent := concurrent.GetNumberConcurrent(elem, true, true)
	return int(math.Log(float64(1+numberOps)) + math.Log(float64(numberConcurrent)))
}

// Add a new element to the chain
//
// Parameter:
//   - elem analysis.TraceElement: Element to add
func (ch *Chain) add(elem trace.Element) {
	if elem == nil {
		return
	}

	ch.Elems = append(ch.Elems, elem)
}

// replace replaces the element at a given index in a chain with another element
//
// Parameter:
//   - index int: index to change at
//   - elem analysis.TraceElement: element to set at index
func (ch *Chain) replace(index int, elem trace.Element) {
	if elem == nil {
		return
	}

	if index < 0 || index >= len(ch.Elems) {
		return
	}
	ch.Elems[index] = elem
}

// Returns if the chain contains a specific element
//
// Parameter:
//   - elem analysis.TraceElement: the element to check for
//
// Returns:
//   - bool: true if the chain contains elem, false otherwise
func (ch *Chain) contains(elem trace.Element) bool {
	if elem == nil {
		return false
	}

	for _, c := range ch.Elems {
		if elem.IsEqual(c) {
			return true
		}
	}

	return false
}

// Remove the first element from the chain
func (ch *Chain) removeHead() {
	ch.Elems = ch.Elems[1:]
}

// Remove the last element from the chain
func (ch *Chain) removeTail() {
	ch.Elems = ch.Elems[:len(ch.Elems)-1]
}

// Return the first element of a chain
//
// Returns:
//   - analysis.TraceElement: the first element in the chain, or nil if chain is empty
func (ch *Chain) firstElement() trace.Element {
	if ch.Len() == 0 {
		return nil
	}
	return ch.Elems[0]
}

// Return the last element of a chain
//
// Returns:
//   - analysis.TraceElement: the last element in the chain, or nil if chain is empty
func (ch *Chain) lastElem() trace.Element {
	if ch.Len() == 0 {
		return nil
	}
	return ch.Elems[len(ch.Elems)-1]
}

// Swap the two elements in the chain given by the indexes.
// If at least on index is not in the chain, nothing is done
//
// Parameter:
//   - i int: index of the first element
//   - j int: index of the second element
func (ch *Chain) swap(i, j int) {
	if i >= 0 && i < len(ch.Elems) && j >= 0 && j < len(ch.Elems) {
		ch.Elems[i], ch.Elems[j] = ch.Elems[j], ch.Elems[i]
	}
}

// Create a copy of the chain
//
// Returns:
//   - chain: a copy of the chain
func (ch *Chain) copy() Chain {
	newElems := make([]trace.Element, len(ch.Elems))

	copy(newElems, ch.Elems)

	newChain := Chain{
		Elems: newElems,
	}
	return newChain
}

// Get the number of elements in a scheduling chain
//
// Returns:
//   - the number of elements in the chain
func (ch *Chain) Len() int {
	return len(ch.Elems)
}

// Get a string representation of a scheduling chain
//
// Returns:
//   - A string representation of the chain
func (ch *Chain) toString() string {
	res := ""
	for _, e := range ch.Elems {
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
func (ch *Chain) isValid() bool {
	if !anaData.HBWasCalc() {
		return true
	}

	for i := range ch.Len() - 1 {
		hbInfo := clock.GetHappensBefore(ch.Elems[i].GetWVc(), ch.Elems[i+1].GetWVc())
		if hbInfo == hb.After {
			return false
		}
	}

	return true
}
