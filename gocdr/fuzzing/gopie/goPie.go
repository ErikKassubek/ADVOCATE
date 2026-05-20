// Copyright (c) 2025 Erik Kassubek
//
// File: goPie.go
// Brief: Main file for goPie fuzzing
//
// Author: Erik Kassubek
// Created: 2025-03-22
//
// License: BSD-3-Clause

package gopie

import (
	"gocdr/analysis/baseA"
	"gocdr/fuzzing/baseF"
	"gocdr/utils/flags"
	"gocdr/utils/log"
	"gocdr/utils/settings"
	"math"
)

const sameElem = true

// CreateMutations create new mutations for GoPie
//
// Parameter:
//   - mutNumber int: number of the mutation file
//   - error
func CreateMutations(mutNumber int) error {
	mutations := make(map[string]baseF.Constraint)
	specMutations := make(map[string]baseF.Constraint) // special mutations that should be run first

	// Original GoPie does not mutate all possible scheduling chains
	// If no SC is given, it creates a new one consisting of two random
	// operations that are in rel2 relation. Otherwise it always mutates the
	// original SC, not newly recorded once
	SchedulingChains = []baseF.Constraint{}
	if flags.FuzzingMode == baseF.GoPie {
		if c, ok := baseF.ChainFiles[mutNumber]; ok {
			c.Old = true
			SchedulingChains = []baseF.Constraint{c}
		}
	}

	if flags.FuzzingMode != baseF.GoPie || len(SchedulingChains) == 0 {
		sc := startChains(maxSCStart)
		for _, c := range sc {
			if c.Len() != 0 {
				SchedulingChains = append(SchedulingChains, c)
			}
		}
	}

	energy := getEnergy()

	log.Infof("Mutate %d scheduling chains", len(SchedulingChains))

	for _, sc := range SchedulingChains {
		muts := baseF.Mutate(sc, energy, rel1, rel2)

		for key, mut := range muts {
			if flags.FuzzingMode != baseF.GoPie && mut.Len() <= 1 {
				NumberTotalMuts++
				continue
			}
			if _, ok := allGoPieMutations[key]; flags.FuzzingMode == baseF.GoPie || !ok {
				// only add if not invalidated by hb
				mutations[key] = mut

				NumberTotalMuts++
				allGoPieMutations[key] = struct{}{}
			} else if flags.FuzzingMode == baseF.GoPie && !ok {
				NumberDoubleMuts++
			}
		}
	}

	if len(specMutations) > 0 {
		log.Infof("Write %d special mutation to file", len(specMutations))
	}

	if baseF.MaxNumberRuns > 0 {
		log.Infof("Write %d mutations to file", max(0, min(len(mutations)+len(specMutations), baseF.MaxNumberRuns-baseF.NumberWrittenMutations)))
	} else {
		log.Infof("Write %d mutations to file", max(0, len(mutations)+len(specMutations)))
	}

	first := baseF.NumberFuzzingRuns <= 1

	for _, mut := range specMutations {
		done, err := baseF.WriteMutConstraint(mut, first)
		first = false

		if done { // max number mutations has been reached
			break
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	for _, mut := range mutations {
		done, err := baseF.WriteMutConstraint(mut, first)
		first = false

		if done { // max number mutations has been reached
			break
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

// Calculate the energy for a schedule. This determines how many mutations
// are created
func getEnergy() int {

	// not interesting
	if baseA.GetTimeoutHappened(false) {
		return 0
	}

	w1 := settings.GoPieW1
	w2 := settings.GoPieW2

	score := 0

	score = int(w1*float64(counterCPOP1) + w2*math.Log(float64(counterCPOP2)))

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100
}
