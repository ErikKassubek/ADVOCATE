// Copyright (c) 2025 Erik Kassubek
//
// File: constraint.go
// Brief: Constraints for guided fuzzing
//
// Author: Erik Kassubek
// Created: 2025-10-21
//
// License: BSD-3-Clause

package roc

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb/concurrent"
	"advocate/fuzzing/baseF"
	"advocate/trace"
	"advocate/utils/settings.go"
	"math/rand"
)

// StartConstraint returns a constraint of given length, consisting of consecutive
// elements from the trace
//
// Parameter:
//   - num int: number of constraints tob be created
//   - length int: max number of elements in the constraint
//
// Returns:
//   - []baseF.Constraint: a set of new constraint consisting of consecutive elements from the constraint
func startConstraint(num, length int) []baseF.Constraint {
	res := make([]baseF.Constraint, 0)

	traces := baseA.MainTrace.GetTraces()

	if len(traces) == 0 {
		return res
	}

	top := []baseF.ElemWithQual{}

	alreadyAdded := make(map[int]struct{})

	for i := 0; i < 1000; i++ {
		key := rand.Intn(len(traces)) + 1
		trace := traces[key]
		if len(trace) == 0 {
			continue
		}

		ind := rand.Intn(len(trace))
		elem := trace[ind]

		if _, ok := alreadyAdded[elem.GetTPost()]; ok {
			continue
		}
		alreadyAdded[elem.GetTPost()] = struct{}{}

		if !baseF.CanBeAddedToConstraint(elem) {
			continue
		}

		sameElem := true
		if concurrent.GetNumberConcurrent(elem, sameElem, settings.SameElementTypeInSC, false) == 0 {
			continue
		}

		q := baseF.Quality(elem)

		e := baseF.ElemWithQual{Elem: elem, Quality: q}

		// find the num with the best quality
		inserted := false
		for i, v := range top {
			if e.Quality > v.Quality {
				top = append(top[:i+1], top[i:]...)
				top[i] = e
				inserted = true
				break
			}
		}
		if !inserted && len(top) < num {
			top = append(top, e)
		}
	}

	if len(top) == 0 {
		return res
	}

	if len(top) > num {
		top = top[:num]
	}

	for _, e := range top {
		c := baseF.NewConstraint()
		c.Add(e.Elem)

		for i := 0; i < length; i++ {
			posNext := concurrent.GetConcurrent(c.LastElem(), true, true, settings.SameElementTypeInSC, true)
			if len(posNext) == 0 {
				posNext = concurrent.GetConcurrent(c.LastElem(), true, false, settings.SameElementTypeInSC, true)
				if len(posNext) == 0 {
					break
				}
			}

			concToAll := make([]trace.Element, 0)

			for _, next := range posNext {
				isPos := true
				for _, e := range c.Elems {
					if !concurrent.IsConcurrentWeak(next, e) {
						isPos = false
						break
					}
				}
				if isPos {
					concToAll = append(concToAll, next)
				}
			}

			if len(concToAll) == 0 {
				break
			}

			next := concToAll[rand.Intn(len(concToAll))]
			c.Add(next)
		}

		res = append(res, c)
	}

	return res
}
