//
// File: analyzer.go
// Brief: Starting point for the analyzer and replay
//
// Created: 2025-04-26
//
// License: BSD-3-Clause

package toolchain

import (
	"fmt"
	"goCR/analysis/analysis"
	"goCR/analysis/data"
	"goCR/io"
	"goCR/results/results"
	"goCR/utils/control"
	"goCR/utils/log"
)

// runAnalyzer is the starting point to the analyzer.
// This function will read the trace at a stored path, analyze it and,
// if needed, rewrite the trace.
//
// Parameter:
//   - pathTrace string: path to the trace to be analyzed
//   - outReadable string: path to the readable result file
//   - outMachine string: path to the machine result file
//   - ignoreAtomics bool: if true, atomics are ignored for replay
//   - fifo bool: assume, that the channels work as a fifo queue
//   - ignoreCriticalSection bool: ignore the ordering of lock/unlock for the hb analysis
//   - fuzzingRun int: number of fuzzing run (0 for recording, then always add 1)
//
// Returns:
//   - error
func runAnalyzer(pathTrace string,
	outReadable string, outMachine string,
	ignoreAtomics bool, fifo bool, ignoreCriticalSection bool,
	fuzzingRun int) error {

	if pathTrace == "" {
		return fmt.Errorf("Please provide a path to the trace files. Set with -trace [folder]")
	}

	// run the analysis and, if requested, create a reordered trace file
	// based on the analysis results

	results.InitResults(outReadable, outMachine)

	numberOfRoutines, numberElems, err := io.CreateTraceFromFiles(pathTrace, ignoreAtomics)

	// if err != nil && fuzzingRun <= 0 {
	// 	log.Error("Could not read trace: ", err.Error())
	// 	return err
	// }

	if numberElems == 0 {
		log.Infof("Trace at %s does not contain any elements", pathTrace)
	}

	log.Infof("Read trace with %d elements in %d routines", numberElems, numberOfRoutines)

	log.Info("Start Analysis")

	analysis.RunAnalysis(fifo, ignoreCriticalSection, fuzzingRun >= 0)

	if control.CheckCanceled() {
		// analysis.LogSizes()
		data.Clear()
		if control.CheckCanceledRAM() {
			return fmt.Errorf("Analysis was canceled due to insufficient small RAM")
		}
		return fmt.Errorf("Analysis was canceled due to unexpected panic")
	}
	log.Info("Analysis finished")

	_, err = results.CreateResultFiles(noWarningFlag, true)
	if err != nil {
		log.Error("Error in printing summary: ", err.Error())
	}

	return nil
}
