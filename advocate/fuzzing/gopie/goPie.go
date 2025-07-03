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
	anadata "advocate/analysis/data"
	"advocate/fuzzing/data"
	"advocate/io"
	"advocate/utils/helper"
	"advocate/utils/log"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// CreateGoPieMut create new mutations for GoPie
//
// Parameter:
//   - pkgPath string: path to where the new traces should be created
//   - numberFuzzingRun int: number of fuzzing run
//   - mutNumber int: number of the mutation file
//   - error
func CreateGoPieMut(pkgPath string, numberFuzzingRuns int, mutNumber int) error {
	mutations := make(map[string]Chain)

	// Original GoPie does not mutate all possible scheduling chains
	// If no SC is given, it creates a new one consisting of two random
	// operations that are in rel2 relation. Otherwise it always mutates the
	// original SC, not newly recorded once
	SchedulingChains = []Chain{}
	if data.FuzzingMode == data.GoPie {
		if c, ok := chainFiles[mutNumber]; ok {
			SchedulingChains = []Chain{c}
		}
	}

	if len(SchedulingChains) == 0 {
		for range maxSCStart {
			sc := startChain()
			if sc.Len() > 0 {
				SchedulingChains = append(SchedulingChains, sc)
			}
		}
	}

	energy := getEnergy()

	log.Infof("Mutate %d scheduling chains", len(SchedulingChains))

	for _, sc := range SchedulingChains {
		muts := mutate(sc, energy)
		for key, mut := range muts {
			if data.FuzzingMode != data.GoPie && mut.Len() <= 1 {
				continue
			}
			if _, ok := allGoPieMutations[key]; data.FuzzingMode == data.GoPie || !ok {
				// only add if not invalidated by hb
				if !data.UseHBInfoFuzzing || mut.isValid() {
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

	log.Infof("Write %d mutations to file", min(len(mutations), data.MaxNumberRuns-numberWrittenGoPieMuts))
	for _, mut := range mutations {
		if data.MaxNumberRuns != -1 && numberWrittenGoPieMuts > data.MaxNumberRuns {
			break
		}
		numberWrittenGoPieMuts++

		traceCopy, err := anadata.CopyMainTrace()
		if err != nil {
			return err
		}

		tPosts := make([]int, len(mut.Elems))
		routines := make(map[int]struct{})
		for i, elem := range mut.Elems {
			tPosts[i] = elem.GetTPost()
			routines[elem.GetRoutine()] = struct{}{}
		}

		sort.Ints(tPosts)

		changedRoutinesMap := make(map[int]struct{})

		for i, elem := range mut.Elems {
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
		if data.FuzzingMode == data.GoPie {
			writeMutActive(fuzzingTracePath, &traceCopy, &mut, 0)
		} else {
			writeMutActive(fuzzingTracePath, &traceCopy, &mut, mut.firstElement().GetTSort())
		}

		traceCopy.Clear()

		mut := data.Mutation{MutType: data.MutPiType, MutPie: numberWrittenGoPieMuts}

		data.AddMutToQueue(mut)
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
	if anadata.GetTimeoutHappened() {
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
