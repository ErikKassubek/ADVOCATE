// Copyright (c) 2025 Erik Kassubek
//
// File: traceMin.go
// Brief: Functions and structs to implement a minimal representation of a trace
//
// Author: Erik Kassubek
// Created: 2025-10-22
//
// License: BSD-3-Clause

package equivalence

import (
	"advocate/analysis/hb/pog"
	"advocate/trace"
)

type TraceEq struct {
	trace        []trace.Element
	partialOrder pog.PoGraph
}

// NewTraceEq creates a new, empty trace
//
// Returns:
//   - TraceMin: a new, empty trace
func NewTraceEq() TraceEq {
	return TraceEq{
		trace: make([]trace.Element, 0),
	}
}

// Trace returns the elements in the TraceMin
//
// Returns:
//   - int: the elements in the trace
func (this *TraceEq) Trace() []trace.Element {
	return this.trace
}

// Len returns the number of elements in the TraceMin
//
// Returns:
//   - int: the number of elements in the trace
func (this *TraceEq) Len() int {
	return len(this.trace)
}

// Create a copy of the TraceMin
//
// Returns:
//   - TraceMin: the copy
func (this *TraceEq) Clone() TraceEq {
	copy := make([]trace.Element, 0)
	for i := 0; i < len(this.trace); i++ {
		copy = append(copy, this.trace[i])
	}
	return TraceEq{
		trace: copy,
	}
}

// Create a subtrace of the TraceMin
//
// Parameter:
//   - start int: start index
//   - end int: end index
//
// Returns:
//   - TraceMin: the trace min with subtrace
func (this *TraceEq) CloneSub(start, end int) TraceEq {
	copy := this.trace[start:end]
	return TraceEq{
		trace: copy,
	}
}

// Swap two elements in the TraceMin
//
// Parameter:
//   - i: first index
//   - j: second index
func (this *TraceEq) Flip(i, j int) {
	this.trace[i], this.trace[j] = this.trace[j], this.trace[i]
}

// AddElem adds an elem to the min trace
//
// Parameter:
//   - elem ElemMin: the element to add
func (this *TraceEq) AddElem(elem trace.Element) {
	this.trace = append(this.trace, elem)
}

// IsEqual returns if two traceMin are equal
//
// Parameter:
//   - other *TraceMin: the trace to compare against
//
// Returns:
//   - bool: true if the traces are equal, false otherwise
func (this *TraceEq) IsEqual(other *TraceEq) bool {
	if this.Len() != other.Len() {
		return false
	}

	for i := 0; i < this.Len(); i++ {
		if !this.trace[i].IsEqual(other.trace[i]) {
			return false
		}
	}

	return true
}

func (this *TraceEq) Get(index int) trace.Element {
	return this.trace[index]
}

// func (this *TraceMin) Key() string {
// 	res := ""
// 	for _, elem := range this.trace {
// 		res += (elem.Key() + ";")
// 	}
// 	return res
// }

// TraceEqFromTrace creates a trace min from a trace
//
// Parameter:
//   - t *Trace: the full trace
//
// Returns:
//   - TraceMin: the minimal trace
func TraceEqFromTrace(t *trace.Trace) TraceEq {
	res := NewTraceEq()

	traceIter := t.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		if trace.IsOp(elem) {
			res.trace = append(res.trace, elem)
		}
	}

	return res
}
