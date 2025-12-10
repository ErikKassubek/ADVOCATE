// Copyright (c) 2025 Erik Kassubek
//
// File: pog.go
// Brief: Given two partial order graphs, determine if they are equivalent
//
// Author: Erik Kassubek
// Created: 2025-12-10
//
// License: BSD-3-Clause

package equivalence

import (
	"advocate/analysis/hb/pog"
	"advocate/utils/types"
	"sort"
)

// areEquivalentPog takes two partial order graphs and determines, if they
// are equivalent.
//
// Parameter:
//   - pog1: partial order graph 1
//   - pog1: partial order graph 2
//
// Returns:
//   - bool: true if the traces are equivalent
func areEquivalentPog(pog1, pog2 pog.PoGraph) bool {
	nodes1 := extractNodes(pog1)
	nodes2 := extractNodes(pog2)

	common := types.MergeListsSort(nodes1, nodes2)

	if len(common) == 0 {
		return true
	}

	tc1 := transitiveClosure(pog1, nodes1)
	tc2 := transitiveClosure(pog2, nodes2)

	for _, u := range common {
		for _, v := range common {
			r1 := reachable(tc1, nodes1, u, v)
			r2 := reachable(tc2, nodes2, u, v)
			if r1 != r2 {
				return false
			}
		}
	}

	return true
}

// extractNodes returns the list of ids in the graph
//
// Parameter:
//   - pog pog.PoGraph: the partial order graph
//
// Returns:
//   - []int: sorted list of idsMergeListsSort
func extractNodes(pog pog.PoGraph) []int {
	m := make(map[int]struct{})
	for u := range pog.DataSimple {
		m[u] = struct{}{}
		for v := range pog.DataSimple[u] {
			m[v] = struct{}{}
		}
	}

	res := make([]int, 0, len(m))
	for id := range m {
		res = append(res, id)
	}

	sort.Ints(res)
	return res
}

// transitiveClosure calculates the transitive closure of the graph
//
// Parameter:
//   - pog pog.PoGraph: the graph
//   - nodes []int: the sorted list of nodes
//
// Returns:
//   - map[int][]uint64: the transitive closure
func transitiveClosure(pog pog.PoGraph, nodes []int) map[int][]uint64 {
	n := len(nodes)

	// create a map from nodeID -> index in [0..n-1]
	idx := make(map[int]int, n)
	for i, id := range nodes {
		idx[id] = i
	}

	// numbers of uint64s needed to store n bits
	w := (n + 63) / 64

	closure := make(map[int][]uint64, n)
	for _, id := range nodes {
		closure[id] = make([]uint64, w)
	}

	// direct edges
	for u := range pog.DataSimple {
		for v := range pog.DataSimple[u] {
			vi := idx[v]
			closure[u][vi/64] |= 1 << (vi % 64)
		}
	}

	// Floyd-Warshall-like bitset propagation
	for _, kID := range nodes {
		kIdx := idx[kID]
		kWord := kIdx / 64
		kMask := uint64(1 << (kIdx % 64))

		for _, iID := range nodes {
			if closure[iID][kWord]&kMask != 0 {
				// OR in closure[k] into closure[i]
				for widx := 0; widx < w; widx++ {
					closure[iID][widx] |= closure[kID][widx]
				}
			}
		}
	}

	return closure
}

// reachable checks if u is reachable from v
//
// Parameter:
//   - closure map[int][]uint64: the closure data
//   - nodes []int: the sorted list of nodes
//   - u int: the start node id
//   - v int: the end node id
func reachable(closure map[int][]uint64, nodes []int, u, v int) bool {
	if _, ok := closure[u]; !ok {
		return false
	}

	idx := make(map[int]int, len(nodes))
	for i, id := range nodes {
		idx[id] = i
	}

	vi := idx[v]

	return (closure[u][vi/64] & (1 << (vi % 64))) != 0
}
