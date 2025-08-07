//
// File: addEdges.go
// Brief: Functions to add the required edges
//
// Created: 2025-07-07
//
// License: BSD-3-Clause

package cssts

import (
	"goCR/analysis/data"
	"goCR/trace"
	"goCR/utils/types"
)

// AddEdge adds an edge to the cssts
//
// Parameter:
//   - from trace.Element: start node
//   - to trace.Element: end node
//   - weak bool: if true, add to weak hb
func AddEdge(from, to trace.Element, weak bool) {
	if from == nil || to == nil {
		return
	}

	fromInd := getIndicesFromTraceElem(from)
	toInd := getIndicesFromTraceElem(to)
	Csst.AddEdge(fromInd, toInd)
	CsstInverted.AddEdge(toInd, fromInd)
	if weak {
		CsstWeak.AddEdge(fromInd, toInd)
		CsstWeakInverted.AddEdge(toInd, fromInd)
	}
}

// Add an edge to the cssts
//
// Parameter:
//   - from trace.Element: start node
//   - to trace.Element: end node
//   - weak bool: if true, add to weak hb
func addEdgeIndex(from, to types.Pair[int, int], weak bool) {
	Csst.AddEdge(from, to)
	CsstInverted.AddEdge(to, from)

	if weak {
		CsstWeak.AddEdge(from, to)
		CsstWeakInverted.AddEdge(to, from)
	}
}

// AddEdgeFork adds an edge between a fork element and the first element in the newly
// crated routine
//
// Parameter:
//   - elem *trace.ElementFork: the fork element
func AddEdgeFork(elem *trace.ElementFork) {
	routine, index := elem.GetTraceIndex()
	newRout := elem.GetID()
	if data.GetTraceLength(newRout) > 0 {
		addEdgeIndex(
			types.Pair[int, int]{X: routine, Y: index},
			types.Pair[int, int]{X: newRout, Y: 0}, true)
	}
}
