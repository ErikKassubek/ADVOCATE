// Copyright (c) 2024 Erik Kassubek
//
// File: headerUnitTests.go
// Brief: Functions to add and remove the ADVOCATE header into file containing
//    unit tests
//
// Author: Erik Kassubek
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/analysis/data"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"fmt"
)

var (
	currentResFolder = ""
)

// Run is the main function for the toolchain
//
// Parameter:
//   - mode string: mode of the toolchain (main or test or explain)
//   - advocate string: path to the root ADVOCATE folder.
//   - pathToMainFileOrTestDir string: if mode is main, path to main file, if mode test, path to test folder
//   - pathToTest string: specify specific test path, only used for fuzzing
//   - runRecord bool: run the recording. If set to false, but runAnalysis or runReplay is
//     set the trace at tracePath is used
//   - runAnalysis bool: run the analysis on a path
//   - runReplay bool: run replay, if runAnalysis is true, those replays are used
//     otherwise the trace at tracePath is replayed
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - replayAt bool: replay atomics
//   - stats bool: create statistics
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//
// Returns:
//   - string: current result folder path
//   - int: TraceID
//   - int: number results
//   - error
func Run(mode, advocate, pathToTest string,
	runRecord, runAnalysis, runReplay bool, fuzzing int, fuzzingTrace string,
	firstRun bool, fileNumber, testNumber int) (string, int, int, error) {
	pathToAdvocate = helper.CleanPathHome(advocate)
	pathToFileOrDir = helper.CleanPathHome(flags.ProgPath)

	data.Clear()

	switch mode {
	case "main":
		if pathToAdvocate == "" {
			return "", 0, 0, fmt.Errorf("Path to advocate required for mode main")
		}
		if pathToFileOrDir == "" {
			return "", 0, 0, fmt.Errorf("Path to file required")
		}
		if flags.ExecName == "" {
			return "", 0, 0, fmt.Errorf("Name of the executable required")
		}
		if (flags.CreateStatistics || flags.MeasureTime) && flags.ProgName == "" {
			flags.ProgName = helper.GetProgName(flags.ProgPath)
		}
		return runWorkflowMain(pathToAdvocate, pathToFileOrDir, runRecord, runAnalysis, runReplay,
			fuzzing, fuzzingTrace, firstRun)
	case "test", "tests":
		if pathToAdvocate == "" {
			return "", 0, 0, fmt.Errorf("Path to advocate required")
		}
		if pathToFileOrDir == "" {
			return "", 0, 0, fmt.Errorf("Path to test folder required for mode main")
		}
		if (flags.CreateStatistics || flags.MeasureTime) && flags.ProgName == "" {
			flags.ProgName = helper.GetProgName(flags.ProgPath)
		}
		return runWorkflowUnit(pathToAdvocate, pathToFileOrDir, runRecord, runAnalysis, runReplay,
			pathToTest, flags.CreateStatistics && fuzzing == -1, fuzzing, fuzzingTrace,
			firstRun, fileNumber, testNumber)
	default:
		return "", 0, 0, fmt.Errorf("Choose one mode from 'main' or 'test'")
	}
}
