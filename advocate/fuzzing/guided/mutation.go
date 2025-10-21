// Copyright (c) 2025 Erik Kassubek
//
// File: partialOrder.go
// Brief: Fuzzing using the partial order
//
// Author: Erik Kassubek
// Created: 2025-10-04
//
// License: BSD-3-Clause

package guided

var numberMuts = 0

// CreateMutations will create mutations based on the predictive analysis.
// This includes both situations, where the predictive analysis directly
// finds a potential bug and where we use the happens before analysis
// to guide the mutation in a better direction.
func CreateMutations() {
	numberMuts = 0
	predictive()
	guided()
	random()
}

// predictive runs the predictive analysis and creates new runs based on
// the rewritten traces that should contain the detected bugs.
// We create all possible mutations, even if this would exceed maxNumberOfMutsPerRun
func predictive() {
	// analysis.RunAnalysis(false)
	// TODO: continue
}

// guided tries to create interesting runs based on the happens before info
// even if the predictive analysis did not directly indicate a bug
func guided() {
	for numberMuts < maxNumberOfMutsPerRun {
		// TODO: implement
	}
}

// random creates random mutation, if the number of predictive and guided mutations
// has not reached the max number of mutations per run
func random() {
	for numberMuts < maxNumberOfMutsPerRun {
		// TODO: implement
	}
}
