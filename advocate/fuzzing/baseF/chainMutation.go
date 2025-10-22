// Copyright (c) 2025 Erik Kassubek
//
// File: chainMutation.go
// Brief: Mutations for chains
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package baseF

import (
	"advocate/analysis/hb/concurrent"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/settings.go"
	"maps"
	"math/rand/v2"
)

// Create the mutations for a GoPie chain
//
// Parameter:
//   - c chain: The scheduling chain to mutate
//   - energy int: Determines how many mutations are created
//   - rel1 map[trace.Element]map[trace.Element]struct{}: the rel1 info for goPie
//   - rel2 map[trace.Element]map[trace.Element]struct{}: the rel2 info for goPie
//
// Returns:
//   - map[string]chain: Set of mutations
func Mutate(c Chain, energy int, rel1, rel2 map[trace.Element]map[trace.Element]struct{}) map[string]Chain {
	if energy > 100 {
		energy = 100
	}

	bound := settings.GoPieMaxSCLength
	if flags.FuzzingMode == GoPie {
		bound = 3
	}

	mutateBound := settings.ChainMutabound

	// in the original goPie, the fuzzing bound is 3
	if flags.FuzzingMode == GoPie {
		bound = 3
	}

	res := make(map[string]Chain)

	if energy == 0 {
		return res
	}

	if c.Len() == 0 {
		return res
	}

	res[c.ToString()] = c

	maxMutPerStep := 6
	if flags.FuzzingMode == GoPie {
		maxMutPerStep = -1
	}

	for {
		noNew := false
		for _, ch := range res {
			tSet := make(map[string]Chain, 0)

			// Rule 1 -> abridge
			if ch.Len() >= 2 && rand.Int()%2 == 1 {
				newCh1, newCh2 := abridge(ch)
				tSet[newCh1.ToString()] = newCh1
				tSet[newCh2.ToString()] = newCh2
			}

			// Rule 2 -> flip (not in original implementation, not in GoPie,
			// but in GoCR and GoCRHB)
			if flags.FuzzingMode != GoPie {
				if ch.Len() >= 2 && rand.Int()%2 == 1 {
					newChs := flip(ch)

					if maxMutPerStep != -1 {
						shuffle(&newChs, maxMutPerStep)
					}

					for _, newCh := range newChs {
						tSet[newCh.ToString()] = newCh
					}
				}
			}

			// Rule 3 -> substitute
			// if ch.len() <= bound && rand.Int()%2 == 1 {
			if rand.Int()%2 == 1 {
				newChs := substitute(ch, rel1)

				if maxMutPerStep != -1 {
					shuffle(&newChs, maxMutPerStep)
				}

				for _, newCh := range newChs {
					tSet[newCh.ToString()] = newCh
				}
			}

			// Rule 4 -> augment
			if ch.Len() <= bound && rand.Int()%2 == 1 {
				newChs := augment(c, rel2)

				if maxMutPerStep != -1 {
					shuffle(&newChs, maxMutPerStep)
				}

				for _, newCh := range newChs {
					tSet[newCh.ToString()] = newCh
				}
			}

			for k, v := range tSet {
				res[k] = v
			}
		}

		if noNew {
			break
		}

		if len(res) > mutateBound {
			break
		}

		if FuzzingModeGoPie && (rand.Int()%200) < energy {
			break
		}
	}

	// mutates selects
	for _, mut := range res {
		new := mut.MutSelect()
		maps.Copy(res, new)
	}

	return res
}

// Abridge mutation. This creates two new mutations, where either the
// first or the last element has been removed
//
// Parameter:
//   - c chain: the chain to mutate
//
// Returns:
//   - chain: a copy of the chain with the first element removed
//   - chain: a copy of the chain with the last element removed
func abridge(c Chain) (Chain, Chain) {
	ncHead := c.Copy()
	ncHead.RemoveHead()
	ncTail := c.Copy()
	ncTail.RemoveTail()

	return ncHead, ncTail
}

// Flip mutations. For each pair of neighboring elements in the chain, a
// new chain is created where those two elements are flipped
//
// Parameter:
//   - c chain: the chain to mutate
//
// Returns:
//   - []chain: the list of mutated chains
func flip(c Chain) []Chain {
	res := make([]Chain, 0)

	// switch each element with the next element
	// for each flip create a new chain
	for i := 0; i < c.Len()-1; i++ {
		nc := c.Copy()
		suc := nc.Swap(i, i+1)
		if suc {
			res = append(res, nc)
		}
	}
	return res
}

// Substitute mutations. For each element create new mutations, where this
// element is replaced by an element with another trace element from the same
// routine. This new element can not be in the chain already
//
// Parameter:
//   - c chain: the chain to mutate
//   - rel1 map[trace.Element]map[trace.Element]struct{}: the rel1 info for goPie
//
// Returns:
//   - []chain: the list of mutated chains
func substitute(c Chain, rel1 map[trace.Element]map[trace.Element]struct{}) []Chain {
	res := make([]Chain, 0)

	for i, elem := range c.Elems {
		for rel := range rel1[elem] {
			if res != nil && !c.Contains(rel) {
				nc := c.Copy()
				nc.Replace(i, rel)
				res = append(res, nc)
			}
		}
	}

	return res
}

// Augment mutations. For each element in the Rel2 set of the last element
// in the chain that is not in the chain already, created a new chain where
// this element is added at the end.
//
// Parameter:
//   - c chain: the chain to mutate
//   - rel2 map[trace.Element]map[trace.Element]struct{}: the rel2 info for goPie
//
// Returns:
//   - []chain: the list of mutated chains
func augment(c Chain, rel2 map[trace.Element]map[trace.Element]struct{}) []Chain {
	res := make([]Chain, 0)

	if UseHBInfoFuzzing {
		concurrent := concurrent.GetConcurrent(c.LastElem(), true, false, settings.SameElementTypeInSC, true)
		for _, elem := range concurrent {
			if c.Contains(elem) {
				continue
			}

			nc := c.Copy()
			nc.Add(elem)
			res = append(res, nc)
		}
	} else {
		for rel := range rel2[c.LastElem()] {
			if c.Contains(rel) {
				continue
			}

			nc := c.Copy()
			nc.Add(rel)
			res = append(res, nc)
		}
	}

	return res
}

func shuffle(c *[]Chain, n int) {
	if len(*c) <= n {
		return
	}

	rand.Shuffle(len(*c), func(i, j int) {
		(*c)[i], (*c)[j] = (*c)[j], (*c)[i]
	})

	*c = (*c)[:n]
}
