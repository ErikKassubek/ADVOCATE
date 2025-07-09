// Copyright (c) 2025 Erik Kassubek
//
// File: incrementalCSSTs.go
// Brief: Implementation of IncrementalCSSTs.java from CSST paper (see csst.go)
//
// Author: Erik Kassubek
// Created: 2025-07-02
//
// License: BSD-3-Clause

// CSST is based on
// Hünkar Can Tunç, Ameya Prashant Deshmukh, Berk Cirisci, Constantin Enea,
// and Andreas Pavlogiannis. 2024. CSSTs: A Dynamic Data Structure for Partial
// Orders in Concurrent Execution Analysis. In Proceedings of the 29th ACM
// International Conference on Architectural Support for Programming Languages
// and Operating Systems, Volume 3 (ASPLOS '24), Vol. 3. Association for
// Computing Machinery, New York, NY, USA, 223–238.
// https://doi.org/10.1145/3620666.3651358

package cssts

import (
	"advocate/utils/types"
	"math"
)

type IncrementalCSST struct {
	width   int
	lengths []int
	ssts    []map[int]sparseSegmentTree
}

func NewIncrementalCSST(lengths []int) IncrementalCSST {
	iCSST := IncrementalCSST{
		len(lengths), lengths, make([]map[int]sparseSegmentTree, 0),
	}

	for i := 0; i < iCSST.width; i++ {
		iCSST.ssts = append(iCSST.ssts, make(map[int]sparseSegmentTree))
		for j := 0; j < iCSST.width; j++ {
			if i != j {
				if iCSST.lengths[i] == 0 {
					iCSST.ssts[i][j] = newSparseSegmentTree1(1)
				} else {
					iCSST.ssts[i][j] = newSparseSegmentTree1(iCSST.lengths[i])
				}
			}
		}
	}

	return iCSST
}

func (iCSST *IncrementalCSST) getSuccessor1(p types.Pair[int, int]) []int {
	s := make([]int, iCSST.width)
	for i := 0; i < iCSST.width; i++ {
		s[i] = iCSST.getSuccessor2(p, i)
	}

	return s
}

func (iCSST *IncrementalCSST) getSuccessor2(p types.Pair[int, int], i int) int {
	if p.X == i {
		if p.Y < iCSST.lengths[p.X]-1 {
			return p.Y + 1
		}
		return -1
	}

	v := iCSST.ssts[p.X][i].sumRange1(p.Y, iCSST.getChainLength(p.X)-1)
	if v < math.MaxInt {
		return v
	}
	return -1
}

func (iCSST *IncrementalCSST) getPredecessor1(p types.Pair[int, int]) []int {
	s := make([]int, iCSST.width)
	for i := 0; i < iCSST.width; i++ {
		s[i] = iCSST.getPredecessor2(p, i)
	}

	return s
}

func (iCSST *IncrementalCSST) getPredecessor2(p types.Pair[int, int], i int) int {
	if p.X == i {
		if p.Y > 0 {
			return p.Y - 1
		}
		return -1
	}

	v := iCSST.ssts[i][p.X].argMin1(p.Y)
	if v > math.MinInt {
		return v
	}
	return -1
}

func (iCSST *IncrementalCSST) reachable(from, to types.Pair[int, int]) bool {
	if from.X == to.X && from.Y == to.Y {
		return true
	}

	v := iCSST.getSuccessor2(from, to.X)
	return v >= 0 && v <= to.Y
}

func (iCSST *IncrementalCSST) addSuccessor(from, to types.Pair[int, int]) {
	if from.X != to.X {
		iCSST.ssts[from.X][to.X].update1(from.Y, to.Y)
	}
}

func (iCSST *IncrementalCSST) AddEdge(from, to types.Pair[int, int]) types.Set[types.Pair[types.Pair[int, int], types.Pair[int, int]]] {
	addedEdges := types.NewSet[types.Pair[types.Pair[int, int], types.Pair[int, int]]]()

	if iCSST.reachable(from, to) {
		return addedEdges
	}

	q := types.Stack[types.Pair[types.Pair[int, int], types.Pair[int, int]]]{}

	q.Push(types.Pair[types.Pair[int, int], types.Pair[int, int]]{X: from, Y: to})

	for !q.IsEmpty() {
		e := q.Pop()
		f := e.X
		t := e.Y

		if iCSST.reachable(f, t) {
			continue
		}

		iCSST.addSuccessor(f, t)
		addedEdges.Add(e)

		succ := iCSST.getSuccessor1(t)
		pred := iCSST.getPredecessor1(f)

		for i := 0; i < iCSST.width; i++ {
			if i != f.X && i != t.X {
				tt := types.NewPair(i, succ[i])
				if succ[i] >= 0 {
					q.Push(types.NewPair(f, tt))
				}
				ff := types.NewPair(i, pred[i])
				if pred[i] >= 0 {
					q.Push(types.NewPair(ff, t))
				}
			}
		}
	}

	return addedEdges
}

func (iCSST *IncrementalCSST) getChainLength(i int) int {
	return iCSST.lengths[i]
}
