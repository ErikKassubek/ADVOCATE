// Copyright (c) 2025 Erik Kassubek
//
// File: hb.go
// Brief: Function to find concurrent operations by building direct order graph
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/data"
	"advocate/analysis/hb"
	"advocate/analysis/hb/helper"
	"advocate/trace"
	"reflect"
)

// AddEdge adds an edge between start and end
//
// Parameter:
//   - start trace.Element: the start element
//   - end trace.Element: the end element
//   - notWeak bool: if true, add to weak happens before
func AddEdge(start, end trace.Element, weak bool) {
	if start == nil || end == nil {
		return
	}

	po.addEdge(start, end)
	poInverted.addEdge(end, start)

	if weak {
		poWeak.addEdge(start, end)
		poWeakInverted.addEdge(end, start)
	}
}

// GetConcurrent find one or all elements that are concurrent to a given element
//
// Parameter
//   - elem TraceElement: the element the results should be concurrent with
//   - all bool: if true, return all elements that are concurrent, if false, only return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: if true, use weak hb, otherwise strong hb
//
// Returns
//   - []trace.TraceElement: element(s) that are concurrent to elem
func GetConcurrent(elem trace.Element, all bool, sameElem bool, weak bool) []trace.Element {
	reachableFromN := make(map[int]bool)
	reachableToN := make(map[int]bool)

	dfsPartialOrderGraph(elem, nil, reachableFromN, false, weak)
	dfsPartialOrderGraph(elem, nil, reachableToN, true, weak)

	res := make([]trace.Element, 0)

	for rout, trace := range data.MainTrace.GetTraces() {
		if rout == elem.GetRoutine() {
			continue
		}

		for _, tElem := range trace {
			if sameElem && elem.GetID() != tElem.GetID() {
				continue
			}

			if !helper.Valid(tElem) {
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

// GetHappensBefore returns the happens before relation between two operations given there
// POG
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - weak bool: get based on weak happens before
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(t1, t2 trace.Element, weak bool) hb.HappensBefore {
	if dfsPartialOrderGraph(t1, t2, nil, false, weak) {
		return hb.Before
	}
	if dfsPartialOrderGraph(t2, t1, nil, false, weak) {
		return hb.After
	}

	return hb.Concurrent
}

// Pass the partial order graph using dfs
// Store all visited nodes
//
// Parameter:
//   - start trace.TraceElement: element to start from
//   - end trace.TraceElement: if not non, stop when end is reached
//   - visited map[int]bool: traceID of all visited nodes
//   - inverted bool: If false, find all nodes that can be reached from start,
//     if true, find all nodes from which start can be reached
//   - weak bool: if true, use weak partial order
//
// Returns:
//   - bool: if end is nil, return if end has been reached, otherwise return true
func dfsPartialOrderGraph(start, end trace.Element, reachable map[int]bool,
	inverted, weak bool) bool {
	if end != nil && start.GetTraceID() == end.GetTraceID() {
		return true
	}

	if reachable == nil {
		reachable = make(map[int]bool)
	}

	stack := []trace.Element{start}

	var g *poGraph
	if weak {
		if inverted {
			g = &poWeakInverted
		} else {
			g = &poWeak
		}
	} else {
		if inverted {
			g = &poInverted
		} else {
			g = &po
		}
	}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable[curr.GetTraceID()] {
			continue
		}

		reachable[curr.GetTraceID()] = true

		if end != nil && start.GetTraceID() == end.GetTraceID() {
			return true
		}

		for child := range g.getChildren(curr) {
			if child == nil || reflect.ValueOf(child).IsNil() {
				continue
			}
			if !reachable[child.GetTraceID()] {
				stack = append(stack, child)
			}
		}
	}

	return end == nil
}
