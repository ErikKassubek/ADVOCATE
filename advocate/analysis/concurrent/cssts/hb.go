// Copyright (c) 2025 Erik Kassubek
//
// File: concurrent.go
// Brief: Get concurrency information
//
// Author: Erik Kassubek
// Created: 2025-07-07
//
// License: BSD-3-Clause

package cssts

import (
	"advocate/analysis/concurrent/hb"
	"advocate/analysis/concurrent/helper"
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/types"
)

// For a given element, return concurrent events
// Parameter:
//   - elem trace.TraceElem: the element to search for
//   - all bool: if true, return all concurrent events, otherwise return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: get based on weak happens before
//
// Returns:
//   - []trace.TraceElement: the concurrent element(s)
func GetConcurrentAllPairs(elem trace.Element, all, sameElem, weak bool) []trace.Element {
	res := make([]trace.Element, 0)
	id := elem.GetID()
	routine := elem.GetRoutine()

	for r, trace := range data.MainTrace.GetTraces() {
		// same routine
		if routine == r {
			continue
		}

		// different routine
		for _, e := range trace {
			if sameElem && e.GetID() != id {
				continue
			}

			if !helper.Valid(e) {
				continue
			}

			if isConcurrent(elem, e, weak) {
				res = append(res, e)
				if !all {
					return res
				}
			}
		}
	}

	return res
}

// Given two trace elements, return if the elements are concurrent
//
// Parameter:
//   - elem1 trace.Element: the first elem
//   - elem2 trace.Element: the second elem
//   - weak bool: get based on weak happens before
//
// Returns:
//   - bool: true if the elements are concurrent, false otherwise
func isConcurrent(elem1, elem2 trace.Element, weak bool) bool {
	hbInfo := GetHappensBefore(elem1, elem2, weak)
	return hbInfo == hb.Concurrent
}

// For a given element, find one or all elements that are concurrent to it
//
// Parameter
//   - elem TraceElement: the element the results should be concurrent with
//   - all bool: if true, return all elements that are concurrent, if false, only return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: get based on weak happens before
//
// Returns
//   - []trace.TraceElement: element(s) that are concurrent to elem
func GetConcurrent(elem trace.Element, all, sameElem, weak bool) []trace.Element {
	reachableFromN := types.NewSet[types.Pair[int, int]]()
	reachableToN := types.NewSet[types.Pair[int, int]]()

	rout, ind := elem.GetTraceIndex()
	elemInd := types.NewPair(rout, ind)

	dfsCSST(elemInd, reachableFromN, weak, false)
	dfsCSST(elemInd, reachableToN, weak, true)

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

			r, i := tElem.GetTraceIndex()
			tElemInd := types.NewPair(r, i)

			if !reachableFromN.Contains(tElemInd) && !reachableToN.Contains(tElemInd) {
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
// csst
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - weak bool: get based on weak happens before
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(t1, t2 trace.Element, weak bool) hb.HappensBefore {
	t1Ind := getIndicesFromTraceElem(t1)
	t2Ind := getIndicesFromTraceElem(t2)

	if weak {
		if CsstWeak.reachable(t1Ind, t2Ind) {
			return hb.Before
		}
		if CsstWeak.reachable(t2Ind, t1Ind) {
			return hb.After
		}
		return hb.Concurrent
	}

	if Csst.reachable(t1Ind, t2Ind) {
		return hb.Before
	}
	if Csst.reachable(t2Ind, t1Ind) {
		return hb.After
	}
	return hb.Concurrent
}

// Pass the partial order graph using dfs
// Store all visited nodes
//
// Parameter:
//   - start trace.TraceElement: element to start from
//   - reachable map[int]bool: traceID of all visited nodes
//   - weak bool: get based on weak happens before
//   - inverted bool: If false, find all nodes that can be reached from start,
//     if true, find all nodes from which start can be reached
func dfsCSST(from types.Pair[int, int], reachable types.Set[types.Pair[int, int]], weak, inverted bool) types.Set[types.Pair[int, int]] {
	visited := types.NewSet[types.Pair[int, int]]()

	queue := []types.Pair[int, int]{from}

	st := &Csst
	stInverted := &CsstInverted
	if weak {
		st = &CsstWeak
		stInverted = &CsstWeakInverted
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited.Contains(current) {
			continue
		}
		visited.Add(current)

		// Skip adding the origin to the reachable set if needed
		if !(current.X == from.X && current.Y == from.Y) {
			reachable.Add(current)
		}

		// Traverse next event on the same chain
		if inverted {
			if current.Y-1 >= 0 {

				next := types.NewPair(current.X, current.Y-1)
				queue = append(queue, next)
			}
		} else {
			if current.Y+1 < st.getChainLength(current.X) {

				next := types.NewPair(current.X, current.Y+1)
				queue = append(queue, next)
			}
		}

		// Traverse all cross-chain successors
		succs := make([]int, 0)
		if inverted {
			succs = stInverted.getSuccessor1(current)
		} else {
			succs = st.getSuccessor1(current)
		}
		for targetChain, succY := range succs {
			if targetChain == current.X || succY == -1 {
				continue
			}
			succ := types.NewPair(targetChain, succY)
			queue = append(queue, succ)
		}
	}

	return reachable
}
