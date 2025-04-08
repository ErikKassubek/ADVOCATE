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
	"analyzer/analysis"
	"analyzer/io"
	"analyzer/utils"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// store all created mutations to avoid doubling
var allGoPieMutations = make(map[string]struct{})

/*
 * Create new mutations for GoPie
 * Args:
 * 	pkgPath (string): path to where the new traces should be created
 * 	numberFuzzingRun (int): number of fuzzing run
 */
func createGoPieMut(pkgPath string, numberFuzzingRuns int) {
	energy := getEnergy(numberFuzzingRuns != 0, len(schedulingChains))

	mutations := make(map[string]chain)

	utils.LogInfof("Trace contains %d scheduling chains", len(schedulingChains))

	for _, sc := range schedulingChains {
		muts := mutate(sc, energy)
		for key, mut := range muts {
			if _, ok := allGoPieMutations[key]; !ok {
				// only add if not invalidated by hb
				if !useHBInfoFuzzing || mut.isValid() {
					mutations[key] = mut
				}
				allGoPieMutations[key] = struct{}{}
			}
		}
	}

	fuzzingPath := addFuzzingTraceFolder(pkgPath)
	if fuzzingPath == "" {
		return
	}

	for _, mut := range mutations {
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

		fileName := filepath.Join(fuzzingPath, fmt.Sprintf("fuzzingTrace_%d", numberOfWrittenGoPieMuts))
		numberOfWrittenGoPieMuts++

		err := io.WriteTrace(&traceCopy, fileName)
		if err != nil {
			utils.LogError("Could not create pie mutation: ", err.Error())
		}

		mutationQueue = append(mutationQueue, mutation{mutType: mutPiType, mutPie: fileName})
	}
}

/*
 * Create the folder for the fuzzing traces
 * Args:
 * 	path (string): path to the folder
 * Returns:
 * 	string: path to the fuzzingTraces folder, or "" if an error occurred
 */
func addFuzzingTraceFolder(path string) string {
	p := filepath.Join(path, "fuzzingTraces")
	os.RemoveAll(p)
	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		utils.LogError("Could not create fuzzing folder")
		return ""
	}
	return p
}

/*
 * Calculate the energy for a schedule. This determines how many mutations
 * are created
 * Args:
 * 	recordedBasedOnMutation (bool): False if the recording was the first fuzzing run, otherwise true
 * 	numberSchedulChain (int): Number of scheduling chains in the program
 */
func getEnergy(recordingBasedOnMutation bool, numberSchedulChains int) int {
	score := counterCPOP1 + int(math.Log(float64(counterCPOP2)))

	if recordingBasedOnMutation {
		if analysis.GetTimeoutHappened() {
			score = 0
		} else {
			score += 10 * numberSchedulChains
		}
	}

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100

}
