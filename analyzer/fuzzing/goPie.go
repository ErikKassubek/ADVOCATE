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
			} else {
				utils.LogImportantf("B")
			}
		}
	}

	fuzzingPath := addFuzzingTraceFolder(pkgPath)
	if fuzzingPath == "" {
		return
	}

	original := analysis.GetTraces()
	for _, mut := range mutations {
		traces, err := analysis.CopyTrace(original)
		if err != nil {
			utils.LogError("Could not copy current trace")
		}

		tPosts := make([]int, len(mut.elems))
		routines := make(map[int]struct{})
		utils.LogImportantf("Len1: ", len(mut.elems))
		for i, elem := range mut.elems {
			tPosts[i] = elem.GetTPost()
			routines[elem.GetRoutine()] = struct{}{}
		}

		sort.Ints(tPosts)

		utils.LogImportantf("Len2: ", len(mut.elems))
		for i, elem := range mut.elems {
			routine, index := elem.GetTraceIndex()
			utils.LogImportantf("%d %d", routine, index)
			traces[routine][index].SetTSort(tPosts[i])
		}

		// TODO: is this sort necessary, only sort routines that where changed
		for routine := range routines {
			traces[routine] = analysis.SortTrace(traces[routine])
		}

		// remove all elements after the last elem in the chain
		lastTPost := tPosts[len(tPosts)-1]
		analysis.RemoveLater(lastTPost + 1)
		// add a replayEndElem
		analysis.AddTraceElementReplay(lastTPost+2, 0)

		fileName := filepath.Join(fuzzingPath, fmt.Sprintf("fuzzingTrace_%d", numberOfWrittenGoPieMuts))
		numberOfWrittenGoPieMuts++

		err = io.WriteTrace(fileName, analysis.GetNoRoutines())
		if err != nil {
			utils.LogError("Could not create pie mutation: ", err.Error())
		}

		mutationQueue = append(mutationQueue, mutation{mutType: mutPiType, mutPie: fileName})
	}
}

/*
 * Create the folder for the fuzzing traces of not exists
 * Args:
 * 	path (string): path to the folder
 * Returns:
 * 	string: path to the fuzzingTraces folder, or "" if an error occurred
 */
func addFuzzingTraceFolder(path string) string {
	p := filepath.Join(path, "fuzzingTraces")
	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		utils.LogError("Could not create folder")
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
