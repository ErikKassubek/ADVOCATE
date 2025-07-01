// Copyright (c) 2025 Erik Kassubek
//
// File: partialOrderGraph.go
// Brief: Function to find concurrent operations to an operations
//
// Author: Erik Kassubek
// Created: 2025-06-29
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/clock"
	"advocate/trace"
	"advocate/utils/log"
)

// =========== Find by VC based brute force ============

// For a given element, find a/all element(s) that are concurrent to it
// This function assumes that the vector clocks have already been calculated
// The function iterates over all elements, and compares the vector clocks
//
// Parameter:
//   - elem trace.TraceElement: the element to search for
//   - all bool: if true, find all concurrent elements, if false, find only one
func getConcurrentBruteForce(elem trace.TraceElement, all bool) []trace.TraceElement {
	if !HBWasCalc() {
		log.Error("Cannot find concurrent elements: VCs have not been calculated")
		return make([]trace.TraceElement, 0)
	}

	res := make([]trace.TraceElement, 0)
	for rout, trace := range MainTrace.GetTraces() {
		if rout == elem.GetRoutine() {
			continue
		}

		for _, tElem := range trace {
			if clock.GetHappensBefore(elem.GetWVc(), tElem.GetWVc()) == clock.Concurrent {
				res = append(res, tElem)
				if !all {
					return res
				}
			}
		}
	}

	return res
}

// =========== Find based on tree =======================

// For a given element, add it the the children of the last element that
// was analyzed in the same routine.
// Then set this element to be the last element analyzed in the routine
func addEdgePartialOrderGraph(elem trace.TraceElement) {
	routineID := elem.GetRoutine()

	if lastElem, ok := lastAnalyzedElementPerRoutine[routineID]; ok {
		lastElem.AddChild(elem)
	} else {
		// first element, add edge from fork if exists
		if fork, okF := forkOperations[routineID]; okF {
			fork.AddChild(elem)
		}
	}
	lastAnalyzedElementPerRoutine[routineID] = elem
}

// For a given element, find one or all elements that are concurrent to it
//
// Parameter
//   - elem TraceElement: the element the results should be concurrent with
//   - all bool: if true, return all elements that are concurrent, if false, only return one
//
// Returns
//   - []trace.TraceElement: element(s) that are concurrent to elem
func getConcurrentPartialOrderGraph(elem trace.TraceElement, all bool) []trace.TraceElement {
	visitedFromN := make(map[int]bool)
	visitedToN := make(map[int]bool)

	dfsPartialOrderGraph(elem, visitedFromN, false)
	dfsPartialOrderGraph(elem, visitedToN, true)

	res := make([]trace.TraceElement, 0)

	for rout, trace := range MainTrace.GetTraces() {
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
