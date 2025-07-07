// Copyright (c) 2025 Erik Kassubek
//
// File: addEdges.go
// Brief: Functions to add the required edges
//
// Author: Erik Kassubek
// Created: 2025-07-07
//
// License: BSD-3-Clause

package cssts

import (
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/types"
)

// Add an edge to the cssts
//
// Parameter:
//   - from trace.Element: start node
//   - to trace.Element: end node
func addEdgeCSSTsElem(from, to trace.Element) {
	fromInd := getIndicesFromTraceElem(from)
	toInd := getIndicesFromTraceElem(to)
	Csst.InsetEdge(fromInd, toInd)
}

// Add an edge to the cssts
//
// Parameter:
//   - from trace.Element: start node
//   - to trace.Element: end node
func addEdgeCSSTsIndex(from, to types.Pair[int, int]) {
	Csst.InsetEdge(from, to)
	CsstInverted.InsetEdge(to, from)
}

// Add an edge between a fork element and the first element in the newly
// crated routine
//
// Parameter:
//   - elem *trace.ElementFork: the fork element
func AddEdgeCSSTsFork(elem *trace.ElementFork) {
	routine, index := elem.GetTraceIndex()
	newRout := elem.GetID()
	if data.GetTraceLength(newRout) > 0 {
		addEdgeCSSTsIndex(
			types.Pair[int, int]{X: routine, Y: index},
			types.Pair[int, int]{X: newRout, Y: 0})
	}
}
