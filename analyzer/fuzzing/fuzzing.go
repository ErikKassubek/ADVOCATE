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
	"analyzer/stats"
	"analyzer/timer"
	"analyzer/toolchain"
	"analyzer/utils"
	"fmt"
	"math"
	"time"
)

type mutation struct {
	mutType int
	mutSel  map[string][]fuzzingSelect
	mutFlow map[string]int
	mutPie  string
}

const (
	GFuzz     = "GFuzz"     // only GFuzz
	GFuzzHB   = "GFuzzHB"   // GFuzz with use of hb info
	GFuzzFlow = "GFuzzFlow" // GFuzz with use of hb info and flow mutation
	Flow      = "Flow"      // only flow mutation
	GoPie     = "GoPie"     // only goPie

	mutSelType  = 0
	mutPiType   = 1
	mutFlowType = 2
)
const (
	maxNumberRuns = 20
	maxTime       = 60 * time.Minute
	maxRunPerMut  = 2

	factorCaseWithPartner = 2
	maxFlowMut            = 10
)

var (
	numberFuzzingRuns = 0
	mutationQueue     = make([]mutation, 0)
	// count how often a specific mutation has been in the queue
	allMutations     = make(map[string]int)
	fuzzingMode      = ""
	fuzzingModeGFuzz = false
	fuzzingModeGoPie = false
	fuzzingModeFlow  = false
)

/*
* Create the fuzzing data
* Args:
* 	modeMain (bool): if true, run fuzzing on main function, otherwise on test
* 	fm (bool): the mode used for fuzzing
* 	advocate (string): path to advocate
* 	progPath (string): path to the folder containing the prog/test
* 	progName (string): name of the program
* 	name (string): If modeMain, name of the executable, else name of the test
* 	ignoreAtomic (bool): if true, ignore atomics for replay
* 	meaTime (bool): measure runtime
* 	notExec (bool): find never executed operations
* 	stats (bool): create statistics
* 	keepTraces (bool): keep the traces after analysis
* 	cont (bool): continue partial fuzzing
 */
func Fuzzing(modeMain bool, fm, advocate, progPath, progName, name string, ignoreAtomic,
	meaTime, notExec, createStats, keepTraces, cont bool) error {

	modes := []string{GoPie, GFuzz, GFuzzFlow, GFuzzHB, Flow}
	if !utils.ContainsString(modes, fm) {
		return fmt.Errorf("Invalid fuzzing mode '%s'", fm)
	}

	fuzzingMode = fm
	fuzzingModeGoPie = (fuzzingMode == GoPie)
	fuzzingModeGFuzz = (fuzzingMode == GFuzz || fuzzingMode == GFuzzFlow || fuzzingMode == GFuzzHB)
	fuzzingModeFlow = (fuzzingMode == Flow || fuzzingMode == GFuzzFlow)
	useHBInfoFuzzing = (fuzzingMode == GFuzzHB || fuzzingMode == GFuzzFlow || fuzzingMode == Flow)

	if cont {
		utils.LogInfo("Continue fuzzing")
	} else {
		utils.LogInfo("Start fuzzing")
	}

	// run either fuzzing on main or fuzzing on one test
	if modeMain || name != "" {
		if modeMain {
			utils.LogInfo("Run fuzzing on main function")
		} else {
			utils.LogInfo("Run fuzzing on test ", name)
		}

		err := runFuzzing(modeMain, advocate, progPath, progName, "", name, ignoreAtomic,
			meaTime, notExec, createStats, keepTraces, true, cont, 0, 0)

		if createStats {
			err := stats.CreateStatsFuzzing(getPath(progPath), progName)
			if err != nil {
				utils.LogError("Failed to create fuzzing stats: ", err.Error())
			}
			err = stats.CreateStatsTotal(getPath(progPath), progName)
			if err != nil {
				utils.LogError("Failed to create total stats: ", err.Error())
			}
		}

		return err
	}

	utils.LogInfo("Run fuzzing on all tests")

	// run fuzzing on all tests
	testFiles, maxFileNumber, totalFiles, err := toolchain.FindTestFiles(progPath, cont)
	if err != nil {
		return fmt.Errorf("Failed to find test files: %v", err)
	}

	utils.LogInfof("Found %d test files", totalFiles)

	// Process each test file
	fileCounter := 0
	if cont {
		fileCounter = maxFileNumber
	}

	for i, testFile := range testFiles {
		fileCounter++
		utils.LogInfof("Progress %s: %d/%d\n", progName, fileCounter, totalFiles)
		utils.LogInfof("Processing file: %s\n", testFile)

		testFunctions, err := toolchain.FindTestFunctions(testFile)
		if err != nil || len(testFunctions) == 0 {
			utils.LogInfo("Could not find test functions in ", testFile)
			continue
		}

		for j, testFunc := range testFunctions {
			resetFuzzing()
			timer.ResetTest()
			timer.Start(timer.TotalTest)

			utils.LogInfof("Run fuzzing for %s->%s", testFile, testFunc)

			firstRun := (i == 0 && j == 0)

			err := runFuzzing(false, advocate, progPath, progName, testFile, testFunc, ignoreAtomic,
				meaTime, notExec, createStats, keepTraces, firstRun, cont, fileCounter, j+1)
			if err != nil {
				utils.LogError("Error in fuzzing: ", err.Error())
			}

			timer.Stop(timer.TotalTest)

			timer.UpdateTimeFileOverview(progName, testFunc)
		}

	}

	if createStats {
		err := stats.CreateStatsFuzzing(getPath(progPath), progName)
		if err != nil {
			utils.LogError("Failed to create fuzzing stats: ", err.Error())
		}
		err = stats.CreateStatsTotal(getPath(progPath), progName)
		if err != nil {
			utils.LogError("Failed to create total stats: ", err.Error())
		}
	}

	return nil
}

