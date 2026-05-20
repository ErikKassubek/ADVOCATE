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
	"gocdr/fuzzing/baseF"
)

// startChains returns a slice of chain consisting of a
// pair of operations that are in a rel2 relation
//
// Parameter:
//   - num int: number of chains to return
//
// Returns:
//   - the chain, or an empty chain if no pair exists
func startChains(num int) []baseF.Constraint {
	res := make([]baseF.Constraint, 0)

	// start with two random elements in rel2
	i := 0
	for elem1, rel := range rel2 {
		for elem2 := range rel {
			c := baseF.NewConstraint()
			c.Add(elem1, elem2)
			res = append(res, c)
			i++
			if i > num {
				return res
			}
		}
	}

	return res
}
