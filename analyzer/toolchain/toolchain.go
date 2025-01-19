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
	pathToAdvocate string
	pathToFile     string
	programName    string
	executableName string
	testName       string
	timeoutAna     int
	timeoutReplay  int
	numberRerecord int
	replayAtomic   bool
	measureTime    bool
	notExecuted    bool
	createStats    bool
	runAnalyzer    func(pathTrace string, noPrint bool, noRewrite bool, scenarios string, outReadable string, outMachine string, ignoreAtomics bool, fifo bool, ignoreCriticalSection bool, noWarning bool, rewriteAll bool, newTrace string, timeout int, ignoreRewrite string, fuzzing int)
)

func InitFuncAnalyzer(funcAnalyzer func(pathTrace string, noPrint bool,
	noRewrite bool, scenarios string, outReadable string, outMachine string,
	ignoreAtomics bool, fifo bool, ignoreCriticalSection bool, noWarning bool,
	rewriteAll bool, newTrace string, timeout int, ignoreRewrite string, fuzzing int)) {
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
 * 	timeoutA (int): timeout for analysis
 * 	timeoutR (int): timeout for replay
 * 	numRerecorded (int): limit of number of rerecordings
 * 	fuzzing (int): -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
 * 	replayAt (bool): replay atomics
 * 	meaTime (bool): measure runtime
 * 	notExec (bool): find never executed operations
 * 	stats (bool): create statistics
 * 	keepTraces (bool): keep the traces after analysis
 */
func Run(mode, advocate, file, execName, progName, test string,
	timeoutA, timeoutR, numRerecorded, fuzzing int,
	replayAt, meaTime, notExec, stats, keepTraces bool) error {
	home, _ := os.UserHomeDir()
	pathToAdvocate = strings.Replace(advocate, "~", home, -1)
	pathToFile = strings.Replace(file, "~", home, -1)

	executableName = execName
	programName = progName
	testName = test

	timeoutAna = timeoutA
	timeoutR = timeoutReplay
	numberRerecord = numRerecorded

	replayAtomic = replayAt
	measureTime = meaTime
	notExecuted = notExec
	createStats = stats

	analysis.ClearTrace()
	analysis.ClearData()

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
		return runWorkflowMain(pathToAdvocate, pathToFile, executableName, timeoutAna, timeoutReplay, keepTraces, fuzzing)
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
			notExecuted, stats, fuzzing, timeoutAna, timeoutReplay,
			keepTraces)
	case "explain":
		if pathToAdvocate == "" {
			return fmt.Errorf("Path to advocate required")
		}
		if pathToFile == "" {
			fmt.Println("Path to test folder required for mode main")
		}
		generateBugReports(pathToFile)
	default:
		return fmt.Errorf("Choose one mode from 'main' or 'test' or 'explain'")
	}

	return nil
}

func getAbsolutPath(path string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(path, "~", home, -1)
}