/*
* Run Fuzzing on one program/test
* Args:
* 	modeMain (bool): if true, run fuzzing on main function, otherwise on test
* 	advocate (string): path to advocate
* 	progName (string): name of the program
* 	testPath (string): path to the test file
* 	name (string): If modeMain, name of the executable, else name of the test
* 	ignoreAtomic (bool): if true, ignore atomics for replay
* 	hBInfoFuzzing (bool): whether to us HB info in fuzzing
* 	meaTime (bool): measure runtime
* 	notExec (bool): find never executed operations
* 	createStats (bool): create statistics
* 	keepTraces (bool): keep the traces after analysis
* 	firstRun (bool): this is the first run, only set to false for fuzzing (except for the first fuzzing)
* 	cont (bool): continue with an already started run
 */
func runFuzzing(modeMain bool, advocate, progPath, progName, testPath, name string, ignoreAtomic,
	meaTime, notExec, createStats, keepTraces, firstRun, cont bool, fileNumber, testNumber int) error {

	progDir := getPath(progPath)

	clearDataFull()

	startTime := time.Now()

	// while there are available mutations, run them
	for numberFuzzingRuns == 0 || len(mutationQueue) != 0 {
		utils.LogInfo("Fuzzing Run: ", numberFuzzingRuns+1)

		if numberFuzzingRuns != 0 {
			order := popMutation()
			err := writeMutationToFile(progDir, order)
			if err != nil {
				return err
			}
		}

		firstRun = firstRun && (numberFuzzingRuns == 0)

		// Run the test/mutation

		mode := "test"
		if modeMain {
			mode = "main"
		}
		err := toolchain.Run(mode, advocate, progPath, testPath, name, progName, name,
			0, numberFuzzingRuns, ignoreAtomic, meaTime, notExec, createStats, keepTraces, firstRun, cont, fileNumber, testNumber)
		if err != nil {
			utils.LogError("Fuzzing run failed: ", err.Error())
		} else {
			if fuzzingModeGFuzz {
				createGFuzzMut()
			}

			// add new mutations based on flow path expansion
			if fuzzingModeFlow {
				createMutationsFlow()
			}

			// add mutations based on GoPie
			if fuzzingModeGoPie {
				createGoPieMut(progDir)
			}

			utils.LogInfof("Current fuzzing queue size: %d", len(mutationQueue))

			mergeTraceInfoIntoFileInfo()
		}

		// clean up
		clearData()
		timer.ResetFuzzing()

		numberFuzzingRuns++

		// cancel if max number of mutations have been reached
		if numberFuzzingRuns > maxNumberRuns {
			utils.LogInfof("Finish fuzzing because maximum number of mutation runs (%d) have been reached", maxNumberRuns)
			return nil
		}

		if time.Since(startTime) > maxTime {
			utils.LogInfof("Finish fuzzing because maximum runtime for fuzzing (%d min)has been reached", int(maxTime.Minutes()))
			return nil
		}
	}

	toolchain.RemoveFuzzingTrace(progDir)

	utils.LogInfof("Finish fuzzing after %d runs\n", numberFuzzingRuns)

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
	mutationQueue = make([]mutation, 0)
	// count how often a specific mutation has been in the queue
	allMutations = make(map[string]int)
}
