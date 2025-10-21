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
	"advocate/analysis/baseA"
	"advocate/fuzzing/baseF"
	"advocate/fuzzing/flow"
	"advocate/fuzzing/gfuzz"
	"advocate/fuzzing/gopie"
	"advocate/fuzzing/guided"
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

	modes := []string{baseF.GoPie, baseF.GoCR, baseF.GoCRHB, baseF.GFuzz, baseF.GFuzzHBFlow, baseF.GFuzzHB, baseF.Flow, baseF.Guided, baseF.Default}
	if !types.Contains(modes, flags.FuzzingMode) {
		return fmt.Errorf("Invalid fuzzing mode '%s'. Possible values are GoPie, GoCR, GoCRHB, GFuzz, GFuzzFlow, GFuzzHB, Flow", flags.FuzzingMode)
	}

	baseF.MaxNumberRuns = flags.MaxFuzzingRun
	if flags.TimeoutFuzzing > 0 {
		baseF.MaxTime = time.Duration(flags.TimeoutFuzzing) * time.Second
		baseF.MaxTimeSet = true
	}

	baseF.FuzzingModeGoPie = (flags.FuzzingMode == baseF.GoPie || flags.FuzzingMode == baseF.GoCR || flags.FuzzingMode == baseF.GoCRHB)
	baseF.FuzzingModeGoCRHBPlus = (flags.FuzzingMode == baseF.GoCR || flags.FuzzingMode == baseF.GoCRHB)
	baseF.FuzzingModeGFuzz = (flags.FuzzingMode == baseF.GFuzz || flags.FuzzingMode == baseF.GFuzzHBFlow || flags.FuzzingMode == baseF.GFuzzHB)
	baseF.FuzzingModeFlow = (flags.FuzzingMode == baseF.Flow || flags.FuzzingMode == baseF.GFuzzHBFlow)
	baseF.FuzzingModeGuided = (flags.FuzzingMode == baseF.Guided || flags.FuzzingMode == baseF.Default)
	baseF.UseHBInfoFuzzing = (flags.FuzzingMode == baseF.GFuzzHB || flags.FuzzingMode == baseF.GFuzzHBFlow || flags.FuzzingMode == baseF.Flow || flags.FuzzingMode == baseF.GoCR || flags.FuzzingMode == baseF.GoCRHB)

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
			err := stats.CreateStatsFuzzing(baseF.GetPath(flags.ProgPath))
			if err != nil {
				log.Error("Failed to create fuzzing stats: ", err.Error())
			}
			err = stats.CreateStatsTotal(baseF.GetPath(flags.ProgPath))
			if err != nil {
				log.Error("Failed to create total stats: ", err.Error())
			}
		}

		clearDataFull()
		timer.ResetFuzzing()
		control.Reset()

		if !flags.KeepTraces {
			toolchain.RemoveTraces(flags.ProgPath)
		}

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
		err := stats.CreateStatsFuzzing(baseF.GetPath(flags.ProgPath))
		if err != nil {
			log.Error("Failed to create fuzzing stats: ", err.Error())
		}
		err = stats.CreateStatsTotal(baseF.GetPath(flags.ProgPath))
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

	progDir := baseF.GetPath(flags.ProgPath)

	clearDataFull()

	startTime := time.Now()

	// while there are available mutations, run them
	for baseF.NumberFuzzingRuns == 0 || len(baseF.MutationQueue) != 0 {

		// clean up
		clearDataRun()
		timer.ResetFuzzing()
		control.Reset()

		if flags.CancelTestIfBugFound && results.GetBugWasFound() {
			log.Infof("Cancel test after %d runs", baseF.NumberFuzzingRuns)
			break
		}

		log.Info("Fuzzing Run: ", baseF.NumberFuzzingRuns+1)

		fuzzingPath := ""
		progPathDir := helper.GetDirectory(flags.ProgPath)
		var order baseF.Mutation
		if baseF.NumberFuzzingRuns != 0 {
			order = popMutation()
			if order.MutType == baseF.MutPiType {
				fuzzingPath = filepath.Join(progPathDir,
					filepath.Join("fuzzingTraces",
						fmt.Sprintf("fuzzingTrace_%d", order.MutPie)))
			} else {
				err := baseF.WriteMutationToFile(progPathDir, order)
				if err != nil {
					return err
				}
			}
		}

		firstRun = firstRun && (baseF.NumberFuzzingRuns == 0)

		// Run the test/mutation

		mode := "test"
		if flags.ModeMain {
			mode = "main"
		}

		runAnalysis := true
		runRecord := true
		traceID, numberResults, err := toolchain.Run(mode, testPath, runRecord, runAnalysis, runAnalysis,
			baseF.NumberFuzzingRuns, fuzzingPath, firstRun, fileNumber, testNumber)

		baseF.NumberFuzzingRuns++

		if numberResults > flags.MaxNumberElements {
			continue
		}

		if err != nil {
			log.Error("Fuzzing run failed: ", err.Error())
		} else {
			log.Info("Parse recorded trace to calculate fuzzing relations")

			// collect the required data to decide whether run is interesting
			// and to create the mutations
			ParseTrace(&baseA.MainTrace)

			if control.CheckCanceled() {
				log.Error("Fuzzing run was canceled due to memory")
				gopie.ClearDataRun()
				baseA.ClearTrace()
				baseA.ClearData()
				continue
			}

			log.Infof("Create mutations")

			// add mutation based on guided fuzzing
			if baseF.FuzzingModeGuided {
				log.Infof("Create hb guided mutation")
				guided.CreateMutations()
			}

			// Add mutation based on GFuzz
			if baseF.FuzzingModeGFuzz {
				log.Infof("Create GFuzz mutations")
				gfuzz.CreateMutations()
			}

			// add new mutations based on flow path expansion
			if baseF.FuzzingModeFlow {
				log.Infof("Create Flow mutations")
				flow.CreateMutations()
			}

			// add mutations based on GoPie
			if baseF.FuzzingModeGoPie {
				log.Infof("Create GoPie mutations")
				gopie.CreateMutations(progDir, baseF.NumberFuzzingRuns, order.MutPie)
			}

			if flags.CreateStatistics {
				stats.CreateStats(flags.ExecName, traceID, baseF.NumberFuzzingRuns-1)
			}

			log.Infof("Current fuzzing queue size: %d", len(baseF.MutationQueue))

			if baseF.FuzzingModeGFuzz {
				gfuzz.MergeTraceInfoIntoFileInfo()
			}
		}

		// cancel if max number of mutations have been reached
		if baseF.MaxNumberRuns != -1 && baseF.NumberFuzzingRuns >= baseF.MaxNumberRuns {
			log.Infof("Finish fuzzing because maximum number of mutation runs (%d) have been reached", baseF.MaxNumberRuns)
			return nil
		}

		// cancel if max fuzzing time has been reached
		if baseF.MaxTimeSet && time.Since(startTime) > baseF.MaxTime {
			log.Infof("Finish fuzzing because maximum runtime for fuzzing (%d min) has been reached", int(baseF.MaxTime.Minutes()))
			return nil
		}

		// cancel if bug was found
		if baseF.FinishIfBugFound && numberResults > 0 {
			return nil
		}

		baseA.ClearTrace()
		baseA.ClearData()

	}

	if baseF.FuzzingModeGoPie {
		toolchain.ClearFuzzingTrace(progDir)
	}

	log.Infof("Finish fuzzing after %d runs\n", baseF.NumberFuzzingRuns)

	return nil
}

// Remove and return the first mutation from the mutation queue
//
// Returns:
//   - the first mutation from the mutation queue
func popMutation() baseF.Mutation {
	var mut baseF.Mutation
	mut, baseF.MutationQueue = baseF.MutationQueue[0], baseF.MutationQueue[1:]
	return mut
}

// Reset fuzzing
func resetFuzzing() {
	baseF.NumberFuzzingRuns = 0
	baseF.MutationQueue = make([]baseF.Mutation, 0)
	// count how often a specific mutation has been in the queue
	baseF.AllMutations = make(map[string]int)
}

func clearDataFull() {
	baseF.ClearDataFull()
	gopie.ClearData()
	gfuzz.ClearDataFull()
	flow.ClearData()
}

func clearDataRun() {
	gopie.ClearDataRun()
	gfuzz.ClearDataRun()
	flow.ClearData()
}
