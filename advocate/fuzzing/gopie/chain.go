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
	"advocate/utils/settings.go"
	"fmt"
	"math"
	"math/rand"
)

// Chain is a representation of a scheduling Chain
// A Chain is an ordered list of adjacent element from the trace,
// where two neighboring elements must be from different routines
type Chain struct {
	Elems []trace.Element
	old   bool
}

// NewChain create a new, empty chain
//
// Returns: chain: the new chain
func NewChain() Chain {
	elems := make([]trace.Element, 0)
	return Chain{elems, false}
}

type elemWithQual struct {
	elem    trace.Element
	quality float64
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

			if concurrent.GetNumberConcurrent(elem, sameElem, settings.SameElementTypeInSC, false) == 0 {
				continue
			}

			q := quality(elem)

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

		rounds := settings.GoPieMaxSCLength
		if !settings.GoPieMaxSCLengthSet {
			rounds = 1
		}
		for i := 0; i < rounds; i++ {
			if len(res) == 0 {
				for _, e := range top {
					posPartner := concurrent.GetConcurrent(e.elem, true, true, settings.SameElementTypeInSC, true)
					if len(posPartner) == 0 {
						posPartner = concurrent.GetConcurrent(e.elem, true, false, settings.SameElementTypeInSC, true)
						if len(posPartner) == 0 {
							continue
						}
					}

					partner := posPartner[rand.Intn(len(posPartner))]
					c := NewChain()
					c.add(e.elem, partner)
					res = append(res, c)
				}
			} else {
				for i, c := range res {
					lastElem := c.lastElem()
					if lastElem == nil {
						continue
					}

					posNext := concurrent.GetConcurrent(lastElem, true, true, settings.SameElementTypeInSC, true)
					if len(posNext) == 0 {
						continue
					}
					next := posNext[rand.Intn(len(posNext))]
					res[i].add(next)
				}
			}
		}
	} else {
		// start with two random elements in rel2
		i := 0
		for elem1, rel := range rel2 {
			for elem2 := range rel {
				c := NewChain()
				c.add(elem1, elem2)
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
//
// Returns:
//   - float64: the quality
func quality(elem trace.Element) float64 {
	w1 := 0.2
	w2 := 0.3
	w3 := 0.5

	numberOps, _ := anaData.GetOpsPerID(elem.GetID())
	numberConcurrentTotal := concurrent.GetNumberConcurrent(elem, false, false, true)
	numberConcurrentSame := concurrent.GetNumberConcurrent(elem, true, false, true)

	if numberConcurrentSame == 0 && numberConcurrentTotal == 0 {
		return 0
	}

	q := w1*math.Log1p(float64(numberOps)) +
		w2*float64(numberConcurrentSame)/float64(numberConcurrentTotal+1) +
		w3*math.Log1p(float64(numberConcurrentTotal))

	return q * ((rand.Float64() * 0.2) - 0.1)
}

// Add a new element to the chain
//
// Parameter:
//   - elems ...analysis.TraceElement: Elements to add, in the order they are added
func (this *Chain) add(elems ...trace.Element) {
	if elems == nil {
		return
	}

	for _, elem := range elems {
		if elem == nil {
			continue
		}
		this.Elems = append(this.Elems, elem)
	}
}

// replace replaces the element at a given index in a chain with another element
//
// Parameter:
//   - index int: index to change at
//   - elem analysis.TraceElement: element to set at index
func (this *Chain) replace(index int, elem trace.Element) {
	if elem == nil {
		return
	}

	if index < 0 || index >= len(this.Elems) {
		return
	}
	this.Elems[index] = elem
}

// Returns if the chain contains a specific element
//
// Parameter:
//   - elem analysis.TraceElement: the element to check for
//
// Returns:
//   - bool: true if the chain contains elem, false otherwise
func (this *Chain) contains(elem trace.Element) bool {
	if elem == nil {
		return false
	}

	for _, c := range this.Elems {
		if elem.IsEqual(c) {
			return true
		}
	}

	return false
}

// Remove the first element from the chain
func (this *Chain) removeHead() {
	this.Elems = this.Elems[1:]
}

// Remove the last element from the chain
func (this *Chain) removeTail() {
	this.Elems = this.Elems[:len(this.Elems)-1]
}

// Return the first element of a chain
//
// Returns:
//   - analysis.TraceElement: the first element in the chain, or nil if chain is empty
func (this *Chain) firstElement() trace.Element {
	if this.Len() == 0 {
		return nil
	}

	var min trace.Element

	for _, c := range this.Elems {
		if min == nil || c.GetTPost() < min.GetTPost() {
			min = c
		}
	}

	return min
}

// Return the last element of a chain
//
// Returns:
//   - analysis.TraceElement: the last element in the chain, or nil if chain is empty
func (this *Chain) lastElem() trace.Element {
	if this.Len() == 0 {
		return nil
	}
	return this.Elems[len(this.Elems)-1]
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
func (this *Chain) swap(i, j int) bool {
	if hbcalc.GetHappensBefore(this.Elems[i], this.Elems[j], true) != hb.Concurrent {
		return false
	}
	if i >= 0 && i < len(this.Elems) && j >= 0 && j < len(this.Elems) {
		this.Elems[i], this.Elems[j] = this.Elems[j], this.Elems[i]
	}
	return true
}

// Create a copy of the chain
//
// Returns:
//   - chain: a copy of the chain
func (this *Chain) copy() Chain {
	newElems := make([]trace.Element, len(this.Elems))

	copy(newElems, this.Elems)

	newChain := Chain{
		Elems: newElems,
	}
	return newChain
}

// Len returns the number of elements in a scheduling chain
//
// Returns:
//   - the number of elements in the chain
func (this *Chain) Len() int {
	return len(this.Elems)
}

// Get a string representation of a scheduling chain
//
// Returns:
//   - A string representation of the chain
func (this *Chain) toString() string {
	res := ""
	for _, e := range this.Elems {
		res += fmt.Sprintf("%d:%s", e.GetRoutine(), e.GetPos())
		switch f := e.(type) {
		case *trace.ElementSelect:
			res += fmt.Sprintf("%d", f.GetChosenIndex())
		}
		res += "&"
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
func (this *Chain) isValid() bool {
	for i := 0; i < len(this.Elems)-1; i++ {
		for j := 1; j < len(this.Elems); j++ {
			hbInfo := hbcalc.GetHappensBefore(this.Elems[i], this.Elems[j], true)
			if hbInfo == hb.After {
				return false
			}
		}
	}

	return true
}

func (this *Chain) mutSelect() map[string]Chain {
	res := make(map[string]Chain)

	for i, elem := range this.Elems {
		if elem.GetType(false) != trace.Select {
			continue
		}

		sel := elem.(*trace.ElementSelect)
		chosen := sel.GetChosenCase()
		if chosen != nil {
			partner := sel.GetChosenCase().GetPartner()
			if partner != nil && this.contains(partner) {
				continue
			}
		}

		if sel.GetContainsDefault() && !sel.GetChosenDefault() {
			c := this.copy()
			c.Elems[i].(*trace.ElementSelect).SetCaseByIndex(-1)
			res[c.toString()] = c
		}

		for ca := range sel.GetCases() {
			if ca == sel.GetChosenIndex() {
				continue
			}

			c := this.copy()
			c.Elems[i].(*trace.ElementSelect).SetCaseByIndex(ca)
			res[c.toString()] = c
		}
	}

	return res
}
