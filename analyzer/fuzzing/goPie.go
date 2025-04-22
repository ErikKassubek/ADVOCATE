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
)

// store all created mutations to avoid doubling
var allGoPieMutations = make(map[string]struct{})

// Create new mutations for GoPie
//
// Parameter:
//   - pkgPath string: path to where the new traces should be created
//   - numberFuzzingRun int: number of fuzzing run
func createGoPieMut(pkgPath string, numberFuzzingRuns int) {
	energy := getEnergy()

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

	fuzzingPath := filepath.Join(pkgPath, "fuzzingTraces")
	if numberFuzzingRuns == 0 {
		addFuzzingTraceFolder(fuzzingPath)
	}

	for _, mut := range mutations {
		if mut.len() == 0 {
			continue
		}

		traceCopy := analysis.CopyMainTrace()

		// remove all elements that are in the chain and all elements with a
		// tPost that is greater the smallest tPost in the chain

		minTPost := mut.earliestTPost()

		traceCopy.RemoveLater(minTPost - 1)

		// add the control element to the trace. This will switch the replay mode
		// from strict to partial

		traceCopy.AddTraceElementReplayControlCode(minTPost, analysis.ControlCodePartial)

		// set the tPosts of the chain elements to the correct order
		// and add them to the trace
		for i, elem := range mut.elems {
			elem.SetTSort(minTPost + 2*(i+1))
			traceCopy.AddElement(elem)
		}

		fileName := filepath.Join(fuzzingPath, fmt.Sprintf("fuzzingTrace_%d", numberOfWrittenGoPieMuts))
		numberOfWrittenGoPieMuts++

		err := io.WriteTrace(&traceCopy, fileName, true)
		if err != nil {
			utils.LogError("Could not create pie mutation: ", err.Error())
		}

		mutationQueue = append(mutationQueue, mutation{mutType: mutPiType, mutPie: fileName})
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
	numberSchedulChains := len(schedulingChains)

	// not interesting
	if analysis.GetTimeoutHappened() || numberSchedulChains == 0 {
		return 0
	}

	score := counterCPOP1 + int(math.Log(float64(counterCPOP2))) + 10*numberSchedulChains

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100
}
