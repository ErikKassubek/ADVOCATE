// Copyright (c) 2025 Erik Kassubek
//
// File: addEdges.go
// Brief: Method for the different operations to add edges to the partial order
//        graph
//
// Author: Erik Kassubek
// Created: 2025-07-08
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/concurrent/helper"
	"advocate/analysis/data"
	"advocate/trace"
)

// For a given element, add it to the children of the last element that
// was analyzed in the same routine.
// Then set this element to be the last element analyzed in the routine
// If it is the first element in a routine, add an edge to the corresponding fork
//
// Prarameter:
//   - elem trace.Element: the element to add an edge for
func AddEdgePOGSameRoutineAndFork(elem trace.Element) {
	if !helper.Valid(elem) {
		return
	}
	routineID := elem.GetRoutine()

	if lastElem, ok := data.LastAnalyzedElementPerRoutine[routineID]; ok {
		AddEdge(lastElem, elem)
	} else {
		// first element, add edge from fork if exists
		if fork, okF := data.ForkOperations[routineID]; okF {
			AddEdge(fork, elem)
		}
	}
	data.LastAnalyzedElementPerRoutine[routineID] = elem
}

// AddEdgeAtomic adds a new edge for a last writer of an atomic
//
// Parameter:
//   - read *trace.ElementAtomic: the atomic read
//   - lw *trace.ElementAtomic: the last reader
func AddEdgeAtomic(read *trace.ElementAtomic, lw *trace.ElementAtomic) {
	if lw == nil {
		return
	}

	AddEdge(lw, read)
}
