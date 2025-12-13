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

import (
	"advocate/analysis/baseA"
	"advocate/fuzzing/baseF"
	"advocate/fuzzing/equivalence"
	"advocate/utils/log"
)

// CreateMutations will create mutations based on the predictive analysis.
// This includes both situations, where the predictive analysis directly
// finds a potential bug and where we use the happens before analysis
// to guide the mutation in a better direction.
func CreateMutations() {
	numberMuts = 0
	traceID++

	// add new original trace to equivalence
	minTrace := equivalence.TraceEqFromTrace(&baseA.MainTrace)
	equivalence.AddOrig(minTrace, traceID)

	predictive()
	random()
}

// predictive runs the predictive analysis and creates new runs based on
// the rewritten traces that should contain the detected bugs.
// We create all possible mutations, even if this would exceed maxNumberOfMutsPerRun
func predictive() {
	// TODO: continue
}

// random creates random mutation, if the number of predictive and guided mutations
// has not reached the max number of mutations per run
func random() {
	numTry := 0

	constraint := startConstraint(maxNumberConstraints, lengthConstraint)

	for _, c := range constraint {
		for numberMuts < maxNumberOfMutsPerConst || numTry > maxTries {
			numTry++

			mutatedConstr := baseF.Mutate(c, -1, nil, nil)

			for _, ch := range mutatedConstr {
				minTrace := equivalence.TraceEqFromConstraint(ch)

				if equivalence.HasEquivalent(minTrace, traceID) {
					continue
				}

				firstMut := baseF.NumberFuzzingRuns <= 1 && numberMuts == 0
				_, err := baseF.WriteMutConstraint(ch, firstMut)
				if err != nil {
					log.Error("Error in writing mutation: ", err.Error())
				}
				numberMuts++
			}
		}
	}
}
