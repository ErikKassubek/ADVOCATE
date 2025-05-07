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
	"advocate/analysis"
	"advocate/io"
	"advocate/utils"
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
const maxSCStart = 5

// Create new mutations for GoPie
//
// Parameter:
//   - pkgPath string: path to where the new traces should be created
//   - numberFuzzingRun int: number of fuzzing run
//   - mutNumber int: number of the mutation file
func createGoPieMut(pkgPath string, numberFuzzingRuns int, mutNumber int) {
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
			schedulingChains = append(schedulingChains, randomChain())
		}
	}

	energy := getEnergy()

	utils.LogInfof("Mutate %d scheduling chains", len(schedulingChains))

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

	for _, mut := range mutations {
		if maxNumberRuns != -1 && numberFuzzingRuns+len(mutationQueue) > maxNumberRuns {
			break
		}
		numberWrittenGoPieMuts++
		traceCopy := analysis.CopyMainTrace()

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

		err := io.WriteTrace(&traceCopy, fuzzingTracePath, true)
		if err != nil {
			utils.LogError("Could not create pie mutation: ", err.Error())
			continue
		}

		// write the active map to a "replay_active.log"
		if fuzzingMode == GoPie {
			writeMutActive(fuzzingTracePath, &traceCopy, &mut, 0)
		} else {
			writeMutActive(fuzzingTracePath, &traceCopy, &mut, mut.firstElement().GetTSort())
		}

		mut := mutation{mutType: mutPiType, mutPie: numberWrittenGoPieMuts}

		addMutToQueue(mut)
	}
}

// Create the folder for the fuzzing traces
//
// Parameter:
//   - path string: path to the folder
func addFuzzingTraceFolder(path string) {
	os.RemoveAll(path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		utils.LogError("Could not create fuzzing folder")
	}
}

// Calculate the energy for a schedule. This determines how many mutations
// are created
func getEnergy() int {

	// not interesting
	if analysis.GetTimeoutHappened() {
		return 0
	}

	w1 := utils.GoPieW1
	w2 := utils.GoPieW2

	score := int(w1*float64(counterCPOP1) + w2*math.Log(float64(counterCPOP2)))

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100
}
