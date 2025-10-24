// Copyright (c) 2025 Erik Kassubek
//
// File: traceMin.go
// Brief: Functions and structs to implement a minimal representation of a trace
//
// Author: Erik Kassubek
// Created: 2025-10-22
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"fmt"
)

var minIndex = 0

type ElemMin struct {
	Index   int
	ID      int
	Op      OperationType
	Pos     string
	Routine int
	Vc      clock.VectorClock
	Channel []int // id of channel in select cases
}

func (this *ElemMin) GetType(op bool) OperationType {
	if op {
		return this.Op
	}
	return GetElemTypeFromObjectType(this.Op)
}

func (this *ElemMin) IsSameElement(elem *ElemMin) bool {
	return this.ID == elem.ID
}

func (this *ElemMin) Key() string {
	return fmt.Sprintf("%d-%d", this.Routine, this.Index)
}

// TODO: make better check (include element/channel). First must make sure, that id is always the same
func (this *ElemMin) IsEqual(other *ElemMin) bool {
	return this.Pos == other.Pos
}

func (this *ElemMin) HasCommonChannel(elem *ElemMin) bool {
	seen := make(map[int]struct{}, len(this.Channel))

	for _, v := range this.Channel {
		seen[v] = struct{}{}
	}

	for _, v := range elem.Channel {
		if _, ok := seen[v]; ok {
			return true
		}
	}

	return false
}

func (this *ElemMin) IsInCases(elem *ElemMin) bool {
	for _, v := range this.Channel {
		if v == elem.ID {
			return true
		}
	}

	return false
}

type TraceMin struct {
	trace []ElemMin
}

// NewTraceMin creates a new, empty trace
//
// Returns:
//   - TraceMin: a new, empty trace
func NewTraceMin() TraceMin {
	return TraceMin{
		trace: make([]ElemMin, 0),
	}
}

// Len returns the number of elements in the TraceMin
//
// Returns:
//   - int: the number of elements in the trace
func (this *TraceMin) Len() int {
	return len(this.trace)
}

// Create a copy of the TraceMin
//
// Returns:
//   - TraceMin: the copy
func (this *TraceMin) Clone() TraceMin {
	copy := make([]ElemMin, 0)
	for i := 0; i < len(this.trace); i++ {
		copy = append(copy, this.trace[i])
	}
	return TraceMin{
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
func (this *TraceMin) CloneSub(start, end int) TraceMin {
	copy := this.trace[start:end]
	return TraceMin{
		trace: copy,
	}
}

// Swap two elements in the TraceMin
//
// Parameter:
//   - i: first index
//   - j: second index
func (this *TraceMin) Flip(i, j int) {
	this.trace[i], this.trace[j] = this.trace[j], this.trace[i]
}

// AddElem adds an elem to the min trace
//
// Parameter:
//   - elem ElemMin: the element to add
func (this *TraceMin) AddElem(elem ElemMin) {
	this.trace = append(this.trace, elem)
}

// IsEqual returns if two traceMin are equal
//
// Parameter:
//   - other *TraceMin: the trace to compare against
//
// Returns:
//   - bool: true if the traces are equal, false otherwise
func (this *TraceMin) IsEqual(other *TraceMin) bool {
	if this.Len() != other.Len() {
		return false
	}

	for i := 0; i < this.Len(); i++ {
		if !this.trace[i].IsEqual(&other.trace[i]) {
			return false
		}
	}

	return true
}

func (this *TraceMin) Get(index int) *ElemMin {
	return &this.trace[index]
}

func (this *TraceMin) Key() string {
	res := ""
	for _, elem := range this.trace {
		res += (elem.Key() + ";")
	}
	return res
}

// TraceMinFromTrace creates a trace min from a trace
//
// Parameter:
//   - trace *Trace: the full trace
//
// Returns:
//   - TraceMin: the minimal trace
func TraceMinFromTrace(trace *Trace) TraceMin {
	res := NewTraceMin()

	traceIter := trace.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		minElem, val := elem.GetElemMin()
		if val {
			res.trace = append(res.trace, minElem)
		}
	}

	return res
}
