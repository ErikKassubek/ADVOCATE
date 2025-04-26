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
	"analyzer/timer"
	"analyzer/utils"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

// Run ADVOCATE on a program with a main function
//
// Parameter:
//   - pathToAdvocate string: path to the ADVOCATE folder
//   - pathToFile string: path to the file containing the main function
//   - runRecord bool: run the recording. If set to false, but runAnalysis or runReplay is
//     set the trace at tracePath is used
//   - runAnalysis bool: run the analysis on a path
//   - runReplay bool: run replay, if runAnalysis is true, those replays are used
//     otherwise the trace at tracePath is replayed
//   - executableName string: name of the executable
//   - keepTraces bool: do not delete the traces after analysis
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//   - onlyRecord bool: if true, only record th trace, but do not run any analysis
//
// Returns:
//   - error
func runWorkflowMain(pathToAdvocate string, pathToFile string,
	runRecord, runAnalysis, runReplay bool,
	executableName string, keepTraces bool, fuzzing int, fuzzingTrace string,
	firstRun bool) error {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", pathToFile)
	}

	utils.LogInfo("Run main")

	pathToPatchedGoRuntime := filepath.Join(pathToAdvocate, "go-patch/bin/go")

	if runtime.GOOS == "windows" {
		pathToPatchedGoRuntime += ".exe"
	}

	pathToGoRoot := filepath.Join(pathToAdvocate, "go-patch")

	// Change to the directory of the main file
	dir := filepath.Dir(pathToFile)
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("Failed to change directory: %v", err)
	}
	resultPath := filepath.Join(dir, "advocateResult")

	if firstRun {
		os.RemoveAll("advocateResult")
		if err := os.MkdirAll("advocateResult", os.ModePerm); err != nil {
			return fmt.Errorf("Failed to create advocateResult directory: %v", err)
		}

		// Remove possibly leftover traces from unexpected aborts that could interfere with replay
		removeTraces(dir)
		removeLogs(dir)
	}

	output := "output.log"
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
	// Unset GOROOT
	defer os.Unsetenv("GOROOT")

	if runRecord {
		utils.LogInfo("Run Recording")

		// Remove header
		if err := headerRemoverMain(pathToFile); err != nil {
			return fmt.Errorf("Error removing header: %v", err)
		}

		// build the program
		if measureTime && fuzzing < 1 {
			fmt.Printf("%s build\n", pathToPatchedGoRuntime)
			if err := runCommand(pathToPatchedGoRuntime, "build"); err != nil {
				utils.LogError("Error in building program, removing header and stopping workflow")
				headerRemoverMain(pathToFile)
				return err
			}

			// run the program
			timer.Start(timer.Run)
			execPath := utils.MakePathLocal(executableName)
			if err := runCommand(execPath); err != nil {
				headerRemoverMain(pathToFile)
			}
			timer.Stop(timer.Run)

		}

		// Add header
		if err := headerInserterMain(pathToFile, false, "1", timeoutReplay, false, fuzzing, fuzzingTrace); err != nil {
			return fmt.Errorf("Error in adding header: %v", err)
		}

		// build the program
		fmt.Printf("%s build\n", pathToPatchedGoRuntime)
		if err := runCommand(pathToPatchedGoRuntime, "build"); err != nil {
			utils.LogError("Error in building program, removing header and stopping workflow")
			headerRemoverMain(pathToFile)
			return err
		}

		// run the recording
		timer.Start(timer.Recording)
		execPath := utils.MakePathLocal(executableName)
		timer.Stop(timer.Recording)
		if err := runCommand(execPath); err != nil {
			headerRemoverMain(pathToFile)
		}

		// Remove header
		if err := headerRemoverMain(pathToFile); err != nil {
			return fmt.Errorf("Error removing header: %v", err)
		}
	}

	// Apply analyzer
	if runAnalysis {
		analyzerOutput := filepath.Join(dir, "advocateTrace")

		err = runAnalyzer(analyzerOutput, noRewriteFlag, analysisCasesFlag,
			"results_readable.log", "results_machine.log",
			ignoreAtomicsFlag, fifoFlag, ignoreCriticalSectionFlag, rewriteAllFlag,
			"rewrittenTrace", fuzzing, onlyAPanicAndLeakFlag)

		if err != nil {
			return err
		}
	}

	rewrittenTraces := make([]string, 0)
	if runReplay {
		utils.LogInfo("Run replay")
		// Find rewrittenTrace directories
		if runAnalysis {
			rewrittenTraces, err = filepath.Glob(filepath.Join(dir, "rewrittenTrace*"))
			if err != nil {
				return fmt.Errorf("Error finding rewritten traces: %v", err)
			}
		} else {
			if tracePathFlag != "" {
				rewrittenTraces = append(rewrittenTraces, tracePathFlag)
			}
		}

		// Apply replay header and run tests for each trace
		timeoutRepl := time.Duration(0)
		if timeoutReplay == -1 {
			timeoutRepl = 500 * timer.GetTime(timer.Recording)
			timeoutRepl = max(min(timeoutRepl, 10*time.Minute), time.Duration(timeoutRecording)*time.Second*2)
		} else {
			timeoutRepl = time.Duration(timeoutReplay) * time.Second
		}

		timer.Start(timer.Replay)
		for _, trace := range rewrittenTraces {
			traceNum := extractTraceNum(trace)
			fmt.Printf("Apply replay header for file f %s and trace %s\n", pathToFile, traceNum)
			if err := headerInserterMain(pathToFile, true, traceNum, int(timeoutRepl.Seconds()), false, fuzzing, fuzzingTrace); err != nil {
				return err
			}

			// build the program
			if err := runCommand(pathToPatchedGoRuntime, "build"); err != nil {
				utils.LogError("Error in building program, removing header and stopping workflow")
				headerRemoverMain(pathToFile)
				continue
			}

			// run the program
			execPath := utils.MakePathLocal(executableName)
			runCommand(execPath)

			fmt.Printf("Remove replay header from %s\n", pathToFile)
			if err := headerRemoverMain(pathToFile); err != nil {
				return err
			}
		}
		timer.Stop(timer.Replay)
	}

	if !keepTraces {
		removeTraces(dir)
	}

	total := fuzzing != -1
	collect(dir, dir, resultPath, total)

	// Generate Bug Reports
	if runAnalysis {
		fmt.Println("Generate Bug Reports")
		generateBugReports(resultPath, fuzzing)

		timer.UpdateTimeFileDetail(programName, "Main", len(rewrittenTraces))
	}

	if notExecuted {
		complete.Check(filepath.Join(dir, "advocateResult"), dir)
	}

	if createStats {
		// create statistics
		fmt.Println("Create statistics")
		stats.CreateStats(dir, programName, "", movedTraces, fuzzing)
	}

	if total {
		removeLogs(dir)
	}

	return nil
}

// Given a path to a trace file, return the trace number
//
// Parameter:
//   - tracePath: path to the file
//
// Returns:
//   - string: trace number
func extractTraceNum(tracePath string) string {
	re := regexp.MustCompile(`[0-9]+$`)
	return re.FindString(tracePath)
}
