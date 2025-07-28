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
	anaData "advocate/analysis/data"
	"advocate/analysis/hb"
	"advocate/analysis/hb/concurrent"
	"advocate/analysis/hb/hbcalc"
	"advocate/fuzzing/data"
	"advocate/trace"
	"fmt"
	"math"
	"math/rand"
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

type elemWithQual struct {
	elem    trace.Element
	quality int
}

// startChains returns a slice of chain consisting of a
// pair of operations that are in a rel2 relation
//
// Parameter:
//   - num int: number of chains to return
//
// Returns:
//   - the chain, or an empty chain if no pair exists
func startChains(num int) []Chain {
	res := make([]Chain, 0)

	if data.UseHBInfoFuzzing {
		traces := anaData.MainTrace.GetTraces()

		if len(traces) == 0 {
			return res
		}

		top := []elemWithQual{}

		for i := 0; i < 1000; i++ {
			key := rand.Intn(len(traces)) + 1
			trace := traces[key]
			if len(trace) == 0 {
				continue
			}

			ind := rand.Intn(len(trace))
			elem := trace[ind]

			if !data.CanBeAddedToChain(elem) {
				continue
			}

			if concurrent.GetNumberConcurrent(elem, sameElem, false) == 0 {
				continue
			}

			q := quality(elem, sameElem)

			e := elemWithQual{elem, q}

			// find the num with the best quality
			inserted := false
			for i, v := range top {
				if e.quality > v.quality {
					top = append(top[:i+1], top[i:]...)
					top[i] = e
					inserted = true
					break
				}
			}
			if !inserted && len(top) < num {
				top = append(top, e)
			}

			// Trim if longer than n
			if len(top) > num {
				top = top[:num]
			}
		}

		if len(top) == 0 {
			return res
		}

		for _, e := range top {
			posPartner := concurrent.GetConcurrent(e.elem, true, sameElem, true)
			if len(posPartner) == 0 {
				continue
			}

			partner := posPartner[rand.Intn(len(posPartner))]
			c := NewChain()
			c.add(e.elem)
			c.add(partner)
			res = append(res, c)
		}
	} else {
		// start with two random elements in rel2
		i := 0
		for elem1, rel := range rel2 {
			for elem2 := range rel {
				c := NewChain()
				c.add(elem1)
				c.add(elem2)
				res = append(res, c)
				i++
				if i > num {
					return res
				}
			}
		}
	}

	return res
}

// quality calculates how fit for mutation an element is
// This is based on how many times was an operation called on the same element
// and how many concurrent operation has the operations
//
// Parameters:
//   - elem trace.Element: the element to check for
//   - sameElem bool: only consider concurrent elements on the same element
//
// Returns:
//   - the quality
func quality(elem trace.Element, sameElem bool) int {
	numberOps, _ := anaData.GetOpsPerID(elem.GetID())
	numberConcurrent := concurrent.GetNumberConcurrent(elem, sameElem, true)
	return int(math.Log(float64(1+numberOps)) + math.Log(float64(1+numberConcurrent)))
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
//
// Return:
//   - bool: true if swap was possible, false otherwise
func (ch *Chain) swap(i, j int) bool {
	if hbcalc.GetHappensBefore(ch.Elems[i], ch.Elems[j], true) != hb.Concurrent {
		return false
	}
	if i >= 0 && i < len(ch.Elems) && j >= 0 && j < len(ch.Elems) {
		ch.Elems[i], ch.Elems[j] = ch.Elems[j], ch.Elems[i]
	}
	return true
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

// Len returns the number of elements in a scheduling chain
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
	for i := range ch.Len() - 1 {
		hbInfo := hbcalc.GetHappensBefore(ch.Elems[i], ch.Elems[i+1], true)
		if hbInfo == hb.After {
			return false
		}
	}

	return true
}
