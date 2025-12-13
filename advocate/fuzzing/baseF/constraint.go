// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-sc.go
// Brief: scheduling Chains for GoPie
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package baseF

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb"
	"advocate/analysis/hb/concurrent"
	"advocate/analysis/hb/hbcalc"
	"advocate/trace"
	"advocate/utils/flags"
	"fmt"
	"math"
	"math/rand/v2"
)

// Constraint is a representation of a scheduling Constraint
// A Constraint is an ordered list of adjacent element from the trace,
// where two neighboring elements must be from different routines
type Constraint struct {
	Elems []trace.Element
	Old   bool
}

// NewConstraint create a new, empty chain
//
// Returns: chain: the new chain
func NewConstraint() Constraint {
	elems := make([]trace.Element, 0)
	return Constraint{elems, false}
}

type ElemWithQual struct {
	Elem    trace.Element
	Quality float64
}

// Add a new element to the chain
//
// Parameter:
//   - elems ...analysis.TraceElement: Elements to Add, in the order they are added
func (this *Constraint) Add(elems ...trace.Element) {
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

// Replace replaces the element at a given index in a chain with another element
//
// Parameter:
//   - index int: index to change at
//   - elem analysis.TraceElement: element to set at index
func (this *Constraint) Replace(index int, elem trace.Element) {
	if elem == nil {
		return
	}

	if index < 0 || index >= len(this.Elems) {
		return
	}
	this.Elems[index] = elem
}

// Returns if the chain Contains a specific element
//
// Parameter:
//   - elem analysis.TraceElement: the element to check for
//
// Returns:
//   - bool: true if the chain Contains elem, false otherwise
func (this *Constraint) Contains(elem trace.Element) bool {
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
func (this *Constraint) RemoveHead() {
	this.Elems = this.Elems[1:]
}

// Remove the last element from the chain
func (this *Constraint) RemoveTail() {
	this.Elems = this.Elems[:len(this.Elems)-1]
}

// Return the first element of a chain
//
// Returns:
//   - analysis.TraceElement: the first element in the chain, or nil if chain is empty
func (this *Constraint) ElemWithSmallestTPost() trace.Element {
	if this.Len() == 0 {
		return nil
	}

	var min trace.Element

	for _, c := range this.Elems {
		if min == nil || c.GetTSort() < min.GetTSort() {
			min = c
		}
	}

	return min
}

// Return the last element of a chain
//
// Returns:
//   - analysis.TraceElement: the last element in the chain, or nil if chain is empty
func (this *Constraint) LastElem() trace.Element {
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
//   - bool: true if Swap was possible, false otherwise
func (this *Constraint) Swap(i, j int) bool {
	if hbcalc.GetHappensBefore(this.Elems[i], this.Elems[j], true) != hb.Concurrent {
		return false
	}
	if i >= 0 && i < len(this.Elems) && j >= 0 && j < len(this.Elems) {
		this.Elems[i], this.Elems[j] = this.Elems[j], this.Elems[i]
	}
	return true
}

// Create a Copy of the chain
//
// Returns:
//   - chain: a Copy of the chain
func (this *Constraint) Copy() Constraint {
	newElems := make([]trace.Element, len(this.Elems))

	copy(newElems, this.Elems)

	newChain := Constraint{
		Elems: newElems,
	}
	return newChain
}

// Len returns the number of elements in a scheduling chain
//
// Returns:
//   - the number of elements in the chain
func (this *Constraint) Len() int {
	return len(this.Elems)
}

// Get a string representation of a scheduling chain
//
// Returns:
//   - A string representation of the chain
func (this *Constraint) ToString() string {
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
func (this *Constraint) IsValid() bool {
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

func (this *Constraint) MutSelect() map[string]Constraint {
	res := make(map[string]Constraint)

	for i, elem := range this.Elems {
		if elem.GetType(false) != trace.Select {
			continue
		}

		sel := elem.(*trace.ElementSelect)
		chosen := sel.GetChosenCase()
		if chosen != nil {
			partner := sel.GetChosenCase().GetPartner()
			if partner != nil && this.Contains(partner) {
				continue
			}
		}

		if sel.GetContainsDefault() && !sel.GetChosenDefault() {
			c := this.Copy()
			c.Elems[i].(*trace.ElementSelect).SetCaseByIndex(-1)
			res[c.ToString()] = c
		}

		for ca := range sel.GetCases() {
			if ca == sel.GetChosenIndex() {
				continue
			}

			c := this.Copy()
			c.Elems[i].(*trace.ElementSelect).SetCaseByIndex(ca)
			res[c.ToString()] = c
		}
	}

	return res
}

// CanBeAddedToConstraint decides if an element can be added to a scheduling chain
// For GoPie without improvements (!useHBInfoFuzzing) those are only mutex and channel (incl. select)
// With improvements those are all not ignored fuzzing elements
//
// Parameter:
//   - elem analysis.TraceElement: Element to check
//
// Returns:
//   - true if it can be added to a scheduling chain, false otherwise
func CanBeAddedToConstraint(elem trace.Element) bool {
	t := elem.GetType(false)
	if flags.FuzzingMode == GoPie {
		// for standard GoPie, only mutex, channel and select operations are considered
		return t == trace.Mutex || t == trace.Channel || t == trace.Select
	}

	return t != trace.Atomic && !IgnoreFuzzing(elem, true)
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
func Quality(elem trace.Element) float64 {
	w1 := 0.2
	w2 := 0.3
	w3 := 0.5

	numberOps, _ := baseA.GetOpsPerID(elem.GetObjId())
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
