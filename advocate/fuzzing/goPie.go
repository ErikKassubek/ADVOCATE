// Copyright (c) 2025 Erik Kassubek
//
// File: goPie.go
// Brief: Main file for goPie fuzzing
//
// Author: Erik Kassubek
// Created: 2025-03-22
//
// License: BSD-3-Clause

package fuzzing

import (
	"advocate/analysis/analysis"
	"advocate/utils/helper"
	"advocate/utils/io"
	"advocate/utils/log"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// store all created mutations to avoid doubling
var allGoPieMutations = make(map[string]struct{})

// for each mutation file, store the file number and the chain
var chainFiles = make(map[int]chain)

// number of different starting points for chains in GoPie (in the original: cfg.MaxWorker)
var maxSCStart = helper.GoPieSCStart

// Create new mutations for GoPie
//
// Parameter:
//   - pkgPath string: path to where the new traces should be created
//   - numberFuzzingRun int: number of fuzzing run
//   - mutNumber int: number of the mutation file
//   - error
func createGoPieMut(pkgPath string, numberFuzzingRuns int, mutNumber int) error {
	mutations := make(map[string]chain)

	// Original GoPie does not mutate all possible scheduling chains
	// If no SC is given, it creates a new one consisting of two random
	// operations that are in rel2 relation. Otherwise it always mutates the
	// original SC, not newly recorded once
	schedulingChains = []chain{}
	if fuzzingMode == GoPie {
		if c, ok := chainFiles[mutNumber]; ok {
			schedulingChains = []chain{c}
		}
	}

	if len(schedulingChains) == 0 {
		for range maxSCStart {
			sc := randomChain()
			if sc.len() > 0 {
				schedulingChains = append(schedulingChains, sc)
			}
		}
	}

	energy := getEnergy()

	log.Infof("Mutate %d scheduling chains", len(schedulingChains))

	for _, sc := range schedulingChains {
		muts := mutate(sc, energy)
		for key, mut := range muts {
			if fuzzingMode != GoPie && mut.len() <= 1 {
				continue
			}
			if _, ok := allGoPieMutations[key]; fuzzingMode == GoPie || !ok {
				// only add if not invalidated by hb
				if !useHBInfoFuzzing || mut.isValid() {
					mutations[key] = mut
				}
				allGoPieMutations[key] = struct{}{}
			}
		}
	}

	fuzzingPath := filepath.Join(pkgPath, "fuzzingTraces")
	if numberFuzzingRuns == 0 {
		addFuzzingTraceFolder(fuzzingPath)
	}

	log.Infof("Write %d mutations to file", min(len(mutations), maxNumberRuns-numberWrittenGoPieMuts))
	for _, mut := range mutations {
		if maxNumberRuns != -1 && numberWrittenGoPieMuts > maxNumberRuns {
			break
		}
		numberWrittenGoPieMuts++

		traceCopy, err := analysis.CopyMainTrace()
		if err != nil {
			return err
		}

		tPosts := make([]int, len(mut.elems))
		routines := make(map[int]struct{})
		for i, elem := range mut.elems {
			tPosts[i] = elem.GetTPost()
			routines[elem.GetRoutine()] = struct{}{}
		}

		sort.Ints(tPosts)

		changedRoutinesMap := make(map[int]struct{})

		for i, elem := range mut.elems {
			routine, index := elem.GetTraceIndex()
			traceCopy.SetTSortAtIndex(tPosts[i], routine, index)
			changedRoutinesMap[routine] = struct{}{}
		}

		changedRoutines := make([]int, 0, len(changedRoutinesMap))
		for k := range changedRoutinesMap {
			changedRoutines = append(changedRoutines, k)
		}

		traceCopy.SortRoutines(changedRoutines)

		// remove all elements after the last elem in the chain
		lastTPost := tPosts[len(tPosts)-1]
		traceCopy.RemoveLater(lastTPost + 1)

		// add a replayEndElem
		traceCopy.AddTraceElementReplay(lastTPost+2, 0)

		fuzzingTracePath := filepath.Join(fuzzingPath, fmt.Sprintf("fuzzingTrace_%d", numberWrittenGoPieMuts))
		chainFiles[numberWrittenGoPieMuts] = mut

		err = io.WriteTrace(&traceCopy, fuzzingTracePath, true)
		if err != nil {
			log.Error("Could not create pie mutation: ", err.Error())
			continue
		}

		// write the active map to a "replay_active.log"
		if fuzzingMode == GoPie {
			writeMutActive(fuzzingTracePath, &traceCopy, &mut, 0)
		} else {
			writeMutActive(fuzzingTracePath, &traceCopy, &mut, mut.firstElement().GetTSort())
		}

		traceCopy.Clear()

		mut := mutation{mutType: mutPiType, mutPie: numberWrittenGoPieMuts}

		addMutToQueue(mut)
	}

	return nil
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
	if analysis.GetTimeoutHappened() {
		return 0
	}

	w1 := helper.GoPieW1
	w2 := helper.GoPieW2

	score := int(w1*float64(counterCPOP1) + w2*math.Log(float64(counterCPOP2)))

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100
}
