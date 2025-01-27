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
	"log"
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
	mutationQueue     = make([]map[string][]fuzzingSelect, 0)
	// count how often a specific mutation has been in the queue
	allMutations = make(map[string]int)
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
	log.Println("Start fuzzing")
	startTime := time.Now()
	var order map[string][]fuzzingSelect

	// while there are available mutations, run them
	for numberFuzzingRuns == 0 || len(mutationQueue) != 0 {
		log.Println("Run: ", numberFuzzingRuns)
		fmt.Println("Run: ", numberFuzzingRuns)

		order = popMutation()

		if numberFuzzingRuns != 0 {
			err := writeMutationsToFile(testPath, order)
			if err != nil {
				panic(err.Error())
			}
		}

		// Run the test/mutation
		err := toolchain.Run("test", advocate, testPath, "", progName, testName,
			-1, -1, 0, numberFuzzingRuns, true, false, false, false, false)
		if err != nil {
			fmt.Println(err.Error())
		}

		// TODO: the function to get all the infos required for isInteresting and
		// numberMutations is called in modeAnalyzer in the main.go file, which
		// itself is run by the toolchain.Run function via function injection.
		// At some point this should be refactored to make it less complicated

		// add new mutations based on GFuzz select
		if isInterestingSelect() {
			fmt.Println("Create mutations")
			numberMut := numberMutations()
			flipProb := getFlipProbability()
			createMutations(numberMut, flipProb)
		}

		mergeTraceInfoIntoFileInfo()

		numberFuzzingRuns++

		// cancel if max number of mutations have been reached
		if numberFuzzingRuns > maxNumberRuns {
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
