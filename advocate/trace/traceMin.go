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
	Op      ObjectType
	Pos     string
	Routine int
	vc      *clock.VectorClock
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
