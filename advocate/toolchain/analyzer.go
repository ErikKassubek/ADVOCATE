// Copyright (c) 2025 Erik Kassubek
//
// File: analyzer.go
// Brief: Starting point for the analyzer and replay
//
// Author: Erik Kassubek
// Created: 2025-04-26
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/analysis"
	"advocate/io"
	"advocate/memory"
	"advocate/results"
	"advocate/rewriter"
	"advocate/timer"
	"advocate/utils"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// runAnalyzer is the starting point to the analyzer.
// This function will read the trace at a stored path, analyze it and,
// if needed, rewrite the trace.
//
// Parameter:
//   - pathTrace string: path to the trace to be analyzed
//   - noRewrite bool: if set, rewrite is disabled
//   - analysisCases map[string]bool: map of analysis cases to run
//   - outReadable string: path to the readable result file
//   - outMachine string: path to the machine result file
//   - ignoreAtomics bool: if true, atomics are ignored for replay
//   - fifo bool: assume, that the channels work as a fifo queue
//   - ignoreCriticalSection bool: ignore the ordering of lock/unlock for the hb analysis
//   - rewriteAll bool: rewrite bugs that have been rewritten before
//   - newTrace string: path to where the rewritten trace should be created
//   - fuzzingRun int: number of fuzzing run (0 for recording, then always add 1)
//   - onlyAPanicAndLeak bool: only check for actual leaks and panics, do not calculate HB information
//
// Returns:
//   - error
func runAnalyzer(pathTrace string, noRewrite bool,
	analysisCases map[string]bool, outReadable string, outMachine string,
	ignoreAtomics bool, fifo bool, ignoreCriticalSection bool,
	rewriteAll bool, newTrace string, fuzzingRun int, onlyAPanicAndLeak bool) error {

	if pathTrace == "" {
		return fmt.Errorf("Please provide a path to the trace files. Set with -trace [folder]")
	}

	// run the analysis and, if requested, create a reordered trace file
	// based on the analysis results

	results.InitResults(outReadable, outMachine)

	numberOfRoutines, numberElems, err := io.CreateTraceFromFiles(pathTrace, ignoreAtomics)

	if err != nil {
		utils.LogError("Could not open trace: ", err.Error())
	}

	if numberElems == 0 {
		utils.LogInfof("Trace at %s does not contain any elements", pathTrace)
		return nil
	}

	utils.LogInfof("Read trace with %d elements in %d routines", numberElems, numberOfRoutines)

	if onlyAPanicAndLeak {
		utils.LogInfo("Start Analysis for actual panics and leaks")
	} else if analysisCases["all"] {
		utils.LogInfo("Start Analysis for all scenarios")
	} else {
		info := "Start Analysis for the following scenarios: "
		for key, value := range analysisCases {
			if value {
				info += (key + ",")
			}
		}
		utils.LogInfo(info)
	}

	analysis.RunAnalysis(fifo, ignoreCriticalSection, analysisCases, fuzzingRun >= 0, onlyAPanicAndLeak)

	if memory.WasCanceled() {
		// analysis.LogSizes()
		analysis.Clear()
		if memory.WasCanceledRAM() {
			return fmt.Errorf("Analysis was canceled due to insufficient small RAM")
		}
		return fmt.Errorf("Analysis was canceled due to unexpected panic")
	}
	utils.LogInfo("Analysis finished")

	numberOfResults, err := results.CreateResultFiles(noWarningFlag, true)
	if err != nil {
		utils.LogError("Error in printing summary: ", err.Error())
	}

	if noRewrite {
		utils.LogInfo("Skip rewrite")
		return nil
	}

	numberRewrittenTrace := 0
	failedRewrites := 0
	notNeededRewrites := 0

	if numberOfResults != 0 {
		utils.LogInfo("Start rewriting")
	}

	if err != nil {
		utils.LogError("Failed to create result files: ", err)
		return nil
	}

	rewrittenBugs := make(map[utils.ResultType][]string) // bugtype -> paths string

	file := filepath.Base(pathTrace)
	rewriteNr := "0"
	spl := strings.Split(file, "_")
	if len(spl) > 1 {
		rewriteNr = spl[len(spl)-1]
	}

	for resultIndex := 0; resultIndex < numberOfResults; resultIndex++ {
		needed, err := rewriteTrace(outMachine,
			newTrace+"_"+strconv.Itoa(resultIndex+1)+"/", resultIndex, numberOfRoutines, &rewrittenBugs)

		if !needed {
			notNeededRewrites++
			fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
		} else if err != nil {
			failedRewrites++
			fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
		} else { // needed && err == nil
			numberRewrittenTrace++
			fmt.Printf("Bugreport info: %s_%d,suc\n", rewriteNr, resultIndex+1)
		}

		if memory.WasCanceled() {
			failedRewrites += max(0, numberOfResults-resultIndex-1)
			break
		}
	}
	if memory.WasCanceledRAM() {
		utils.LogError("Rewrite Canceled: Not enough RAM")
	} else {
		utils.LogInfo("Finished Rewrite")
	}
	utils.LogInfo("Number Results: ", numberOfResults)
	utils.LogInfo("Successfully rewrites: ", numberRewrittenTrace)
	utils.LogInfo("No need/not possible to rewrite: ", notNeededRewrites)
	if failedRewrites > 0 {
		utils.LogInfo("Failed rewrites: ", failedRewrites)
	} else {
		utils.LogInfo("Failed rewrites: ", failedRewrites)
	}

	return nil
}

// Rewrite the trace file based on given analysis results
//
// Parameter:
//   - outMachine string: The path to the analysis result file
//   - newTrace string: The path where the new traces folder will be created
//   - resultIndex int: The index of the result to use for the reordered trace file
//   - numberOfRoutines int: The number of routines in the trace
//   - rewrittenTrace *map[utils.ResultType][]string: set of bugs that have been already rewritten
//
// Returns:
//   - bool: true, if a rewrite was necessary, false if not (e.g. actual bug, warning)
//   - error: An error if the trace file could not be created
func rewriteTrace(outMachine string, newTrace string, resultIndex int,
	numberOfRoutines int, rewrittenTrace *map[utils.ResultType][]string) (bool, error) {
	timer.Start(timer.Rewrite)
	defer timer.Stop(timer.Rewrite)

	actual, bug, err := io.ReadAnalysisResults(outMachine, resultIndex)
	if err != nil {
		return false, err
	}

	if actual {
		return false, nil
	}

	// the same bug was found and confirmed by replay in an earlier run,
	// either in fuzzing or in another test
	// It is therefore not needed to rewrite it again
	if !rewriteAllFlag && results.WasAlreadyConfirmed(bug.GetBugString()) {
		return false, nil
	}

	traceCopy := analysis.CopyMainTrace()

	rewriteNeeded, code, err := rewriter.RewriteTrace(&traceCopy, bug, *rewrittenTrace)

	if err != nil {
		return rewriteNeeded, err
	}

	err = io.WriteTrace(&traceCopy, newTrace, true)
	if err != nil {
		return rewriteNeeded, err
	}

	err = io.WriteRewriteInfoFile(newTrace, bug, code, resultIndex)
	if err != nil {
		return rewriteNeeded, err
	}

	return rewriteNeeded, nil
}
