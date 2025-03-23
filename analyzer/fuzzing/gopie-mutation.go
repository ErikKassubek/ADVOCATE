// Copyright (c) 2025 Erik Kassubek
//
// File: gopie-mutation.go
// Brief: Mutations for gopie
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package fuzzing

import "math/rand/v2"

const (
	BOUND       = 3
	MUTATEBOUND = 128
)

func mutate(c chain, energy int) map[string]chain {
	// TODO: energy
	if energy > 100 {
		energy = 100
	}

	if c.len() == 0 {
		// TODO: implement
	}

	set := make(map[string]chain)

	set[c.toString()] = c

	for {
		for _, ch := range set {
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
			if ch.len() <= BOUND {
				if rand.Int()%2 == 1 {
					newChs := substitute(ch)
					for _, newCh := range newChs {
						tset[newCh.toString()] = newCh
					}
				}
			}

			// Rule 4 -> augment
			if ch.len() <= BOUND {
				if rand.Int()%2 == 1 {
					newChs := augment(c)
					for _, newCh := range newChs {
						tset[newCh.toString()] = newCh
					}
				}
			}

			for k, v := range tset {
				set[k] = v
			}
		}
		if len(set) > MUTATEBOUND {
			break
		}

		if (rand.Int() % 200) < energy {
			break
		}
	}

	return set
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
	// TODO: implement
	return res
}

func augment(c chain) []chain {
	res := make([]chain, 0)
	// TODO: implement
	return res
}
