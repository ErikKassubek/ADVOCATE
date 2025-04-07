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
 * Returns:
 * 	error
 */
func createGoPieMut(pkgPath string, numberFuzzingRuns int) error {
	energy := getEnergy(numberFuzzingRuns != 0, len(schedulingChains))

	mutations := make(map[string]chain)

	utils.LogInfof("Trace contains %d scheduling chains", len(schedulingChains))

	for _, sc := range schedulingChains {
		muts := mutate(sc, energy)
		for key, mut := range muts {
			if _, ok := allGoPieMutations[key]; !ok && mut.isValid() { // is mut new and HB valid
				mutations[key] = mut
				allGoPieMutations[key] = struct{}{}
			}
		}
	}

	for _, mut := range mutations {
		traces := analysis.GetTraces()
		oldTrace, err := analysis.CopyCurrentTrace()
		if err != nil {
			return fmt.Errorf("Could not copy trace: ", err.Error())
		}

		tPosts := make([]int, 0)
		routines := make(map[int]struct{})
		for _, elem := range mut.elems {
			tPosts = append(tPosts, elem.GetTPost())
			routines[elem.GetRoutine()] = struct{}{}
		}

		sort.Ints(tPosts)

		for i, elem := range mut.elems {
			routine, index := elem.GetTraceIndex()
			traces[routine][index].SetTSort(tPosts[i])
		}

		// TODO: is this sort necessary
		for routine := range routines {
			traces[routine] = analysis.SortTrace(traces[routine])
		}

		// remove all elements after the last elem in the chain
		lastTPost := tPosts[len(tPosts)-1]
		analysis.RemoveLater(lastTPost + 1)
		// add a replayEndElem
		analysis.AddTraceElementReplay(lastTPost+2, 0)

		fileName := filepath.Join(pkgPath, fmt.Sprintf("fuzzingTrace_%d", numberOfWrittenGoPieMuts))
		numberOfWrittenGoPieMuts++

		err = io.WriteTrace(fileName, analysis.GetNoRoutines())
		if err != nil {
			analysis.SetTrace(oldTrace)
			return fmt.Errorf("Could not create pie mutation")
		}

		mutationQueue = append(mutationQueue, mutation{mutType: mutPiType, mutPie: fileName})

		analysis.SetTrace(oldTrace)

	}

	return nil
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
