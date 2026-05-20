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
	"fmt"
	"gocdr/analysis/analysis"
	"gocdr/analysis/baseA"
	"gocdr/io"
	"gocdr/results/results"
	"gocdr/utils/control"
	"gocdr/utils/log"
)

// runAnalyzer is the starting point to the analyzer.
// This function will read the trace at a stored path, analyze it and,
// if needed, rewrite the trace.
//
// Parameter:
//   - pathTrace string: path to the trace to be analyzed
//   - outReadable string: path to the readable result file
//   - outMachine string: path to the machine result file
//   - fifo bool: assume, that the channels work as a fifo queue
//   - newTrace string: path to where the rewritten trace should be created
//   - fuzzingRun int: number of fuzzing run (0 for recording, then always add 1)
//
// Returns:
//   - error
func runAnalyzer(pathTrace string,
	outReadable string, outMachine string,
	newTrace string, fuzzingRun int) error {

	if pathTrace == "" {
		return fmt.Errorf("Please provide a path to the trace files. Set with -trace [folder]")
	}

	// run the analysis and, if requested, create a reordered trace file
	// based on the analysis results

	results.InitResults(outReadable, outMachine)

	numberOfRoutines, numberElems, err := io.CreateTraceFromFiles(pathTrace)

	if err != nil && fuzzingRun <= 0 {
		// log.Error("Could not read trace: ", err.Error())
		return err
	}

	if numberElems == 0 {
		log.Infof("Trace at %s does not contain any elements", pathTrace)
	}

	log.Infof("Read trace with %d elements in %d routines", numberElems, numberOfRoutines)

	analysis.RunAnalysis(fuzzingRun >= 0)

	if control.WasCanceled() {
		// analysis.LogSizes()
		if control.WasCanceledRAM() {
			baseA.Clear()
			return fmt.Errorf("Analysis was canceled due to insufficient RAM")
		}
		baseA.Clear()
		return fmt.Errorf("Analysis was canceled due to unexpected panic")
	}
	log.Info("Analysis finished")

	_, err = results.CreateResultFiles(true)
	if err != nil {
		log.Error("Error in printing summary: ", err.Error())
	}

	return nil
}
