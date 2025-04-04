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
	BOUND       = 3
	MUTATEBOUND = 128
)

/*
 * Create the mutations for a GoPie chain
 * Args:
 * 	c (chain): The scheduling chain to mutate
 * 	energy (int): Determines how many mutations are created
 * Returns:
 * 	map[string]chain: Set of mutations
 */
func mutate(c chain, energy int) map[string]chain {
	if energy > 100 {
		energy = 100
	}

	res := make(map[string]chain)

	if c.len() == 0 {
		return res
	}

	res[c.toString()] = c

	for {
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
			if ch.len() <= BOUND && rand.Int()%2 == 1 {
				newChs := substitute(ch)
				for _, newCh := range newChs {
					tset[newCh.toString()] = newCh
				}
			}

			// Rule 4 -> augment
			if ch.len() <= BOUND && rand.Int()%2 == 1 {
				newChs := augment(c)
				for _, newCh := range newChs {
					tset[newCh.toString()] = newCh
				}
			}

			for k, v := range tset {
				res[k] = v
			}
		}

		if len(res) > MUTATEBOUND {
			break
		}

		if (rand.Int() % 200) < energy {
			break
		}
	}

	return res
}

func abridge(c chain) (chain, chain) {
	ncHead := c.copy()
	ncHead.removeHead()
	ncTail := c.copy()
	ncTail.removeTail()

	return ncHead, ncTail
}

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
