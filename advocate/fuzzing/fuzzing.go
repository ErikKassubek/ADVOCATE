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
	"advocate/fuzzing/partialorder"
	"advocate/results/results"
	"advocate/results/stats"
	"advocate/toolchain"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
	"advocate/utils/types"
	"fmt"
	"path/filepath"
	"time"
)

// Fuzzing creates the fuzzing data and runs the fuzzing executions
func Fuzzing() error {

	if flags.FuzzingMode == "" {
		return fmt.Errorf("No fuzzing mode selected. Select with -fuzzingMode [mode]. Possible values are GoPie, GoCR, GoCRHB, GFuzz, GFuzzFlow, GFuzzHB, Flow")
	}

	modes := []string{data.GoPie, data.GoCR, data.GoCRHB, data.GFuzz, data.GFuzzHBFlow, data.GFuzzHB, data.Flow}
	if !types.Contains(modes, flags.FuzzingMode) {
		return fmt.Errorf("Invalid fuzzing mode '%s'. Possible values are GoPie, GoCR, GoCRHB, GFuzz, GFuzzFlow, GFuzzHB, Flow", flags.FuzzingMode)
	}

	data.MaxNumberRuns = flags.MaxFuzzingRun
	if flags.TimeoutFuzzing > 0 {
		data.MaxTime = time.Duration(flags.TimeoutFuzzing) * time.Second
		data.MaxTimeSet = true
	}

	data.FuzzingModeGoPie = (flags.FuzzingMode == data.GoPie || flags.FuzzingMode == data.GoCR || flags.FuzzingMode == data.GoCRHB)
	data.FuzzingModeGoCRHBPlus = (flags.FuzzingMode == data.GoCR || flags.FuzzingMode == data.GoCRHB)
	data.FuzzingModeGFuzz = (flags.FuzzingMode == data.GFuzz || flags.FuzzingMode == data.GFuzzHBFlow || flags.FuzzingMode == data.GFuzzHB)
	data.FuzzingModeFlow = (flags.FuzzingMode == data.Flow || flags.FuzzingMode == data.GFuzzHBFlow)
	data.UseHBInfoFuzzing = (flags.FuzzingMode == data.GFuzzHB || flags.FuzzingMode == data.GFuzzHBFlow || flags.FuzzingMode == data.Flow || flags.FuzzingMode == data.GoCR || flags.FuzzingMode == data.GoCRHB)

	if flags.Continue {
		log.Info("Continue fuzzing")
	} else {
		log.Infof("Start fuzzing in mode %s", flags.FuzzingMode)
	}

	// run either fuzzing on main or fuzzing on one test
	if flags.ModeMain || flags.ExecName != "" {
		if flags.ModeMain {
			log.Info("Run fuzzing on main function")
		} else {
			log.Info("Run fuzzing on test ", flags.ExecName)
		}

		err := runFuzzing("", true, 0, 0)

		if flags.CreateStatistics {
			err := stats.CreateStatsFuzzing(data.GetPath(flags.ProgPath))
			if err != nil {
				log.Error("Failed to create fuzzing stats: ", err.Error())
			}
			err = stats.CreateStatsTotal(data.GetPath(flags.ProgPath))
			if err != nil {
				log.Error("Failed to create total stats: ", err.Error())
			}
		}

		clearDataFull()
		timer.ResetFuzzing()
		control.Reset()

		return err
	}

	log.Info("Run fuzzing on all tests")

	// run fuzzing on all tests
	testFiles, maxFileNumber, totalFiles, err := toolchain.FindTestFiles(flags.ProgPath, flags.Continue)
	if err != nil {
		return fmt.Errorf("Failed to find test files: %v", err)
	}

	log.Infof("Found %d test files", totalFiles)

	// Process each test file
	fileCounter := 0
	if flags.Continue {
		fileCounter = maxFileNumber
	}

	for i, testFile := range testFiles {
		fileCounter++
		log.Progressf("Progress %s: %d/%d\n", flags.ProgName, fileCounter, totalFiles)
		log.Infof("Processing file: %s\n", testFile)

		testFunctions, err := toolchain.FindTestFunctions(testFile)
		if err != nil || len(testFunctions) == 0 {
			log.Info("Could not find test functions in ", testFile)
			continue
		}

		for j, testFunc := range testFunctions {
			flags.ExecName = testFunc

			resetFuzzing()
			timer.ResetTest()
			control.Reset()

			timer.Start(timer.TotalTest)

			log.Infof("Run fuzzing for %s->%s", testFile, testFunc)

			firstRun := (i == 0 && j == 0)

			err := runFuzzing(testFile, firstRun, fileCounter, j+1)
			if err != nil {
				log.Error("Error in fuzzing: ", err.Error())
				clearDataRun()
			}

			timer.Stop(timer.TotalTest)

			timer.UpdateTimeFileOverview(testFunc)
		}
	}

	if flags.CreateStatistics {
		err := stats.CreateStatsFuzzing(data.GetPath(flags.ProgPath))
		if err != nil {
			log.Error("Failed to create fuzzing stats: ", err.Error())
		}
		err = stats.CreateStatsTotal(data.GetPath(flags.ProgPath))
		if err != nil {
			log.Error("Failed to create total stats: ", err.Error())
		}
	}

	if !flags.KeepTraces {
		toolchain.RemoveTraces(flags.ProgPath)
	}

	return nil
}

