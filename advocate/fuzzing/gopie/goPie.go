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
	"advocate/analysis/baseA"
	"advocate/analysis/hb/concurrent"
	"advocate/fuzzing/baseF"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/settings.go"
	"math"
)

const sameElem = true

// CreateMutations create new mutations for GoPie
//
// Parameter:
//   - mutNumber int: number of the mutation file
//   - error
func CreateMutations(mutNumber int) error {
	mutations := make(map[string]baseF.Chain)
	specMutations := make(map[string]baseF.Chain) // special mutations that should be run first

	// check for special chains, that could indicate a bug
	if flags.FuzzingMode != baseF.GoPie && baseF.UseHBInfoFuzzing {
		specialMuts := getSpecialMuts()

		for key, mut := range specialMuts {
			isValid := mut.IsValid()
			if _, ok := allGoPieMutations[key]; !ok {
				if !baseF.UseHBInfoFuzzing || isValid {
					specMutations[key] = mut
				}

				if !isValid {
					NumberInvalidMuts++
				}
				NumberTotalMuts++

				if ok {
					NumberDoubleMuts++
				}

				allGoPieMutations[key] = struct{}{}
			}
		}
	}

	// Original GoPie does not mutate all possible scheduling chains
	// If no SC is given, it creates a new one consisting of two random
	// operations that are in rel2 relation. Otherwise it always mutates the
	// original SC, not newly recorded once
	SchedulingChains = []baseF.Chain{}
	if flags.FuzzingMode == baseF.GoPie {
		if c, ok := baseF.ChainFiles[mutNumber]; ok {
			c.Old = true
			SchedulingChains = []baseF.Chain{c}
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
				isValid := mut.IsValid()
				if !baseF.UseHBInfoFuzzing || mut.IsValid() {
					mutations[key] = mut
				}

				if !isValid {
					NumberInvalidMuts++
				}
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
		done, err := baseF.WriteMutChain(mut, first)
		first = false

		if done { // max number mutations has been reached
			break
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	for _, mut := range mutations {
		done, err := baseF.WriteMutChain(mut, first)
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
	if baseA.GetTimeoutHappened() {
		return 0
	}

	w1 := settings.GoPieW1
	w2 := settings.GoPieW2

	score := 0

	if baseF.UseHBInfoFuzzing {
		for _, sc := range SchedulingChains {
			for _, elem := range sc.Elems {
				c := concurrent.GetNumberConcurrent(elem, sameElem, false, true)
				score += c
			}
		}
	} else {
		score = int(w1*float64(counterCPOP1) + w2*math.Log(float64(counterCPOP2)))
	}

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100
}

// Pass the trace and look for
//
//	channel close with concurrent send on the same channel
//
// # Based on those, create chains where the close if before the send
//
// Returns:
//   - map[string]Chain: map with the special chains
func getSpecialMuts() map[string]baseF.Chain {
	res := make(map[string]baseF.Chain)

	// send on closed
	for _, c := range baseA.CloseData {
		conc := concurrent.GetConcurrent(c, true, false, true, false)
		for _, s := range conc {
			switch t := s.(type) {
			case *trace.ElementSelect:
				for _, cc := range t.GetCases() {
					if cc.GetType(true) == trace.ChannelSend {
						chain := baseF.NewChain()
						chain.Add(c, s)
						res[chain.ToString()] = chain
					}
				}
			default:
				if s.GetType(true) == trace.ChannelSend {
					chain := baseF.NewChain()
					chain.Add(c, s)
					res[chain.ToString()] = chain
				}
			}
		}
	}

	// negative wg counter
	for id, dones := range baseA.WgDoneData {
		for _, done := range dones {
			for _, add := range baseA.WGAddData[id] {
				if add.GetTPost() > done.GetTPost() {
					continue
				}

				if concurrent.IsConcurrent(done, add) {
					chain := baseF.NewChain()
					chain.Add(done, add)
					res[chain.ToString()] = chain
				}
			}
		}
	}

	return res
}
