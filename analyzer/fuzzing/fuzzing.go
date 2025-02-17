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
	"analyzer/analysis"
	"analyzer/stats"
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
* 	modeMain (bool): if true, run fuzzing on main function, otherwise on test
* 	advocate (string): path to advocate
* 	progPath (string): path to the folder containing the prog/test
* 	progName (string): name of the program
* 	name (string): If modeMain, name of the executable, else name of the test
* 	ignoreAtomic (bool): if true, ignore atomics for replay
* 	hBInfoFuzzing (bool): whether to us HB info in fuzzing
* 	fullAnalysis (bool): if true, run full analysis and replay, otherwise only detect actual bugs
* 	meaTime (bool): measure runtime
* 	notExec (bool): find never executed operations
* 	stats (bool): create statistics
* 	keepTraces (bool): keep the traces after analysis
 */
func Fuzzing(modeMain bool, advocate, progPath, progName, name string, ignoreAtomic,
	hBInfoFuzzing, fullAnalysis, meaTime, notExec, createStats, keepTraces bool) error {

	log.Println("Start fuzzing")

	// run either fuzzing on main or fuzzing on one test
	if modeMain || name != "" {
		if modeMain {
			log.Println("Run fuzzing on main function")
		} else {
			log.Println("Run fuzzing on test ", name)
		}

		err := runFuzzing(modeMain, advocate, progPath, progName, name, ignoreAtomic,
			hBInfoFuzzing, fullAnalysis, meaTime, notExec, createStats, keepTraces, true)

		if createStats {
			err := stats.CreateStatsFuzzing(getPath(progPath), progName)
			if err != nil {
				log.Println("Failed to create fuzzing stats: ", err.Error())
			}
			err = stats.CreateStatsTotal(getPath(progPath), progName)
			if err != nil {
				log.Println("Failed to create total stats: ", err.Error())
			}
		}

		return err
	}

	log.Println("Run fuzzing on all tests")

	// run fuzzing on all tests
	testFiles, err := toolchain.FindTestFiles(progPath)
	if err != nil {
		return fmt.Errorf("Failed to find test files: %v", err)
	}

	totalFiles := len(testFiles)

	log.Printf("Found %d test files", totalFiles)

	// Process each test file
	for i, testFile := range testFiles {
		log.Printf("Progress %s: %d/%d\n", progName, i+1, totalFiles)
		log.Printf("Processing file: %s\n", testFile)

		testFunctions, err := toolchain.FindTestFunctions(testFile)
		if err != nil || len(testFunctions) == 0 {
			log.Println("Could not find test functions in ", testFile)
			continue
		}

		for j, testFunc := range testFunctions {
			resetFuzzing()

			log.Printf("Run fuzzing for %s->%s", testFile, testFunc)

			firstRun := (i == 0 && j == 0)

			err := runFuzzing(false, advocate, progPath, progName, testFunc, ignoreAtomic,
				hBInfoFuzzing, fullAnalysis, meaTime, notExec, createStats, keepTraces, firstRun)
			if err != nil {
				log.Println("Error in fuzzing: ")
			}
		}
	}

	if createStats {
		err := stats.CreateStatsFuzzing(getPath(progPath), progName)
		if err != nil {
			log.Println("Failed to create fuzzing stats: ", err.Error())
		}
		err = stats.CreateStatsTotal(getPath(progPath), progName)
		if err != nil {
			log.Println("Failed to create total stats: ", err.Error())
		}
	}

	return nil
}

/*
* Run Fuzzing on one program/test
* Args:
* 	modeMain (bool): if true, run fuzzing on main function, otherwise on test
* 	advocate (string): path to advocate
* 	progPath (string): path to the folder containing the prog/test
* 	progName (string): name of the program
* 	name (string): If modeMain, name of the executable, else name of the test
* 	ignoreAtomic (bool): if true, ignore atomics for replay
* 	hBInfoFuzzing (bool): whether to us HB info in fuzzing
* 	fullAnalysis (bool): if true, run full analysis and replay, otherwise only detect actual bugs
* 	meaTime (bool): measure runtime
* 	notExec (bool): find never executed operations
* 	createStats (bool): create statistics
* 	keepTraces (bool): keep the traces after analysis
* 	firstRun (bool): this is the first run, only set to false for fuzzing (except for the first fuzzing)
 */
func runFuzzing(modeMain bool, advocate, progPath, progName, name string, ignoreAtomic,
	hBInfoFuzzing, fullAnalysis, meaTime, notExec, createStats, keepTraces, firstRun bool) error {
	useHBInfoFuzzing = hBInfoFuzzing
	runFullAnalysis = fullAnalysis

	progDir := getPath(progPath)

	startTime := time.Now()
	var order map[string][]fuzzingSelect

	// while there are available mutations, run them
	for numberFuzzingRuns == 0 || len(mutationQueue) != 0 {
		log.Println("Fuzzing Run: ", numberFuzzingRuns+1)

		order = popMutation()

		if numberFuzzingRuns != 0 {
			err := writeMutationsToFile(progDir, order)
			if err != nil {
				panic(err.Error())
			}
		}

		firstRun = firstRun && (numberFuzzingRuns == 0)

		// Run the test/mutation

		mode := "test"
		if modeMain {
			mode = "main"
		}
		err := toolchain.Run(mode, advocate, progPath, name, progName, name,
			-1, -1, 0, numberFuzzingRuns, ignoreAtomic, meaTime, notExec, createStats, keepTraces, firstRun)
		if err != nil {
			fmt.Println(err.Error())
		}

		// TODO: the function to get all the infos required for isInteresting and
		// numberMutations is called in modeAnalyzer in the main.go file, which
		// itself is run by the toolchain.Run function via function injection.
		// At some point this should be refactored to make it less complicated

		// add new mutations based on GFuzz select
		if isInterestingSelect() {
			numberMut := numberMutations()
			flipProb := getFlipProbability()
			numMutAdd := createMutations(numberMut, flipProb)
			log.Printf("Add %d mutations to queue\n", numMutAdd)
		} else {
			log.Println("Add 0 mutations to queue")
		}

		mergeTraceInfoIntoFileInfo()

		// clean up
		clearData()
		analysis.ClearData()
		analysis.ClearTrace()

		numberFuzzingRuns++

		// cancel if max number of mutations have been reached
		if numberFuzzingRuns > maxNumberRuns {
			return fmt.Errorf("Maximum number of mutation runs (%d) have been reached", maxNumberRuns)
		}

		if time.Since(startTime) > maxTime {
			return fmt.Errorf(("Maximum runtime for fuzzing has been reached"))
		}
	}

	log.Printf("Finish fuzzing after %d runs\n", numberFuzzingRuns)

	return nil
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

func resetFuzzing() {
	numberFuzzingRuns = 0
	mutationQueue = make([]map[string][]fuzzingSelect, 0)
	// count how often a specific mutation has been in the queue
	allMutations = make(map[string]int)
}
