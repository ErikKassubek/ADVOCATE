// Copyright (c) 2024 Erik Kassubek
//
// File: main.go
// Brief: Main file and starting point for the toolchain
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gocdr/analysis/baseA"
	"gocdr/fuzzing"
	"gocdr/fuzzing/baseF"
	"gocdr/results/stats"
	"gocdr/toolchain"
	"gocdr/utils/control"
	"gocdr/utils/flags"
	"gocdr/utils/helper"
	"gocdr/utils/log"
	"gocdr/utils/paths"
	"gocdr/utils/settings"
	"gocdr/utils/timer"
)

var (
	help bool
)

// Main function
func main() {
	flag.BoolVar(&help, "h", false, "Print help")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.StringVar(&flags.ProgPath, "path", "", "Path to the program folder, for main: path to main file, for test: path to test folder")

	flag.StringVar(&flags.ProgName, "prog", "", "Name of the program")
	flag.StringVar(&flags.ExecName, "exec", "", "Name of the executable or test")

	flag.StringVar(&flags.TracePath, "trace", "", "Path to the trace folder to replay")

	flag.IntVar(&flags.TimeoutRecording, "timeoutRec", 600, "Set the timeout in seconds for the recording. Default: 600s. To disable set to -1")
	flag.IntVar(&flags.TimeoutReplay, "timeoutRep", 900, "Set a timeout in seconds for the replay. Default: 600s. To disable set to -1")
	flag.IntVar(&flags.TimeoutFuzzing, "timeoutFuz", 420, "Timeout of fuzzing per test/program in seconds. Default: 7min. To Disable, set to -1")
	flag.IntVar(&flags.MaxFuzzingRun, "maxFuzzingRuns", -1, "Maximum number of fuzzing runs per test/prog. Default: -1. To Disable, set to -1")
	flag.IntVar(&flags.MaxNumberElements, "maxNumberElements", 10000000, "Set the maximum number of elements in a trace. Traces with more elements will be skipped. To disable set -1. Default: 10000000")

	flag.BoolVar(&flags.MeasureTime, "time", false, "measure the runtime")

	flag.BoolVar(&flags.IgnoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")

	flag.BoolVar(&flags.NoSkipRewrite, "replayAll", false, "Replay a bug even if it has already been confirmed")
	flag.BoolVar(&flags.NoRewrite, "noRewrite", false, "Do not rewrite the trace file (default false)")
	// flag.BoolVar(&flags.KeepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")
	flag.BoolVar(&flags.SkipExisting, "skipExisting", false, "If set, all tests that already have a results folder will be skipped. Also skips failed tests.")

	flag.BoolVar(&flags.Continue, "cont", false, "Continue a partial analysis of tests")

	flag.BoolVar(&flags.NoWarning, "noWarning", false, "Only show critical bugs")
	flag.BoolVar(&flags.NoInfo, "noInfo", false, "Do not show infos in the terminal (will only show results, errors, important and progress)")
	flag.BoolVar(&flags.NoProgress, "noProgress", false, "Do not show progress info")
	flag.BoolVar(&flags.Output, "output", false, "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

	flag.BoolVar(&flags.AlwaysPanic, "panic", false, "Panic if the analysis panics")
	flag.BoolVar(&flags.NoMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.BoolVar(&flags.KeepResultDir, "keepResultDir", false, "Do not replace old result dirs")

	flag.StringVar(&flags.FuzzingMode, "mode", "",
		"Mode for fuzzing. Possible values are:n\tGFuzz\n\tGoPie")

	flag.BoolVar(&flags.ModeMain, "main", false, "set to run on main function")

	flag.BoolVar(&flags.CancelTestIfBugFound, "cancelTestIfBugFound", false, "Skip further fuzzing runs of a test if one bug has been found")

	flag.Parse()

	var mode string
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		mode = os.Args[1]
		flag.CommandLine.Parse(os.Args[2:])
		if help {
			helper.PrintHelpMode(mode)
			return
		}
	} else {
		if help {
			helper.PrintHelp()
			return
		}
		helper.PrintHelp()
		return
	}

	// If -main is set, the path needs to be the path to the main file
	// If the given path is to a folder, check if a main.go file exists in this folder
	// If so, fix the path. Otherwise return error and finish
	if flags.ModeMain {
		var err error
		flags.ProgPath, err = paths.GetMainPath(flags.ProgPath)
		if err != nil {
			log.Error("Could not find main file. If -main is set, -path should point to the main file.")
			log.Error(err)

			return
		}
	}

	settings.SetSettings()
	paths.BuildPaths(flags.ModeMain)

	progPathDir := paths.GetDirectory(flags.ProgPath)
	timer.Init(progPathDir)
	timer.Start(timer.Total)
	defer timer.Stop(timer.Total)

	control.SetMaxNumberElem()
	if !flags.NoMemorySupervisor {
		go control.Supervisor(baseA.ClearTrace, baseA.ClearData, fuzzing.ResetFuzzing) // cancel analysis if not enough ram
	}

	// don't run any HB Analysis for direct GFuzz, GoPie and GoCR
	if mode == "fuzzing" && (flags.FuzzingMode == baseF.GFuzz ||
		flags.FuzzingMode == baseF.GoPie || flags.FuzzingMode == baseF.GoCR) {
		flags.Scenarios = "-"
	}

	if flags.FuzzingMode == "" {
		flags.FuzzingMode = baseF.Guided
	}

	var err error
	baseA.AnalysisCasesMap, err = flags.ParseAnalysisCases()
	if err != nil {
		log.Error("Could not read analysis cases: ", err)
		return
	}

	modeMainTest := "test"
	if flags.ModeMain {
		modeMainTest = "main"
	}

	helper.CheckGoMod()
	helper.RunGoModTidy()

	if flags.ModeMain && flags.ExecName == "" {
		log.Error("Could not determine executable name from go.mod. Provide with -exec [ExecutableName]")
		panic(fmt.Errorf("Could not determine executable name"))
	}

	modeTool := false
	run := false
	record := false
	analysis := false
	replay := false

	flags.KeepTraces = true

	switch mode {
	case "fuzzing":
		modeFuzzing()
	case "record", "recording":
		modeTool = true
		record = true
	case "analysis":
		modeTool = true
		analysis = true
	case "replay":
		modeTool = true
		replay = true
	case "run":
		modeTool = true
		run = true
	default:
		log.Errorf("Unknown mode %s\n", os.Args[1])
		log.Error("Select one mode from  'analysis', 'fuzzing' or 'record'")
		helper.PrintHelp()
	}

	if modeTool {
		modeToolchain(modeMainTest, run, record, analysis, replay)
	}

	numberBugs, _, numberTestWithRes, numberErr, numberTimeout := log.GetLoggingNumbers()
	if numberErr == 0 {
		log.Info("Finished with 0 errors")
	} else {
		log.Errorf("Finished with %d errors", numberErr)
	}
	if numberTimeout == 0 {
		log.Info("No internal replay timeouts occurred")
	} else {
		log.Errorf("%d internal replay timeouts occurred", numberTimeout)
	}
	if mode == "analysis" || mode == "fuzzing" {
		if numberTestWithRes == 0 {
			log.Info("No bugs have been found/indicated")
		} else {
			if !flags.ModeMain {
				log.Resultf(false, false, "", "Tests with indicated bugs: %d", numberTestWithRes)
			}
			log.Resultf(false, false, "", "Number indicated bugs:  %d", numberBugs)
		}
	}
	timer.UpdateTimeFileOverview("*Total*")
}

// modeFuzzing starts the fuzzing
func modeFuzzing() {
	if flags.ProgName == "" {
		flags.ProgName = paths.GetProgName(flags.ProgPath)
	}

	var err error
	flags.ProgPath, err = paths.CheckPath(flags.ProgPath)
	if err != nil {
		log.Error("Error on checking prog path: ", err)
		log.Error("Set path with -path [path]")
		panic(err)
	}

	err = fuzzing.Fuzzing()
	if err != nil {
		log.Error("Fuzzing Failed: ", err.Error())
	}
}

// Start point for the toolchain
// This will run, analyze and replay a given program or test
//
// Parameter:
//   - mode string: main for main function, test for test function
//   - record bool: if true, the toolchain will run the recording
//   - analysis bool: if true, the toolchain will run analysis
//   - replay bool: if true, the toolchain will run replays
//
// Note:
//   - If recording is false, but analysis or replay is set, -trace must be set
func modeToolchain(mode string, run bool, record bool, analysis bool, replay bool) {
	var err error
	flags.ProgPath, err = paths.CheckPath(flags.ProgPath)
	if err != nil {
		log.Error("Error on checking prog path: ", err)
		panic(err)
	}

	if !record && (analysis || replay) {
		flags.TracePath, err = paths.CheckPath(flags.TracePath)
		if err != nil {
			log.Errorf("Error on checking trace path: %s. Set with -trace [path]", err)
			panic(err)
		}
	}

	if mode == "test" && !record && replay && flags.ExecName == "" {
		log.Error("When running replay of test without recording, -exec [TestName] must be set")
		panic("When running replay of test without recording, -exec [TestName] must be set")
	}

	firstRun := true
	fileNumber, testNumber := 1, 0
	_, _, err = toolchain.Run(mode, "", run, record, analysis,
		replay, -1, "", firstRun, fileNumber, testNumber)
	if err != nil {
		log.Error("Failed to run toolchain: ", err.Error())
	}

	if flags.CreateStatistics {
		err = stats.CreateStatsTotal(flags.ProgPath)
		if err != nil {
			log.Error("Failed to create stats total: ", err.Error())
		}
	}
}
