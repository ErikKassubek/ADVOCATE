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
	"analyzer/analysis"
	"fmt"
	"os"
	"strings"
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
//   - execName string: name of the executable, only needed for mode main
//   - progName string: name of the program, used for stats
//   - test string: which test to run, if empty run all tests
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - replayAt bool: replay atomics
//   - meaTime bool: measure runtime
//   - notExec bool: find never executed operations
//   - stats bool: create statistics
//   - keepTraces bool: keep the traces after analysis
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//   - cont bool: continue an already started run
func Run(mode, advocate, pathToMainFileOrTestDir, pathToTest string,
	runRecord, runAnalysis, runReplay bool,
	execName, progName, test string, fuzzing int, fuzzingTrace string,
	ignoreAtomic, meaTime, notExec, stats, keepTraces, skipExisting bool,
	firstRun, cont bool, fileNumber, testNumber int) error {
	home, _ := os.UserHomeDir()
	pathToAdvocate = strings.Replace(advocate, "~", home, -1)
	pathToFileOrDir = strings.Replace(pathToMainFileOrTestDir, "~", home, -1)

	executableName = execName
	programName = progName
	testName = test

	replayAtomic = !ignoreAtomic
	measureTime = meaTime
	notExecuted = notExec
	createStats = stats

	analysis.Clear()

	switch mode {
	case "main":
		if pathToAdvocate == "" {
			return fmt.Errorf("Path to advocate required for mode main")
		}
		if pathToFileOrDir == "" {
			return fmt.Errorf("Path to file required")
		}
		if executableName == "" {
			return fmt.Errorf("Name of the executable required")
		}
		if (stats || measureTime) && progName == "" {
			return fmt.Errorf("If -stats or -recordTime is set, -prog [name] must be set as well")
		}
		return runWorkflowMain(pathToAdvocate, pathToFileOrDir, runRecord, runAnalysis, runReplay,
			executableName, keepTraces, fuzzing, fuzzingTrace, firstRun)
	case "test", "tests":
		if pathToAdvocate == "" {
			return fmt.Errorf("Path to advocate required")
		}
		if pathToFileOrDir == "" {
			return fmt.Errorf("Path to test folder required for mode main")
		}
		if (stats || measureTime) && progName == "" {
			return fmt.Errorf("If -stats or -recordTime is set, -prog [name] must be set as well")
		}
		return runWorkflowUnit(pathToAdvocate, pathToFileOrDir, runRecord, runAnalysis, runReplay,
			pathToTest, progName, notExecuted, stats, fuzzing, fuzzingTrace, keepTraces,
			firstRun, skipExisting, cont, fileNumber, testNumber)
	default:
		return fmt.Errorf("Choose one mode from 'main' or 'test'")
	}
}
