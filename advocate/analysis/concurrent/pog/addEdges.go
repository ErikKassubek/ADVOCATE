// Copyright (c) 2025 Erik Kassubek
//
// File: addEdges.go
// Brief: Add edges to the graph
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
func AddEdgeSameRoutineAndFork(elem trace.Element) {
	if !helper.Valid(elem) {
		return
	}
	routineID := elem.GetRoutine()

	if lastElem, ok := data.LastAnalyzedElementPerRoutine[routineID]; ok {
		AddEdge(lastElem, elem, true)
	} else {
		// first element, add edge from fork if exists
		if fork, okF := data.ForkOperations[routineID]; okF {
			AddEdge(fork, elem, true)
		}
	}
	data.LastAnalyzedElementPerRoutine[routineID] = elem
}
