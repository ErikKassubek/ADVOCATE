// Copyright (c) 2025 Erik Kassubek
//
// File: csst.go
// Brief: Data to build and use the Collective Sparse Segment Trees
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package data

import (
	"advocate/analysis/concurrent/cssts"
	"advocate/trace"
)

var (
	Csst cssts.IncrementalCSST
)

func InitCSSTs() {
	lengths := make([]int, MainTrace.GetNoRoutines())
	for i, trace := range MainTrace.GetTraces() {
		lengths[i] = len(trace)
	}

	Csst = cssts.NewIncrementalCSST(lengths)
}

func AddEdgeCSSTs(from, to trace.TraceElement) {
	fromInd := cssts.GetIndicesFromTraceElem(from)
	toInd := cssts.GetIndicesFromTraceElem(to)
	Csst.InsetEdge(fromInd, toInd)
}
