// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-mutation.go
// Brief: Mutations for gopie
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package gopie

import (
	"advocate/analysis/hb/concurrent"
	"advocate/fuzzing/data"
	"advocate/utils/helper"
	"math/rand/v2"
)

const (
	maxNoNew = 5
)

// TODO: limit the number of mutations that can be created from one mutation step

// Create the mutations for a GoPie chain
//
// Parameter:
//   - c chain: The scheduling chain to mutate
//   - energy int: Determines how many mutations are created
//
// Returns:
//   - map[string]chain: Set of mutations
func mutate(c Chain, energy int) map[string]Chain {
	if energy > 100 {
		energy = 100
	}

	bound := helper.GoPieBound
	mutateBound := helper.GoPieMutabound

	res := make(map[string]Chain)

	if energy == 0 {
		return res
	}

	if c.Len() == 0 {
		return res
	}

	res[c.toString()] = c

	countNoNew := 0

	for {
		noNew := false
		for _, ch := range res {
			tset := make(map[string]Chain, 0)

			// Rule 1 -> abridge
			if ch.Len() >= 2 {
				newCh1, newCh2 := abridge(ch)
				tset[newCh1.toString()] = newCh1
				tset[newCh2.toString()] = newCh2
			}

			// Rule 2 -> flip (not in original implementation, not in GoPie,
			// but in GoPie+ and GoPieHB)
			if true || data.FuzzingMode != data.GoPie {
				if ch.Len() >= 2 {
					newChs := flip(ch)
					for _, newCh := range newChs {
						tset[newCh.toString()] = newCh
					}
				}
			}

			// Rule 3 -> substitute
			// if ch.len() <= bound && rand.Int()%2 == 1 {
			if rand.Int()%2 == 1 {
				newChs := substitute(ch)
				for _, newCh := range newChs {
					tset[newCh.toString()] = newCh
				}
			}

			// Rule 4 -> augment
			if ch.Len() <= bound && rand.Int()%2 == 1 {
				newChs := augment(c)
				for _, newCh := range newChs {
					tset[newCh.toString()] = newCh
				}
			}

			lenBefore := len(res)
			for k, v := range tset {
				res[k] = v
			}

			if len(res) == lenBefore { // if no new elements where added
				countNoNew++
				// if no new mutation has been added for maxNoNew rounds, end creation of mutations
				if countNoNew > maxNoNew {
					noNew = true
				}
			}
		}

		if noNew {
			break
		}

		if len(res) > mutateBound {
			break
		}

		if (rand.Int() % 200) < energy {
			break
		}
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
	ncHead := c.copy()
	ncHead.removeHead()
	ncTail := c.copy()
	ncTail.removeTail()

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
		nc := c.copy()
		nc.swap(i, i+1)
		res = append(res, nc)
	}
	return res
}

// Substitute mutations. For each element create new mutations, where this
// element is replaced by an element with another trace element from the same
// routine. This new element can not be in the chain already
//
// Parameter:
//   - c chain: the chain to mutate
//
// Returns:
//   - []chain: the list of mutated chains
func substitute(c Chain) []Chain {
	res := make([]Chain, 0)

	for i, elem := range c.Elems {
		for rel := range rel1[elem] {
			if res != nil && !c.contains(rel) {
				nc := c.copy()
				nc.replace(i, rel)
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
//
// Returns:
//   - []chain: the list of mutated chains
func augment(c Chain) []Chain {
	res := make([]Chain, 0)

	if data.UseHBInfoFuzzing {
		concurrent := concurrent.GetConcurrent(c.lastElem(), true, true, true)
		for _, elem := range concurrent {
			if c.contains(elem) {
				continue
			}

			nc := c.copy()
			nc.add(elem)
			res = append(res, nc)
		}
	} else {
		for rel := range rel2[c.lastElem()] {
			if c.contains(rel) {
				continue
			}

			nc := c.copy()
			nc.add(rel)
			res = append(res, nc)
		}
	}

	return res
}
