// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: runFullWorkflowMain.go
// Brief: Function to run the whole ADVOCATE workflow, including running,
//    analysis and replay on a program with a main function
//
// Author: Erik Kassubek, Mario Occhinegro
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"analyzer/complete"
	"analyzer/stats"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

/*
 * Run ADVOCATE on a program with a main function
 * Args:
 *    pathToAdvocate (string): path to the ADVOCATE folder
 *    pathToFile (string): path to the file containing the main function
 *    executableName (string): name of the executable
 *    timeoutAna (int): timeout for the analyzer
 *    timeoutReplay (int): timeout for replay
 *    keepTraces (bool): do not delete the traces after analysis
 *    fuzzing (int): -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
 * Returns:
 *    error
 */
func runWorkflowMain(pathToAdvocate string, pathToFile string, executableName string,
	timeoutAna int, timeoutReplay int, keepTraces bool, fuzzing int) error {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", pathToFile)
	}

	pathToPatchedGoRuntime := filepath.Join(pathToAdvocate, "go-patch/bin/go")

	if runtime.GOOS == "windows" {
		pathToPatchedGoRuntime += ".exe"
	}

	pathToGoRoot := filepath.Join(pathToAdvocate, "go-patch")
	pathToAnalyzer := filepath.Join(pathToAdvocate, "analyzer/analyzer")

	// Change to the directory of the main file
	dir := filepath.Dir(pathToFile)
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("Failed to change directory: %v", err)
	}
	fmt.Printf("In directory: %s\n", dir)

	os.RemoveAll("advocateResult")
	if err := os.MkdirAll("advocateResult", os.ModePerm); err != nil {
		return fmt.Errorf("Failed to create advocateResult directory: %v", err)
	}

	fmt.Println("Run program and analysis...")

	output := filepath.Join(dir, "output.log")
	outFile, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open log file: %v", err)
	}
	defer outFile.Close()

	origStdout := os.Stdout
	origStderr := os.Stderr

	os.Stdout = outFile
	os.Stderr = outFile

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Set GOROOT environment variable
	if err := os.Setenv("GOROOT", pathToGoRoot); err != nil {
		return fmt.Errorf("Failed to set GOROOT: %v", err)
	}
	fmt.Println("GOROOT exported")
	// Unset GOROOT
	defer os.Unsetenv("GOROOT")

	// Remove header
	if err := headerRemoverMain(pathToFile); err != nil {
		return fmt.Errorf("Error removing header: %v", err)
	}

	var durationRun time.Duration
	var durationRecord time.Duration
	var durationAnalysis time.Duration
	var durationReplay time.Duration

	// build the program
	if measureTime {
		fmt.Printf("%s build\n", pathToPatchedGoRuntime)
		if err := runCommand(pathToPatchedGoRuntime, "build"); err != nil {
			log.Println("Error in building program, removing header and stopping workflow")
			headerRemoverMain(pathToFile)
			return err
		}

		// run the program
		fmt.Printf("./%s\n", executableName)
		timeStart := time.Now()
		if err := runCommand("./" + executableName); err != nil {
			headerRemoverMain(pathToFile)
		}
		durationRun = time.Since(timeStart)
	}

	// Add header
	fmt.Printf("Add header to %s\n", pathToFile)
	if err := headerInserterMain(pathToFile, false, "1", timeoutReplay, false); err != nil {
		return fmt.Errorf("Error in adding header: %v", err)
	}

	// build the program
	fmt.Printf("%s build\n", pathToPatchedGoRuntime)
	if err := runCommand(pathToPatchedGoRuntime, "build"); err != nil {
		log.Println("Error in building program, removing header and stopping workflow")
		headerRemoverMain(pathToFile)
		return err
	}

	// run the program
	fmt.Printf("./%s\n", executableName)
	timeStart := time.Now()
	if err := runCommand("./" + executableName); err != nil {
		headerRemoverMain(pathToFile)
	}
	durationRecord = time.Since(timeStart)

	// Remove header
	if err := headerRemoverMain(pathToFile); err != nil {
		return fmt.Errorf("Error removing header: %v", err)
	}

	// Apply analyzer
	analyzerOutput := filepath.Join(dir, "advocateTrace")
	timeStart = time.Now()
	runAnalyzer(analyzerOutput, noPrintFlag, noRewriteFlag, analyisCasesFlag,
		"results_readable.log", "results_machine.log",
		ignoreAtomicsFlag, fifoFlag, ignoreCriticalSectionFlag, noWarningFlag, rewriteAllFlag,
		"rewritten_trace", timeoutAna, ignoreRewriteFlag, fuzzing)
	if err := runCommand(pathToAnalyzer, "run", "-trace", analyzerOutput, "-timeout", strconv.Itoa(timeoutAna)); err != nil {
		return fmt.Errorf("Error applying analyzer: %v", err)
	}
	durationAnalysis = time.Since(timeStart)

	// Find rewritten_trace directories
	rewrittenTraces, err := filepath.Glob(filepath.Join(dir, "rewritten_trace*"))
	if err != nil {
		return fmt.Errorf("Error finding rewritten traces: %v", err)
	}

	// Apply replay header and run tests for each trace
	timeoutRepl := time.Duration(0)
	if timeoutReplay == -1 {
		timeoutRepl = 100 * durationRecord
	} else {
		timeoutRepl = time.Duration(timeoutReplay) * time.Second
	}

	timeStart = time.Now()
	for _, trace := range rewrittenTraces {
		rtraceNum := extractTraceNum(trace)
		fmt.Printf("Apply replay header for file f %s and trace %s\n", pathToFile, rtraceNum)
		if err := headerInserterMain(pathToFile, true, rtraceNum, int(timeoutRepl.Seconds()), false); err != nil {
			return err
		}

		// build the program
		fmt.Printf("%s build\n", pathToPatchedGoRuntime)
		if err := runCommand(pathToPatchedGoRuntime, "build"); err != nil {
			log.Println("Error in building program, removing header and stopping workflow")
			headerRemoverMain(pathToFile)
			continue
		}

		// run the program
		fmt.Printf("./%s\n", executableName)
		runCommand("./" + executableName)

		fmt.Printf("Remove replay header from %s\n", pathToFile)
		if err := headerRemoverMain(pathToFile); err != nil {
			return err
		}
	}

	durationReplay = time.Since(timeStart)

	resultPath := filepath.Join(dir, "advocateResult")

	if !keepTraces {
		removeTraces(dir)
	}

	moveResults(dir, resultPath)

	// Generate Bug Reports
	fmt.Println("Generate Bug Reports")
	generateBugReports(resultPath)

	resTimes := map[string]time.Duration{
		"run":      durationRun,
		"record":   durationRecord,
		"analyzer": durationAnalysis,
		"replay":   durationReplay,
	}

	if measureTime {
		updateTimeFiles(programName, "Main", resultPath, resTimes, len(rewrittenTraces), 1)
	}

	if notExecuted {
		fmt.Println("Check for untriggered selects and not executed progs")
		complete.Check(filepath.Join(dir, "advocateResult"), dir)
	}

	if createStats {
		// create statistics
		fmt.Println("Create statistics")
		stats.CreateStats(dir, programName, "")
	}

	return nil
}

func extractTraceNum(tracePath string) string {
	re := regexp.MustCompile(`[0-9]+$`)
	return re.FindString(tracePath)
}
