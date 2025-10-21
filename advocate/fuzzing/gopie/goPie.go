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
	"advocate/io"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/settings.go"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

const sameElem = true

// CreateMutations create new mutations for GoPie
//
// Parameter:
//   - pkgPath string: path to where the new traces should be created
//   - numberFuzzingRun int: number of fuzzing run
//   - mutNumber int: number of the mutation file
//   - error
func CreateMutations(pkgPath string, numberFuzzingRuns int, mutNumber int) error {
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
		if c, ok := chainFiles[mutNumber]; ok {
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
		muts := mutate(sc, energy)

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

	fuzzingPath := filepath.Join(pkgPath, "fuzzingTraces")
	if numberFuzzingRuns <= 1 {
		addFuzzingTraceFolder(fuzzingPath)
	}

	if len(specMutations) > 0 {
		log.Infof("Write %d special mutation to file", len(specMutations))
	}

	if baseF.MaxNumberRuns > 0 {
		log.Infof("Write %d mutations to file", max(0, min(len(mutations)+len(specMutations), baseF.MaxNumberRuns-numberWrittenGoPieMuts)))
	} else {
		log.Infof("Write %d mutations to file", max(0, len(mutations)+len(specMutations)))
	}

	for _, mut := range specMutations {
		done, err := writeMut(mut, fuzzingPath)

		if done { // max number mutations has been reached
			break
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	for _, mut := range mutations {
		done, err := writeMut(mut, fuzzingPath)

		if done { // max number mutations has been reached
			break
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

// Write the mutation to file and add it to the queue
//
// Parameter:
//   - mut Chain: the mutation to write
//   - fuzzingPath string: path to where the muts are written
//
// Returns:
//   - bool: true if max number muts in reached
//   - error
func writeMut(mut baseF.Chain, fuzzingPath string) (bool, error) {
	if baseF.MaxNumberRuns != -1 && numberWrittenGoPieMuts > baseF.MaxNumberRuns {
		return true, nil
	}
	numberWrittenGoPieMuts++

	traceCopy, err := baseA.CopyMainTrace()
	if err != nil {
		return false, err
	}

	t1 := -1
	for _, elem := range mut.Elems {
		tPost := elem.GetTPost()
		if t1 == -1 || tPost < t1 {
			t1 = tPost
		}
	}

	// remove all elements after the first elem in the chain
	traceCopy.ShortenTrace(t1, false)

	// add in all the elements in the chain
	mapping := make(map[string]trace.Element)
	for i, elem := range mut.Elems {
		c := elem.Copy(mapping)
		c.SetTSort(t1 + i*2)
		traceCopy.AddElement(c)
	}

	fuzzingTracePath := filepath.Join(fuzzingPath, fmt.Sprintf("fuzzingTrace_%d", numberWrittenGoPieMuts))
	chainFiles[numberWrittenGoPieMuts] = mut

	err = io.WriteTrace(&traceCopy, fuzzingTracePath, true)
	if err != nil {
		return false, fmt.Errorf("Could not create pie mutation: %s", err.Error())
	}

	// write the active map to a "replay_active.log"
	if flags.FuzzingMode == baseF.GoPie || settings.WithoutReplay {
		baseF.WriteMutActive(fuzzingTracePath, &traceCopy, &mut, 0)
	} else {
		baseF.WriteMutActive(fuzzingTracePath, &traceCopy, &mut, mut.FirstElement().GetTPost())
	}

	traceCopy.Clear()

	muta := baseF.Mutation{MutType: baseF.MutPiType, MutPie: numberWrittenGoPieMuts}

	baseF.AddMutToQueue(muta, false, false)

	return false, nil
}

// Create the folder for the fuzzing traces
//
// Parameter:
//   - path string: path to the folder
func addFuzzingTraceFolder(path string) {
	os.RemoveAll(path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Error("Could not create fuzzing folder")
	}
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
