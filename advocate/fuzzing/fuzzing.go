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
	anaData "advocate/analysis/data"
	"advocate/fuzzing/data"
	"advocate/fuzzing/flow"
	"advocate/fuzzing/gfuzz"
	"advocate/fuzzing/gopie"
	"advocate/results/results"
	"advocate/results/stats"
	"advocate/toolchain"
	"advocate/utils/control"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
	"advocate/utils/types"
	"fmt"
	"path/filepath"
	"time"
)

// Fuzzing creates the fuzzing data and runs the fuzzing executions
//
// Parameter:
//   - modeMain bool: if true, run fuzzing on main function, otherwise on test
//   - fm bool: the mode used for fuzzing
//   - advocate string: path to advocate
//   - progPath string: path to the folder containing the prog/test
//   - progName string: name of the program
//   - name string: If modeMain, name of the executable, else name of the test
//   - ignoreAtomic bool: if true, ignore atomics for replay
//   - meaTime bool: measure runtime
//   - notExec bool: find never executed operations
//   - stats bool: create statistics
//   - keepTraces bool: keep the traces after analysis
//   - skipExisting bool: skip existing runs
//   - cont bool: continue partial fuzzing
//   - mTime int: maximum time in seconds spend for one test/prog\
//   - mRun int: maximum number of times a test/prog is run
//   - cancelTestIfFound int: do not run further fuzzing runs on tests if one
//     bug has been found, mainly used for benchmarks
func Fuzzing(modeMain bool, fm, advocate, progPath, progName, name string, ignoreAtomic,
	meaTime, notExec, createStats, keepTraces, cont bool, mTime, mRun int,
	cancelTestIfFound bool) error {

	if fm == "" {
		return fmt.Errorf("No fuzzing mode selected. Select with -fuzzingMode [mode]. Possible values are GoPie, GoPie+, GoPieHB, GFuzz, GFuzzFlow, GFuzzHB, Flow")
	}

	modes := []string{data.GoPie, data.GoPiePlus, data.GoPieHB, data.GFuzz, data.GFuzzHBFlow, data.GFuzzHB, data.Flow}
	if !types.Contains(modes, fm) {
		return fmt.Errorf("Invalid fuzzing mode '%s'. Possible values are GoPie, GoPie+, GoPieHB, GFuzz, GFuzzFlow, GFuzzHB, Flow", fm)
	}

	data.MaxNumberRuns = mRun
	if data.MaxTime > 0 {
		data.MaxTime = time.Duration(mTime) * time.Second
		data.MaxTimeSet = true
	}

	data.FuzzingMode = fm
	data.FuzzingModeGoPie = (data.FuzzingMode == data.GoPie || data.FuzzingMode == data.GoPiePlus || data.FuzzingMode == data.GoPieHB)
	data.FuzzingModeGFuzz = (data.FuzzingMode == data.GFuzz || data.FuzzingMode == data.GFuzzHBFlow || data.FuzzingMode == data.GFuzzHB)
	data.FuzzingModeFlow = (data.FuzzingMode == data.Flow || data.FuzzingMode == data.GFuzzHBFlow)
	data.UseHBInfoFuzzing = (data.FuzzingMode == data.GFuzzHB || data.FuzzingMode == data.GFuzzHBFlow || data.FuzzingMode == data.Flow || data.FuzzingMode == data.GoPiePlus || data.FuzzingMode == data.GoPieHB)

	data.CancelTestIfBugFound = cancelTestIfFound

	if cont {
		log.Info("Continue fuzzing")
	} else {
		log.Info("Start fuzzing")
	}

	// run either fuzzing on main or fuzzing on one test
	if modeMain || name != "" {
		if modeMain {
			log.Info("Run fuzzing on main function")
		} else {
			log.Info("Run fuzzing on test ", name)
		}

		clearData()
		timer.ResetFuzzing()
		control.Reset()

		err := runFuzzing(modeMain, advocate, progPath, progName, "", name, ignoreAtomic,
			meaTime, notExec, createStats, keepTraces, true, cont, 0, 0)

		if createStats {
			err := stats.CreateStatsFuzzing(data.GetPath(progPath), progName)
			if err != nil {
				log.Error("Failed to create fuzzing stats: ", err.Error())
			}
			err = stats.CreateStatsTotal(data.GetPath(progPath), progName)
			if err != nil {
				log.Error("Failed to create total stats: ", err.Error())
			}
		}

		return err
	}

	log.Info("Run fuzzing on all tests")

	// run fuzzing on all tests
	testFiles, maxFileNumber, totalFiles, err := toolchain.FindTestFiles(progPath, cont)
	if err != nil {
		return fmt.Errorf("Failed to find test files: %v", err)
	}

	log.Infof("Found %d test files", totalFiles)

	// Process each test file
	fileCounter := 0
	if cont {
		fileCounter = maxFileNumber
	}

	for i, testFile := range testFiles {
		fileCounter++
		log.Progressf("Progress %s: %d/%d\n", progName, fileCounter, totalFiles)
		log.Progressf("Processing file: %s\n", testFile)

		testFunctions, err := toolchain.FindTestFunctions(testFile)
		if err != nil || len(testFunctions) == 0 {
			log.Info("Could not find test functions in ", testFile)
			continue
		}

		for j, testFunc := range testFunctions {
			resetFuzzing()
			timer.ResetTest()
			control.Reset()

			timer.Start(timer.TotalTest)

			log.Progressf("Run fuzzing for %s->%s", testFile, testFunc)

			firstRun := (i == 0 && j == 0)

			err := runFuzzing(false, advocate, progPath, progName, testFile, testFunc, ignoreAtomic,
				meaTime, notExec, createStats, keepTraces, firstRun, cont, fileCounter, j+1)
			if err != nil {
				log.Error("Error in fuzzing: ", err.Error())
				clearData()
			}

			timer.Stop(timer.TotalTest)

			timer.UpdateTimeFileOverview(progName, testFunc)
		}

	}

	if createStats {
		err := stats.CreateStatsFuzzing(data.GetPath(progPath), progName)
		if err != nil {
			log.Error("Failed to create fuzzing stats: ", err.Error())
		}
		err = stats.CreateStatsTotal(data.GetPath(progPath), progName)
		if err != nil {
			log.Error("Failed to create total stats: ", err.Error())
		}
	}

	return nil
}

