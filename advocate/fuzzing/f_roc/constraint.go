// Copyright (c) 2025 Erik Kassubek
//
// File: constraint.go
// Brief: Constraints for guided fuzzing
//
// Author: Erik Kassubek
// Created: 2025-10-21
//
// License: BSD-3-Clause

package f_roc

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_concurrent"
	"advocate/fuzzing/f_base"
	"advocate/trace"
	"advocate/utils/settings"
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
func startConstraint(num, length int) []f_base.Constraint {
	res := make([]f_base.Constraint, 0)

	traces := a_base.MainTrace.GetTraces()

	if len(traces) == 0 {
		return res
	}

	top := []f_base.ElemWithQual{}

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

		if !f_base.CanBeAddedToConstraint(elem) {
			continue
		}

		sameElem := true
		if a_concurrent.GetNumberConcurrent(elem, sameElem, settings.SameElementTypeInSC, false) == 0 {
			continue
		}

		q := f_base.Quality(elem)

		e := f_base.ElemWithQual{Elem: elem, Quality: q}

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
		c := f_base.NewConstraint()
		c.Add(e.Elem)

		for i := 0; i < length; i++ {
			posNext := a_concurrent.GetConcurrent(c.LastElem(), true, true, settings.SameElementTypeInSC, true)
			if len(posNext) == 0 {
				posNext = a_concurrent.GetConcurrent(c.LastElem(), true, false, settings.SameElementTypeInSC, true)
				if len(posNext) == 0 {
					break
				}
			}

			concToAll := make([]trace.Element, 0)

			for _, next := range posNext {
				isPos := true
				for _, e := range c.Elems {
					if !a_concurrent.IsConcurrentWeak(next, e) {
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
