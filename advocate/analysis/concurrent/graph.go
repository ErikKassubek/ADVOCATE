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

// For a given element, add it the the children of the last element that
// was analyzed in the same routine.
// Then set this element to be the last element analyzed in the routine
func AddEdgePartialOrderGraph(elem trace.TraceElement) {
	routineID := elem.GetRoutine()

	if lastElem, ok := data.LastAnalyzedElementPerRoutine[routineID]; ok {
		lastElem.AddChild(elem)
	} else {
		// first element, add edge from fork if exists
		if fork, okF := data.ForkOperations[routineID]; okF {
			fork.AddChild(elem)
		}
	}
	data.LastAnalyzedElementPerRoutine[routineID] = elem
}

// For a given element, find one or all elements that are concurrent to it
//
// Parameter
//   - elem TraceElement: the element the results should be concurrent with
//   - all bool: if true, return all elements that are concurrent, if false, only return one
//
// Returns
//   - []trace.TraceElement: element(s) that are concurrent to elem
func GetConcurrentPartialOrderGraph(elem trace.TraceElement, all bool) []trace.TraceElement {
	visitedFromN := make(map[int]bool)
	visitedToN := make(map[int]bool)

	dfsPartialOrderGraph(elem, visitedFromN, false)
	dfsPartialOrderGraph(elem, visitedToN, true)

	res := make([]trace.TraceElement, 0)

	for rout, trace := range data.MainTrace.GetTraces() {
		if rout == elem.GetRoutine() {
			continue
		}

		for _, tElem := range trace {
			if !visitedFromN[tElem.GetTraceID()] && !visitedToN[tElem.GetTraceID()] {
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
func dfsPartialOrderGraph(start trace.TraceElement, visited map[int]bool, inverted bool) {
	stack := []trace.TraceElement{start}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[curr.GetTraceID()] {
			continue
		}
		visited[curr.GetTraceID()] = true
		if inverted {
			for _, parent := range curr.GetParents() {
				if !visited[parent.GetTraceID()] {
					stack = append(stack, parent)
				}
			}
		} else {
			for _, child := range curr.GetChildren() {
				if !visited[child.GetTraceID()] {
					stack = append(stack, child)
				}
			}
		}
	}
}
