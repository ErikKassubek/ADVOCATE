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

import "advocate/analysis/hb/clock"

type ElemMin struct {
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

func NewTraceMin() TraceMin {
	return TraceMin{
		trace: make([]ElemMin, 0),
	}
}

// AddElem adds an elem to the min trace,
func (this *TraceMin) AddElem(elem ElemMin) {
	this.trace = append(this.trace, elem)
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
