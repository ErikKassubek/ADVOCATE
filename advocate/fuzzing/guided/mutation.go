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
	"advocate/trace"
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
	minTrace := trace.TraceMinFromTrace(&baseA.MainTrace)
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
	for numberMuts < maxNumberOfMutsPerRun || numTry > maxTries {
		numTry++

		chain := startChain(lengthChain)

		mutatedChains := baseF.Mutate(chain, -1, nil, nil)

		for _, ch := range mutatedChains {
			minTrace := traceMinFromChain(ch)

			if equivalence.HasEquivalent(minTrace, traceID) {
				continue
			}

			firstMut := baseF.NumberFuzzingRuns <= 1 && numberMuts == 0
			_, err := baseF.WriteMutChain(ch, firstMut)
			if err != nil {
				log.Error("Error in writing mutation: ", err.Error())
			}
			numberMuts++
		}
	}
}

// traceMinFromChain creates a trace min from a chain
//
// Parameter:
//   - chain Chain: the chain
func traceMinFromChain(chain baseF.Chain) trace.TraceMin {
	res := trace.NewTraceMin()

	minTPost := chain.ElemWithSmallestTPost().GetTSort()

	traceIter := baseA.MainTrace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		if elem.GetTSort() >= minTPost {
			break
		}

		minElem, val := elem.GetElemMin()
		if val {
			res.AddElem(minElem)
		}
	}

	return res
}
