//
// File: headerUnitTests.go
// Brief: Functions to add and remove the ADVOCATE header into file containing
//    unit tests
//
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"fmt"
	"goCR/analysis/data"
	"goCR/utils/helper"
)

var (
	currentResFolder = ""
)

// Run is the main function for the toolchain
//
// Parameter:
//   - mode string: mode of the toolchain (main or test or explain)
//   - goCR string: path to the root ADVOCATE folder.
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
//   - stats bool: create statistics
//   - keepTraces bool: keep the traces after analysis
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//   - cont bool: continue an already started run
//
// Returns:
//   - string: current result folder path
//   - int: TraceID
//   - int: number results
//   - error
func Run(mode, goCR, pathToMainFileOrTestDir, pathToTest string,
	runRecord, runAnalysis, runReplay bool,
	execName, progName, test string, fuzzing int, fuzzingTrace string,
	ignoreAtomic, meaTime, stats, keepTraces, skipExisting bool,
	firstRun, cont bool, fileNumber, testNumber int) (string, int, int, error) {
	pathToGoCR = helper.CleanPathHome(goCR)
	pathToFileOrDir = helper.CleanPathHome(pathToMainFileOrTestDir)

	executableName = execName
	programName = progName
	testName = test

	replayAtomic = !ignoreAtomic
	measureTime = meaTime
	createStats = stats

	data.Clear()

	switch mode {
	case "main":
		if pathToGoCR == "" {
			return "", 0, 0, fmt.Errorf("Path to goCR required for mode main")
		}
		if pathToFileOrDir == "" {
			return "", 0, 0, fmt.Errorf("Path to file required")
		}
		if executableName == "" {
			return "", 0, 0, fmt.Errorf("Name of the executable required")
		}
		if (stats || measureTime) && progName == "" {
			progName = helper.GetProgName(pathToMainFileOrTestDir)
		}
		return runWorkflowMain(pathToGoCR, pathToFileOrDir, runRecord, runAnalysis,
			executableName, keepTraces, fuzzing, fuzzingTrace, firstRun)
	case "test", "tests":
		if pathToGoCR == "" {
			return "", 0, 0, fmt.Errorf("Path to goCR required")
		}
		if pathToFileOrDir == "" {
			return "", 0, 0, fmt.Errorf("Path to test folder required for mode main")
		}
		if (stats || measureTime) && progName == "" {
			progName = helper.GetProgName(pathToMainFileOrTestDir)
		}
		return runWorkflowUnit(pathToGoCR, pathToFileOrDir, runRecord, runAnalysis,
			pathToTest, progName, stats && fuzzing == -1, fuzzing, fuzzingTrace, keepTraces,
			firstRun, skipExisting, cont, fileNumber, testNumber)
	default:
		return "", 0, 0, fmt.Errorf("Choose one mode from 'main' or 'test'")
	}
}
