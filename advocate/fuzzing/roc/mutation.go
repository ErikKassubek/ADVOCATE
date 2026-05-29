// Copyright (c) 2026 Erik Kassubek
//
// File: mutation.go
// Brief: ROC Creation and mutation
//
// Author: Erik Kassubek
// Created: 2025-10-04
//
// License: BSD-3-Clause

package roc

import (
	"advocate/analysis/baseA"
	"advocate/fuzzing/baseF"
	"advocate/fuzzing/gfuzz"
	"advocate/utils/log"
	"math/rand"
)

// CreateMutations will create randomized mutations.
// This includes both situations, where the predictive analysis directly
// finds a potential bug and where we use the happens before analysis
// to guide the mutation in a better direction.
// Mutations based on predictive analysis are automatically created when a potential
// bug is detected
func CreateMutations() {
	numberMuts = 0

	// drop runs where the mutation could not be fully satisfied
	if baseA.GetTimeoutHappened(false) {
		return
	}

	traceID++

	// select based mutations
	gfuzz.CreateMutations(true)

	// add new original trace to equivalence
	// minTrace := equivalence.TraceEqFromTrace(&baseA.MainTrace)
	// equivalence.AddOrig(minTrace, traceID)

	numTry := 0

	constraint := startConstraint(maxNumberConstraints, lengthConstraint)
	for _, c := range constraint {
		for numberMuts < maxNumberOfMutsPerConst || numTry > maxTries {
			mutatedConstr := baseF.Mutate(c, -1, nil, nil)

			for _, cr := range mutatedConstr {
				baseF.TotalRuns++
				numTry++

				if numTry > maxTries {
					return
				}

				// minTrace := equivalence.TraceEqFromConstraint(ch)

				// if minTrace.IllFormedImpossible {
				// 	baseF.IllFormed++
				// 	continue
				// }

				if isEquivalent(cr) {
					baseF.Equiv++
					if rand.Float64() < propToSkipEquiv {
						continue
					}
				}
				baseF.TotalRuns++

				firstMut := baseF.NumberFuzzingRuns <= 1 && numberMuts == 0
				_, err := baseF.WriteMutConstraint(cr, firstMut)
				if err != nil {
					log.Error("Error in writing mutation: ", err.Error())
				}
				numberMuts++
			}
		}
	}
}
