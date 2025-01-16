// Copyright (c) 2024 Erik Kassubek
//
// File: fuzzing.go
// Brief: Main file for fuzzing
//
// Author: Erik Kassubek
// Created: 2024-12-03
//
// License: BSD-3-Clause

package fuzzing

import (
	"analyzer/toolchain"
	"fmt"
	"math"
	"time"
)

const (
	maxNumberRuns = 20
	maxTime       = 20 * time.Minute
	maxRunPerMut  = 2
)

var (
	numberFuzzingRuns = 0
	mutationQueue     []map[string][]fuzzingSelect
	// count how often a specific mutation has been in the queue
	allMutations map[string]int
)

/*
* Create the fuzzing data
* Args:
* 	advocate (string): path to advocate
* 	testPath (string): path to the folder containing the test
* 	progName (string): name of the program
* 	testName (string): name of the test to run
 */
func Fuzzing(advocate, testPath, progName, testName string) error {
	startTime := time.Now()
	var order map[string][]fuzzingSelect

	// while there are available mutations, run them
	for numberFuzzingRuns == 0 || len(mutationQueue) != 0 {
		// if the program has not been run yet, run it directly, otherwise run it with order from queue
		if numberFuzzingRuns == 0 {
			err := toolchain.Run("test", advocate, testPath, "", progName, testName,
				-1, -1, 0, false, true, false, false, false, false)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			order, mutationQueue = mutationQueue[0], mutationQueue[1:]

			writeMutationsToFile(testPath, order)

			err := toolchain.Run("test", advocate, testPath, "", progName, testName,
				-1, -1, 0, true, true, false, false, false, false)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		// TODO: make sure that this works even if the trace has been rewritten by the analyzer
		parseTrace()

		// add new mutations
		if isInteresting() {
			createMutations(numberMutations(), getFlipProbability())
		}

		// TODO: keep file data internal

		numberFuzzingRuns++

		// cancel if max number of mutations have been reached
		if maxNumberRuns > numberFuzzingRuns {
			return fmt.Errorf("Maximum number of mutation runs (%d) have been reached", maxNumberRuns)
		}

		if time.Since(startTime) > maxTime {
			return fmt.Errorf(("Maximum runtime for fuzzing has been reached"))
		}
	}

	return nil

	// fuzzingFilePath := filepath.Join(pathFuzzing, fmt.Sprintf("fuzzingFile_%s.info", progName))
	// readFile(fuzzingFilePath)

	// io.CreateTraceFromFiles(pathTrace, true)
	// parseTrace()

	// // if the run was not interesting, there is nothing else to do
	// if !isInteresting() {
	// 	return lastID
	// }

	// numMut := numberMutations()
	// muts := createMutations(numMut)

	// updateFileData()
	// writeFileInfo(fuzzingFilePath)

	// return writeMutationsToFile(pathFuzzing, lastID, muts, progName)
}

/*
 * Get the probability that a select changes its preferred case
 * It is selected in such a way, that at least one of the selects if flipped
 * with a probability of at least 99%.
 * Additionally the flip probability is at least 10% for each select.
 */
func getFlipProbability() float64 {
	p := 0.99   // min prob that at least one case is flipped
	pMin := 0.1 // min prob that a select is flipt

	return max(pMin, 1-math.Pow(1-p, 1/float64(numberSelects)))
}
