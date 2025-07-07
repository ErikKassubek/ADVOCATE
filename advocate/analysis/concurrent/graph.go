// Copyright (c) 2025 Erik Kassubek
//
// File: compVC.go
// Brief: Function to find concurrent operations by building direct order graph
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package concurrent

import (
	"advocate/analysis/data"
	"advocate/trace"
)

// Add an edge between start and end
//
// Parameter:
//   - start trace.Element: the start element
//   - end trace.Element: the end element
func AddEdgePartialOrderGraph(start, end trace.Element) {
	if start == nil || end == nil {
		return
	}

	start.AddChild(end)
	end.AddParent(start)
}

// For a given element, add it to the children of the last element that
// was analyzed in the same routine.
// Then set this element to be the last element analyzed in the routine
// If it is the first element in a routine, add an edge to the corresponding fork
//
// Prarameter:
//   - elem trace.Element: the element to add an edge for
func AddEdgePartialOrderGraphSameRoutineAndFork(elem trace.Element) {
	if !valid(elem) {
		return
	}
	routineID := elem.GetRoutine()

	if lastElem, ok := data.LastAnalyzedElementPerRoutine[routineID]; ok {
		AddEdgePartialOrderGraph(lastElem, elem)
	} else {
		// first element, add edge from fork if exists
		if fork, okF := data.ForkOperations[routineID]; okF {
			AddEdgePartialOrderGraph(fork, elem)
		}
	}
	data.LastAnalyzedElementPerRoutine[routineID] = elem
}

// For a given element, find one or all elements that are concurrent to it
//
// Parameter
//   - elem TraceElement: the element the results should be concurrent with
//   - all bool: if true, return all elements that are concurrent, if false, only return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//
// Returns
//   - []trace.TraceElement: element(s) that are concurrent to elem
func GetConcurrentPartialOrderGraph(elem trace.Element, all bool, sameElem bool) []trace.Element {
	reachableFromN := make(map[int]bool)
	reachableToN := make(map[int]bool)

	dfsPartialOrderGraph(elem, reachableFromN, false)
	dfsPartialOrderGraph(elem, reachableToN, true)

	res := make([]trace.Element, 0)

	for rout, trace := range data.MainTrace.GetTraces() {
		if rout == elem.GetRoutine() {
			continue
		}

		for _, tElem := range trace {
			if sameElem && elem.GetID() != tElem.GetID() {
				continue
			}

			if !valid(tElem) {
				continue
			}

			if !reachableFromN[tElem.GetTraceID()] && !reachableToN[tElem.GetTraceID()] {
				res = append(res, tElem)
				if !all {
					return res
				}
			}
		}
	}
	return res
}

// Pass the partial order graph using dfs
// Store all visited nodes
//
// Parameter:
//   - start trace.TraceElement: element to start from
//   - visited map[int]bool: traceID of all visited nodes
//   - inverted bool: If false, find all nodes that can be reached from start,
//     if true, find all nodes from which start can be reached
func dfsPartialOrderGraph(start trace.Element, reachable map[int]bool, inverted bool) {
	stack := []trace.Element{start}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable[curr.GetTraceID()] {
			continue
		}
		reachable[curr.GetTraceID()] = true
		if inverted {
			for _, parent := range curr.GetParents() {
				if !reachable[parent.GetTraceID()] {
					stack = append(stack, parent)
				}
			}
		} else {
			for _, child := range curr.GetChildren() {
				if !reachable[child.GetTraceID()] {
					stack = append(stack, child)
				}
			}
		}
	}
}
