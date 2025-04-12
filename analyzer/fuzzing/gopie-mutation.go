// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-mutation.go
// Brief: Mutations for gopie
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package fuzzing

import (
	"math/rand/v2"
)

const (
	bound       = 3
	mutateBound = 128
	maxNoNew    = 5
)

// Create the mutations for a GoPie chain
//
// Parameter:
//   - c (chain): The scheduling chain to mutate
//   - energy (int): Determines how many mutations are created
//
// Returns:
//   - map[string]chain: Set of mutations
func mutate(c chain, energy int) map[string]chain {
	if energy > 100 {
		energy = 100
	}

	res := make(map[string]chain)

	if c.len() == 0 {
		return res
	}

	res[c.toString()] = c

	countNoNew := 0

	for {
		noNew := false
		for _, ch := range res {
			tset := make(map[string]chain, 0)

			// Rule 1 -> abridge
			if ch.len() >= 2 {
				newCh1, newCh2 := abridge(ch)
				tset[newCh1.toString()] = newCh1
				tset[newCh2.toString()] = newCh2
			}

			// Rule 2 -> flip (not in original implementation)
			if ch.len() >= 2 {
				newChs := flip(ch)
				for _, newCh := range newChs {
					tset[newCh.toString()] = newCh
				}
			}

			// Rule 3 -> substitute
			if ch.len() <= bound && rand.Int()%2 == 1 {
				newChs := substitute(ch)
				for _, newCh := range newChs {
					tset[newCh.toString()] = newCh
				}
			}

			// Rule 4 -> augment
			if ch.len() <= bound && rand.Int()%2 == 1 {
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
//   - c (chain): the chain to mutate
//
// Returns:
//   - chain: a copy of the chain with the first element removed
//   - chain: a copy of the chain with the last element removed
func abridge(c chain) (chain, chain) {
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
//   - c (chain): the chain to mutate
//
// Returns:
//   - []chain: the list of mutated chains
func flip(c chain) []chain {
	res := make([]chain, 0)

	// switch each element with the next element
	// for each flip create a new chain
	for i := 0; i < c.len()-1; i++ {
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
//   - c (chain): the chain to mutate
//
// Returns:
//   - []chain: the list of mutated chains
func substitute(c chain) []chain {
	res := make([]chain, 0)

	for i, elem := range c.elems {
		for _, rel := range elem.GetRel1() {
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
//   - c (chain): the chain to mutate
//
// Returns:
//   - []chain: the list of mutated chains
func augment(c chain) []chain {
	res := make([]chain, 0)

	rels := c.lastElem().GetRel2()
	for _, rel := range rels {
		if c.contains(rel) {
			continue
		}

		nc := c.copy()
		nc.add(rel)
		res = append(res, nc)
	}

	return res
}
