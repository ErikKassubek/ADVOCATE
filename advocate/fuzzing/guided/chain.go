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
	"advocate/fuzzing/baseF"
	"math/rand"
)

// TODO: get good chain

// Start chain returns a chain of given length, consisting of consecutive
// elements from the trace
//
// Parameter:
//   - length int: number of elements in the chain
//
// Returns:
//   - baseF.Chain: a new chain consisting of consecutive elements from the chain
func startChain(length int) baseF.Chain {
	l := baseA.MainTrace.GetNumberElements()
	start := rand.Intn(max(0, l-length)) // trace index of first element in chain

	c := baseF.NewChain()

	section := baseA.MainTrace.GetTraceSection(start, start+length)

	if len(section) == 0 {
		return c
	}

	c.Add(section...)

	return c
}
