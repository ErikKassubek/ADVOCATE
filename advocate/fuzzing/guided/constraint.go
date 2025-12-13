// Copyright (c) 2025 Erik Kassubek
//
// File: chain.go
// Brief: Chain for guided fuzzing
//
// Author: Erik Kassubek
// Created: 2025-10-21
//
// License: BSD-3-Clause

package guided

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb/concurrent"
	"advocate/fuzzing/baseF"
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

	for i := 0; i < 1000; i++ {
		key := rand.Intn(len(traces)) + 1
		trace := traces[key]
		if len(trace) == 0 {
			continue
		}

		ind := rand.Intn(len(trace))
		elem := trace[ind]

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

	for i := 0; i < length; i++ {
		if len(res) == 0 {
			for _, e := range top {
				posPartner := concurrent.GetConcurrent(e.Elem, true, true, settings.SameElementTypeInSC, true)
				if len(posPartner) == 0 {
					posPartner = concurrent.GetConcurrent(e.Elem, true, false, settings.SameElementTypeInSC, true)
					if len(posPartner) == 0 {
						continue
					}
				}

				partner := posPartner[rand.Intn(len(posPartner))]
				c := baseF.NewConstraint()
				c.Add(e.Elem, partner)
				res = append(res, c)
			}
		} else {
			for i, c := range res {
				lastElem := c.LastElem()
				if lastElem == nil {
					continue
				}

				posNext := concurrent.GetConcurrent(lastElem, true, true, settings.SameElementTypeInSC, true)
				if len(posNext) == 0 {
					continue
				}
				next := posNext[rand.Intn(len(posNext))]
				res[i].Add(next)
			}
		}
	}

	return res
}
