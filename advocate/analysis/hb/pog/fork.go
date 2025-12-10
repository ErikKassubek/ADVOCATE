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
	"advocate/analysis/baseA"
	"advocate/analysis/hb/helper"
	"advocate/trace"
)

// AddEdgeSameRoutineAndFork adds a given element to the children of the last element that
// was analyzed in the same routine.
// Then set this element to be the last element analyzed in the routine
// If it is the first element in a routine, add an edge to the corresponding fork
//
// Parameter:
//   - graph PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - elem trace.Element: the element to add an edge for
func AddEdgeSameRoutineAndFork(graph *PoGraph, elem trace.Element) {
	if !helper.Valid(elem) {
		return
	}
	routineID := elem.GetRoutine()

	gr := graph
	if graph == nil {
		gr = &po
	}

	forks := baseA.ForkOperations
	if graph != nil {
		forks = gr.ForkOps
	}

	if lastElem, ok := gr.lastAdded[routineID]; ok {
		if graph != nil {
			graph.AddEdge(lastElem, elem)
		} else {
			AddEdge(lastElem, elem, true)
		}
	} else {
		// first element, add edge from fork if exists
		if fork, okF := forks[routineID]; okF {
			if graph != nil {
				graph.AddEdge(fork, elem)
			} else {
				AddEdge(fork, elem, true)
			}
		}
	}
	gr.lastAdded[routineID] = elem
}
