// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-sc.go
// Brief: scheduling Chains for GoPie
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package gopie

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb/concurrent"
	"advocate/fuzzing/baseF"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/settings.go"
	"math"
	"math/rand"
)

// startChains returns a slice of chain consisting of a
// pair of operations that are in a rel2 relation
//
// Parameter:
//   - num int: number of chains to return
//
// Returns:
//   - the chain, or an empty chain if no pair exists
func startChains(num int) []baseF.Chain {
	res := make([]baseF.Chain, 0)

	if baseF.UseHBInfoFuzzing {
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

			if !CanBeAddedToChain(elem) {
				continue
			}

			if concurrent.GetNumberConcurrent(elem, sameElem, settings.SameElementTypeInSC, false) == 0 {
				continue
			}

			q := quality(elem)

			e := baseF.ElemWithQual{elem, q}

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

			// Trim if longer than n
			if len(top) > num {
				top = top[:num]
			}
		}

		if len(top) == 0 {
			return res
		}

		rounds := settings.GoPieMaxSCLength
		if !settings.GoPieMaxSCLengthSet {
			rounds = 1
		}
		for i := 0; i < rounds; i++ {
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
					c := baseF.NewChain()
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
	} else {
		// start with two random elements in rel2
		i := 0
		for elem1, rel := range rel2 {
			for elem2 := range rel {
				c := baseF.NewChain()
				c.Add(elem1, elem2)
				res = append(res, c)
				i++
				if i > num {
					return res
				}
			}
		}
	}

	return res
}

// quality calculates how fit for mutation an element is
// This is based on how many times was an operation called on the same element
// and how many concurrent operation has the operations
//
// Parameters:
//   - elem trace.Element: the element to check for
//
// Returns:
//   - float64: the quality
func quality(elem trace.Element) float64 {
	w1 := 0.2
	w2 := 0.3
	w3 := 0.5

	numberOps, _ := baseA.GetOpsPerID(elem.GetObjId())
	numberConcurrentTotal := concurrent.GetNumberConcurrent(elem, false, false, true)
	numberConcurrentSame := concurrent.GetNumberConcurrent(elem, true, false, true)

	if numberConcurrentSame == 0 && numberConcurrentTotal == 0 {
		return 0
	}

	q := w1*math.Log1p(float64(numberOps)) +
		w2*float64(numberConcurrentSame)/float64(numberConcurrentTotal+1) +
		w3*math.Log1p(float64(numberConcurrentTotal))

	return q * ((rand.Float64() * 0.2) - 0.1)
}

// CanBeAddedToChain decides if an element can be added to a scheduling chain
// For GoPie without improvements (!useHBInfoFuzzing) those are only mutex and channel (incl. select)
// With improvements those are all not ignored fuzzing elements
//
// Parameter:
//   - elem analysis.TraceElement: Element to check
//
// Returns:
//   - true if it can be added to a scheduling chain, false otherwise
func CanBeAddedToChain(elem trace.Element) bool {
	t := elem.GetType(false)
	if flags.FuzzingMode == baseF.GoPie {
		// for standard GoPie, only mutex, channel and select operations are considered
		return t == trace.Mutex || t == trace.Channel || t == trace.Select
	}

	return t != trace.Atomic && !baseF.IgnoreFuzzing(elem, true)
}
