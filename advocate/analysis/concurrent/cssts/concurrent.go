// Copyright (c) 2025 Erik Kassubek
//
// File: concurrent.go
// Brief: Find concurrent elements for a start element
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

// For a given element, return concurrent events
// Parameter:
//   - elem trace.TraceElem: the element to search for
//   - all bool: if true, return all concurrent events, otherwise return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//
// Returns:
//   - []trace.TraceElement: the concurrent element(s)
func GetConcurrentCSSTAllPairs(elem trace.Element, all bool, sameElem bool) []trace.Element {
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

			if !valid(e) {
				continue
			}

			if isConcurrent(elem, e) {
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
//
// Returns:
//   - bool: true if the elements are concurrent, false otherwise
func isConcurrent(elem1, elem2 trace.Element) bool {
	r1, i1 := elem1.GetTraceIndex()
	r2, i2 := elem2.GetTraceIndex()
	return !Csst.reachable(types.Pair[int, int]{X: r1, Y: i1}, types.Pair[int, int]{X: r2, Y: i2}) &&
		!Csst.reachable(types.Pair[int, int]{X: r2, Y: i2}, types.Pair[int, int]{X: r1, Y: i1})
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
func GetConcurrentCSST(elem trace.Element, all bool, sameElem bool) []trace.Element {
	reachableFromN := types.NewSet[types.Pair[int, int]]()
	reachableToN := types.NewSet[types.Pair[int, int]]()

	rout, ind := elem.GetTraceIndex()
	elemInd := types.NewPair(rout, ind)

	dfsCSST(elemInd, reachableFromN, false)
	dfsCSST(elemInd, reachableToN, true)

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

// Pass the partial order graph using dfs
// Store all visited nodes
//
// Parameter:
//   - start trace.TraceElement: element to start from
//   - reachable map[int]bool: traceID of all visited nodes
//   - inverted bool: If false, find all nodes that can be reached from start,
//     if true, find all nodes from which start can be reached
func dfsCSST(from types.Pair[int, int], reachable types.Set[types.Pair[int, int]], inverted bool) types.Set[types.Pair[int, int]] {
	visited := types.NewSet[types.Pair[int, int]]()

	queue := []types.Pair[int, int]{from}

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
			if current.Y+1 < Csst.getChainLength(current.X) {

				next := types.NewPair(current.X, current.Y+1)
				queue = append(queue, next)
			}
		}

		// Traverse all cross-chain successors
		succs := make([]int, 0)
		if inverted {
			succs = CsstInverted.getSuccessor1(current)
		} else {
			succs = Csst.getSuccessor1(current)
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