// Run Fuzzing on one program/test
//
// Parameter:
//   - modeMain bool: if true, run fuzzing on main function, otherwise on test
//   - advocate string: path to advocate
//   - progName string: name of the program
//   - testPath string: path to the test file
//   - name string: If modeMain, name of the executable, else name of the test
//   - ignoreAtomic bool: if true, ignore atomics for replay
//   - hBInfoFuzzing bool: whether to us HB info in fuzzing
//   - meaTime bool: measure runtime
//   - notExec bool: find never executed operations
//   - createStats bool: create statistics
//   - keepTraces bool: keep the traces after analysis
//   - skipExisting bool: skip existing runs
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//   - cont bool: continue with an already started run
func runFuzzing(modeMain bool, advocate, progPath, progName, testPath, name string, ignoreAtomic,
	meaTime, notExec, createStats, keepTraces, firstRun, cont bool, fileNumber, testNumber int) error {

	progDir := data.GetPath(progPath)

	clearDataFull()

	startTime := time.Now()

	// while there are available mutations, run them
	for data.NumberFuzzingRuns == 0 || len(data.MutationQueue) != 0 {

		// clean up
		clearData()
		timer.ResetFuzzing()
		control.Reset()

		if data.CancelTestIfBugFound && results.GetBugWasFound() {
			log.Resultf(false, false, "", "Cancel test after %d runs", data.NumberFuzzingRuns)
			break
		}

		log.Info("Fuzzing Run: ", data.NumberFuzzingRuns)

		fuzzingPath := ""
		progPathDir := helper.GetDirectory(progPath)
		var order data.Mutation
		if data.NumberFuzzingRuns != 0 {
			order = popMutation()
			if order.MutType == data.MutPiType {
				fuzzingPath = filepath.Join(progPathDir,
					filepath.Join("fuzzingTraces",
						fmt.Sprintf("fuzzingTrace_%d", order.MutPie)))
			} else {
				err := data.WriteMutationToFile(progPathDir, order)
				if err != nil {
					return err
				}
			}
		}

		firstRun = firstRun && (data.NumberFuzzingRuns == 0)

		// Run the test/mutation

		mode := "test"
		if modeMain {
			mode = "main"
		}
		err := toolchain.Run(mode, advocate, progPath, testPath, true, true, true,
			name, progName, name, data.NumberFuzzingRuns, fuzzingPath, ignoreAtomic,
			meaTime, notExec, createStats, keepTraces, false, firstRun, cont,
			fileNumber, testNumber)
		if err != nil {
			log.Error("Fuzzing run failed: ", err.Error())
			data.NumberFuzzingRuns++
		} else {
			data.NumberFuzzingRuns++

			// cancel if max number of mutations have been reached
			if data.MaxNumberRuns != -1 && data.NumberFuzzingRuns >= data.MaxNumberRuns {
				log.Infof("Finish fuzzing because maximum number of mutation runs (%d) have been reached", data.MaxNumberRuns)
				return nil
			}

			log.Info("Parse recorded trace to calculate fuzzing relations")

			// collect the required data to decide whether run is interesting
			// and to create the mutations
			ParseTrace(&anaData.MainTrace)

			if control.CheckCanceled() {
				log.Error("Fuzzing was canceled due to memory")
				continue
			}

			log.Infof("Create mutations")
			if data.FuzzingModeGFuzz {
				log.Infof("Create GFuzz mutations")
				gfuzz.CreateGFuzzMut()
			}

			// add new mutations based on flow path expansion
			if data.FuzzingModeFlow {
				log.Infof("Create Flow mutations")
				flow.CreateMutationsFlow()
			}

			// add mutations based on GoPie
			if data.FuzzingModeGoPie {
				log.Infof("Create GoPie mutations")
				gopie.CreateGoPieMut(progDir, data.NumberFuzzingRuns, order.MutPie)
			}

			log.Infof("Current fuzzing queue size: %d", len(data.MutationQueue))

			gfuzz.MergeTraceInfoIntoFileInfo()
		}

		anaData.ClearTrace()
		anaData.ClearData()

		if data.MaxTimeSet && time.Since(startTime) > data.MaxTime {
			log.Infof("Finish fuzzing because maximum runtime for fuzzing (%d min)has been reached", int(data.MaxTime.Minutes()))
			return nil
		}
	}

	if data.FuzzingModeGoPie {
		toolchain.ClearFuzzingTrace(progDir, keepTraces)
	}

	log.Infof("Finish fuzzing after %d runs\n", data.NumberFuzzingRuns)

	return nil
}

// Remove and return the first mutation from the mutation queue
//
// Returns:
//   - the first mutation from the mutation queue
func popMutation() data.Mutation {
	var mut data.Mutation
	mut, data.MutationQueue = data.MutationQueue[0], data.MutationQueue[1:]
	return mut
}

// Reset fuzzing
func resetFuzzing() {
	data.NumberFuzzingRuns = 0
	data.MutationQueue = make([]data.Mutation, 0)
	// count how often a specific mutation has been in the queue
	data.AllMutations = make(map[string]int)
}

func clearDataFull() {
	data.ClearDataFull()
	gopie.ClearData()
	gfuzz.ClearData()
	flow.ClearData()
}

func clearData() {
	gopie.ClearData()
	gfuzz.ClearData()
	flow.ClearData()
}