// Run Fuzzing on one program/test
//
// Parameter:
//   - testPath string: path to the test file
//   - hBInfoFuzzing bool: whether to us HB info in fuzzing
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
func runFuzzing(testPath string, firstRun bool, fileNumber, testNumber int) error {

	progDir := data.GetPath(flags.ProgPath)

	clearDataFull()

	startTime := time.Now()

	// while there are available mutations, run them
	for data.NumberFuzzingRuns == 0 || len(data.MutationQueue) != 0 {

		// clean up
		clearDataRun()
		timer.ResetFuzzing()
		control.Reset()

		if flags.CancelTestIfBugFound && results.GetBugWasFound() {
			log.Infof("Cancel test after %d runs", data.NumberFuzzingRuns)
			break
		}

		log.Info("Fuzzing Run: ", data.NumberFuzzingRuns)

		fuzzingPath := ""
		progPathDir := helper.GetDirectory(flags.ProgPath)
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
		if flags.ModeMain {
			mode = "main"
		}

		runAnalysis := true
		runRecord := true
		traceID, numberResults, err := toolchain.Run(mode, testPath, runRecord, runAnalysis, runAnalysis,
			data.NumberFuzzingRuns, fuzzingPath, firstRun, fileNumber, testNumber)

		data.NumberFuzzingRuns++

		if numberResults > flags.MaxNumberElements {
			continue
		}

		if err != nil {
			log.Error("Fuzzing run failed: ", err.Error())
		} else {
			log.Info("Parse recorded trace to calculate fuzzing relations")

			// collect the required data to decide whether run is interesting
			// and to create the mutations
			ParseTrace(&anaData.MainTrace)

			if control.CheckCanceled() {
				log.Error("Fuzzing run was canceled due to memory")
				gopie.ClearDataRun()
				anaData.ClearTrace()
				anaData.ClearData()
				continue
			}

			log.Infof("Create mutations")

			// Add mutation based on happens before relations and predictive analysis
			if data.FuzzingHbAnalysis {
				log.Infof("Check for mutation based on happens before analysis")
				partialorder.CreateMutations()
			}

			// Add mutation based on GFuzz
			if data.FuzzingModeGFuzz {
				log.Infof("Create GFuzz mutations")
				gfuzz.CreateMutations()
			}

			// add new mutations based on flow path expansion
			if data.FuzzingModeFlow {
				log.Infof("Create Flow mutations")
				flow.CreateMutations()
			}

			// add mutations based on GoPie
			if data.FuzzingModeGoPie {
				log.Infof("Create GoPie mutations")
				gopie.CreateMutations(progDir, data.NumberFuzzingRuns, order.MutPie)
			}

			if flags.CreateStatistics {
				stats.CreateStats(flags.ExecName, traceID, data.NumberFuzzingRuns-1)
			}

			log.Infof("Current fuzzing queue size: %d", len(data.MutationQueue))

			gfuzz.MergeTraceInfoIntoFileInfo()
		}

		// cancel if max number of mutations have been reached
		if data.MaxNumberRuns != -1 && data.NumberFuzzingRuns >= data.MaxNumberRuns {
			log.Infof("Finish fuzzing because maximum number of mutation runs (%d) have been reached", data.MaxNumberRuns)
			return nil
		}

		// cancel if max fuzzing time has been reached
		if data.MaxTimeSet && time.Since(startTime) > data.MaxTime {
			log.Infof("Finish fuzzing because maximum runtime for fuzzing (%d min) has been reached", int(data.MaxTime.Minutes()))
			return nil
		}

		// cancel if bug was found
		if data.FinishIfBugFound && numberResults > 0 {
			return nil
		}

		anaData.ClearTrace()
		anaData.ClearData()

	}

	if data.FuzzingModeGoPie {
		toolchain.ClearFuzzingTrace(progDir)
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
	gfuzz.ClearDataFull()
	flow.ClearData()
}

func clearDataRun() {
	gopie.ClearDataRun()
	gfuzz.ClearDataRun()
	flow.ClearData()
}
