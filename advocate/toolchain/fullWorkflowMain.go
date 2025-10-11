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
	"advocate/results/complete"
	"advocate/results/stats"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/timer"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Run ADVOCATE on a program with a main function
//
// Parameter:
//   - runRecord bool: run the recording. If set to false, but runAnalysis or runReplay is
//     set the trace at tracePath is used
//   - runAnalysis bool: run the analysis on a path
//   - runReplay bool: run replay, if runAnalysis is true, those replays are used
//     otherwise the trace at tracePath is replayed
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//   - onlyRecord bool: if true, only record th trace, but do not run any analysis
//
// Returns:
//   - int: TraceID
//   - int: number results
//   - error
func runWorkflowMain(
	runRecord, runAnalysis, runReplay bool,
	fuzzing int, fuzzingTrace string,
	firstRun bool) (int, int, error) {
	if _, err := os.Stat(paths.Prog); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("file %s does not exist", paths.Prog)
	}

	log.Info("Run main")

	// Change to the directory of the main file
	if err := os.Chdir(paths.ProgDir); err != nil {
		return 0, 0, fmt.Errorf("Failed to change directory: %v", err)
	}

	if firstRun {
		os.RemoveAll("advocateResult")
		if err := os.MkdirAll("advocateResult", os.ModePerm); err != nil {
			return 0, 0, fmt.Errorf("Failed to create advocateResult directory: %v", err)
		}

		// Remove possibly leftover traces from unexpected aborts that could interfere with replay
		RemoveTraces(paths.ProgDir)
		removeLogs(paths.ProgDir)
	}

	outFile, err := os.OpenFile(paths.NameOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to open log file: %v", err)
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
	if err := os.Setenv("GOROOT", paths.GoPatch); err != nil {
		return 0, 0, fmt.Errorf("Failed to set GOROOT: %v", err)
	}
	// Unset GOROOT
	defer os.Unsetenv("GOROOT")
	if runRecord {
		// Remove header
		if err := headerRemoverMain(paths.Prog); err != nil {
			return 0, 0, fmt.Errorf("Error removing header: %v", err)
		}

		// build the program
		if flags.MeasureTime && fuzzing < 1 {
			log.Info("Build Program")
			fmt.Printf("%s build\n", paths.Go)
			if err := runCommand(origStdout, origStderr, paths.Go, "build"); err != nil {
				log.Error("Error in building program, removing header and stopping workflow")
				headerRemoverMain(paths.Prog)
				return 0, 0, err
			}

			// run the program
			log.Info("Execute Program")
			timer.Start(timer.Run)
			execPath := helper.MakePathLocal(flags.ExecName)
			if err := runCommand(origStdout, origStderr, execPath); err != nil {
				headerRemoverMain(paths.Prog)
			}
			timer.Stop(timer.Run)
		}

		// Add header
		if err := headerInserterMain(paths.Prog, false, "1", flags.TimeoutReplay, false, fuzzing, fuzzingTrace); err != nil {
			return 0, 0, fmt.Errorf("Error in adding header: %v", err)
		}

		// build the program
		log.Info("Build program for recording")
		if err := runCommand(origStdout, origStderr, paths.Go, "build", "-gcflags=all=-N -l"); err != nil {
			log.Error("Error in building program, removing header and stopping workflow")
			headerRemoverMain(paths.Prog)
			return 0, 0, err
		}

		// run the recording
		log.Info("Run program for recording")
		timer.Start(timer.Recording)
		execPath := helper.MakePathLocal(flags.ExecName)
		if err := runCommand(origStdout, origStderr, execPath); err != nil {
			headerRemoverMain(paths.Prog)
		}
		timer.Stop(timer.Recording)

		// Remove header
		if err := headerRemoverMain(paths.Prog); err != nil {
			return 0, 0, fmt.Errorf("Error removing header: %v", err)
		}
	}

	// Apply analyzer
	if runAnalysis {
		analyzerOutput := filepath.Join(paths.ProgDir, "advocateTrace")

		err = runAnalyzer(analyzerOutput, paths.NameResultReadable, paths.NameResultMachine,
			"rewrittenTrace", fuzzing)

		if err != nil {
			return 0, 0, err
		}
	}

	rewrittenTraces := make([]string, 0)
	if runReplay {
		log.Info("Run replay")
		// Find rewrittenTrace directories
		if runAnalysis {
			rewrittenTraces, err = filepath.Glob(filepath.Join(paths.ProgDir, "rewrittenTrace*"))
			if err != nil {
				return 0, 0, fmt.Errorf("Error finding rewritten traces: %v", err)
			}
		} else {
			if flags.TracePath != "" {
				rewrittenTraces = append(rewrittenTraces, flags.TracePath)
			}
		}

		timer.Start(timer.Replay)
		for _, trace := range rewrittenTraces {
			traceNum := extractTraceNum(trace)
			fmt.Printf("Apply replay header for file f %s and trace %s\n", paths.Prog, traceNum)
			if err := headerInserterMain(paths.Prog, true, traceNum, flags.TimeoutReplay, false, fuzzing, fuzzingTrace); err != nil {
				return 0, 0, err
			}

			// build the program
			log.Info("Build program for replay")
			if err := runCommand(origStdout, origStderr, paths.Go, "build", "-gcflags=all=-N -l"); err != nil {
				log.Error("Error in building program, removing header and stopping workflow")
				headerRemoverMain(paths.Prog)
				continue
			}

			// run the program
			log.Info("Run program for replay")
			execPath := helper.MakePathLocal(flags.ExecName)
			runCommand(origStdout, origStderr, execPath)

			fmt.Printf("Remove replay header from %s\n", paths.Prog)
			if err := headerRemoverMain(paths.Prog); err != nil {
				return 0, 0, err
			}
		}
		timer.Stop(timer.Replay)
	}

	if !flags.KeepTraces && !flags.CreateStatistics {
		RemoveTraces(paths.Prog)
	}

	total := fuzzing != -1
	collect(paths.ProgDir, paths.ProgDir, paths.Result, total)

	// Generate Bug Reports
	var numberResults int
	if runAnalysis {
		log.Info("Generate Bug Reports")
		numberResults = generateBugReports(movedTraces, fuzzing)

		timer.UpdateTimeFileDetail("Main", len(rewrittenTraces))
	}

	if flags.NotExecuted {
		complete.Check(paths.Result, paths.Prog)
	}

	if flags.CreateStatistics {
		stats.CreateStats("", movedTraces, fuzzing)
	}

	if total {
		removeLogs(paths.Prog)
	}

	return movedTraces, numberResults, nil
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
