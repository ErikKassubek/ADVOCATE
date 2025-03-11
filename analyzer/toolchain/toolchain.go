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
	runAnalyzer func(pathTrace string, noRewrite bool, analysisCases map[string]bool, outReadable string, outMachine string, ignoreAtomics bool, fifo bool, ignoreCriticalSection bool, rewriteAll bool, newTrace string, ignoreRewrite string, fuzzing int, onlyAPanicAndLeak bool) error
)

/*
 * Function injection for modeAnalyzer
 */
func InitFuncAnalyzer(funcAnalyzer func(pathTrace string,
	noRewrite bool, analysisCases map[string]bool, outReadable string, outMachine string,
	ignoreAtomics bool, fifo bool, ignoreCriticalSection bool,
	rewriteAll bool, newTrace string, ignoreRewrite string, fuzzing int, onlyAPanicAndLeak bool) error) {
	runAnalyzer = funcAnalyzer
}

/*
 * Main function for the toolchain
 * Args:
 * 	mode (string): mode of the toolchain (main or test or explain)
 * 	advocate (string): path to the root ADVOCATE folder.
 * 	file (string): if mode is main, path to main file, if mode test, path to test folder
 * 	execName (string): name of the executable, only needed for mode main
 * 	progName (string): name of the program, used for stats
 * 	test (string): which test to run, if empty run all tests
 * 	numRerecorded (int): limit of number of rerecordings
 * 	fuzzing (int): -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
 * 	replayAt (bool): replay atomics
 * 	meaTime (bool): measure runtime
 * 	notExec (bool): find never executed operations
 * 	stats (bool): create statistics
 * 	keepTraces (bool): keep the traces after analysis
 * 	firstRun (bool): this is the first run, only set to false for fuzzing (except for the first fuzzing)
 * 	cont (bool): continue an already started run
 */
func Run(mode, advocate, file, execName, progName, test string,
	numRerecorded, fuzzing int,
	ignoreAtomic, meaTime, notExec, stats, keepTraces bool, firstRun, cont bool) error {
	home, _ := os.UserHomeDir()
	pathToAdvocate = strings.Replace(advocate, "~", home, -1)
	pathToFile = strings.Replace(file, "~", home, -1)

	executableName = execName
	programName = progName
	testName = test

	numberRerecord = numRerecorded

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
		if pathToFile == "" {
			return fmt.Errorf("Path to file required")
		}
		if executableName == "" {
			return fmt.Errorf("Name of the executable required")
		}
		if (stats || measureTime) && progName == "" {
			return fmt.Errorf("If -scen or -trace is set, -prog [name] must be set as well")
		}
		return runWorkflowMain(pathToAdvocate, pathToFile, executableName, keepTraces, fuzzing, firstRun)
	case "test", "tests":
		if pathToAdvocate == "" {
			return fmt.Errorf("Path to advocate required")
		}
		if pathToFile == "" {
			return fmt.Errorf("Path to test folder required for mode main")
		}
		if (stats || measureTime) && progName == "" {
			return fmt.Errorf("If -scen or -trace is set, -prog [name] must be set as well")
		}
		return runWorkflowUnit(pathToAdvocate, pathToFile, progName, measureTime,
			notExecuted, stats, fuzzing, keepTraces, firstRun, cont)
	case "explain":
		if pathToAdvocate == "" {
			return fmt.Errorf("Path to advocate required")
		}
		if pathToFile == "" {
			fmt.Println("Path to test folder required for mode main")
		}
		generateBugReports(pathToFile, fuzzing)
	default:
		return fmt.Errorf("Choose one mode from 'main' or 'test' or 'explain'")
	}

	return nil
}

func getAbsolutPath(path string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(path, "~", home, -1)
}
