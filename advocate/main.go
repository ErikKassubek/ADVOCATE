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
	"path/filepath"
	"strings"

	"advocate/analysis/data"
	"advocate/fuzzing"
	fuzzingdata "advocate/fuzzing/data"
	"advocate/fuzzing/gopie"
	"advocate/results/stats"
	"advocate/toolchain"
	"advocate/utils/control"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
)

var (
	help bool

	pathToAdvocate string

	tracePath string
	progPath  string

	progName string
	execName string

	timeoutRecording int
	timeoutReplay    int
	timeoutFuzzing   int
	recordTime       bool

	maxFuzzingRun int

	noFifo                bool
	ignoreCriticalSection bool
	ignoreAtomics         bool

	rewriteAll bool

	noRewrite    bool
	keepTraces   bool
	skipExisting bool

	notExec    bool
	statistics bool

	scenarios         string
	onlyAPanicAndLeak bool

	fuzzingMode string

	modeMain bool

	noWarning bool

	noInfo     bool
	noProgress bool

	cont bool

	alwaysPanic bool

	settings string
	output   bool

	cancelTestIfBugFound bool

	noMemorySupervisor bool

	maxNumberElements int

	sameElemTypeInSC     bool
	scSize               int
	fuzzingWithoutReplay bool
)

// Main function
func main() {
	flag.BoolVar(&help, "h", false, "Print help")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.StringVar(&progPath, "path", "", "Path to the program folder, for main: path to main file, for test: path to test folder")

	flag.StringVar(&progName, "prog", "", "Name of the program")
	flag.StringVar(&execName, "exec", "", "Name of the executable or test")

	flag.StringVar(&tracePath, "trace", "", "Path to the trace folder to replay")

	flag.IntVar(&timeoutRecording, "timeoutRec", 600, "Set the timeout in seconds for the recording. Default: 600s. To disable set to -1")
	flag.IntVar(&timeoutReplay, "timeoutRep", 900, "Set a timeout in seconds for the replay. Default: 600s. To disable set to -1")

	flag.IntVar(&timeoutFuzzing, "timeoutFuz", 420, "Timeout of fuzzing per test/program in seconds. Default: 7min. To Disable, set to -1")
	flag.IntVar(&maxFuzzingRun, "maxFuzzingRuns", -1, "Maximum number of fuzzing runs per test/prog. Default: -1. To Disable, set to -1")

	flag.BoolVar(&recordTime, "time", false, "measure the runtime")

	flag.BoolVar(&noMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.BoolVar(&noFifo, "noFifo", false, "Do not assume a FIFO ordering for buffered channels")
	flag.BoolVar(&ignoreCriticalSection, "ignCritSec", false, "Ignore happens before relations of critical sections (default false)")
	flag.BoolVar(&ignoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")

	flag.BoolVar(&rewriteAll, "replayAll", false, "Replay a bug even if it has already been confirmed")
	rewriteAll = true

	flag.BoolVar(&noRewrite, "noRewrite", false, "Do not rewrite the trace file (default false)")
	flag.BoolVar(&keepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")
	flag.BoolVar(&skipExisting, "skipExisting", false, "If set, all tests that already have a results folder will be skipped. Also skips failed tests.")

	flag.BoolVar(&notExec, "notExec", false, "Find never executed operations")
	flag.BoolVar(&statistics, "stats", false, "Create statistics.")

	flag.BoolVar(&noWarning, "noWarning", false, "Only show critical bugs")

	flag.BoolVar(&cont, "cont", false, "Continue a partial analysis of tests")

	flag.BoolVar(&noInfo, "noInfo", false, "Do not show infos in the terminal (will only show results, errors, important and progress)")
	flag.BoolVar(&noProgress, "noProgress", false, "Do not show progress info")

	flag.BoolVar(&alwaysPanic, "panic", false, "Panic if the analysis panics")

	flag.BoolVar(&output, "output", false, "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

	flag.IntVar(&maxNumberElements, "maxNumberElements", 10000000, "Set the maximum number of elements in a trace. Traces with more elements will be skipped. To disable set -1. Default: 10000000")

	flag.StringVar(&scenarios, "scen", "", "Select which analysis scenario to run, e.g. -scen srd for the option s, r and d."+
		"If not set, all scenarios are run.\n"+
		"Options:\n"+
		"\ts: Send on closed channel\n"+
		"\tr: Receive on closed channel\n"+
		"\tw: Done before add on waitGroup\n"+
		"\tn: Close of closed channel\n"+
		"\tb: Concurrent receive on channel\n"+
		"\tl: Leaking routine\n"+
		"\tu: Unlock of unlocked mutex\n"+
		"\tc: Cyclic deadlock\n",
	)
	// "\tm: Mixed deadlock\n"

	flag.BoolVar(&onlyAPanicAndLeak, "onlyActual", false, "only test for actual bugs leading to panic and actual leaks. This will overwrite `scen`")

	flag.StringVar(&fuzzingMode, "fuzzingMode", "",
		"Mode for fuzzing. Possible values are:\n\tGFuzz\n\tGFuzzHB\n\tGFuzzHBFlow\n\tFlow\n\tGoPie\n\tGoPie+\n\tGoPieHB")

	// partially implemented by may not work, therefore disables, enable again when fixed
	flag.BoolVar(&modeMain, "main", false, "set to run on main function")

	flag.StringVar(&settings, "settings", "", "Set some internal settings. For more info, see ../doc/usage.md")

	flag.BoolVar(&cancelTestIfBugFound, "cancelTestIfBugFound", false, "Skip further fuzzing runs of a test if one bug has been found")

	// for experiments
	flag.BoolVar(&sameElemTypeInSC, "sameElemTypeInSC", false, "Only allow elements of the same type in the same SC")
	flag.IntVar(&scSize, "scSize", 4, "max number of elements in SC")
	flag.BoolVar(&fuzzingWithoutReplay, "fuzzingWithoutReplay", false, "Disable replay before SC in fuzzing")

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
	if modeMain {
		var err error
		progPath, err = helper.GetMainPath(progPath)
		if err != nil {
			log.Error("Could not find main file. If -main is set, -path should point to the main file.")
			log.Error(err)
			return
		}
	}

	log.Init(noInfo, noProgress, !alwaysPanic)

	helper.SetSettings(settings, maxFuzzingRun, fuzzingMode)

	progPathDir := helper.GetDirectory(progPath)
	timer.Init(recordTime, progPathDir)
	timer.Start(timer.Total)
	defer timer.Stop(timer.Total)

	execPath, _ := os.Executable()
	pathToAdvocate = filepath.Dir(filepath.Dir(execPath))

	control.SetMaxNumberElem(maxNumberElements)
	if !noMemorySupervisor {
		go control.Supervisor() // cancel analysis if not enough ram
	}

	// don't run any HB Analysis for direct GFuzz, GoPie and GoPie+
	if mode == "fuzzing" && (fuzzingMode == fuzzingdata.GFuzz ||
		fuzzingMode == fuzzingdata.GoPie || fuzzingMode == fuzzingdata.GoPiePlus) {
		scenarios = "-"
		onlyAPanicAndLeak = true
	}

	var err error
	data.AnalysisCasesMap, err = parseAnalysisCases(scenarios)
	if err != nil {
		log.Error("Could not read analysis cases: ", err)
		return
	}

	toolchain.SetFlags(noRewrite, ignoreAtomics,
		!noFifo, ignoreCriticalSection, rewriteAll, onlyAPanicAndLeak,
		timeoutRecording, timeoutReplay, rewriteAll, noWarning, tracePath, output)

	modeMainTest := "test"
	if modeMain {
		modeMainTest = "main"
	}

	execName = helper.CheckGoMod(progPath, modeMain, execName)

	if modeMain && execName == "" {
		log.Error("Could not determine executable name from go.mod. Provide with -exec [ExecutableName]")
		panic(fmt.Errorf("Could not determine executable name"))
	}

	switch mode {
	case "analysis":
		modeToolchain(modeMainTest, true, true, true)
	case "fuzzing":
		modeFuzzing()
	case "record", "recording":
		keepTraces = true
		modeToolchain(modeMainTest, true, false, false)
	case "replay":
		modeToolchain(modeMainTest, false, false, true)
	default:
		log.Errorf("Unknown mode %s\n", os.Args[1])
		log.Error("Select one mode from  'analysis', 'fuzzing' or 'record'")
		helper.PrintHelp()
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
			log.Resultf(false, false, "", "Tests with indicated bugs: %d", numberTestWithRes)
			log.Resultf(false, false, "", "Number indicated bugs:  %d", numberBugs)
		}
	}
	timer.UpdateTimeFileOverview(progName, "*Total*")
	log.Important("Total time: ", timer.GetTime(timer.Total))
}

// modeFuzzing starts the fuzzing
func modeFuzzing() {
	if progName == "" {
		progName = helper.GetProgName(progPath)
	}

	progPath, err := helper.CheckPath(progPath)
	if err != nil {
		log.Error("Error on checking prog path: ", err)
		log.Error("Set path with -path [path]")
		panic(err)
	}

	helper.GoPieSCStart = scSize
	gopie.SameElementTypeInSC = sameElemTypeInSC
	gopie.WithoutReplay = fuzzingWithoutReplay

	err = fuzzing.Fuzzing(modeMain, fuzzingMode, pathToAdvocate, progPath,
		progName, execName, ignoreAtomics, recordTime, notExec, statistics,
		keepTraces, cont, timeoutFuzzing, maxFuzzingRun, cancelTestIfBugFound)
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
func modeToolchain(mode string, record bool, analysis bool, replay bool) {
	progPath, err := helper.CheckPath(progPath)
	if err != nil {
		log.Error("Error on checking prog path: ", err)
		panic(err)
	}

	if !record && (analysis || replay) {
		tracePath, err = helper.CheckPath(tracePath)
		if err != nil {
			log.Error("Error on checking trace path: ", err)
			panic(err)
		}
	}

	if mode == "test" && !record && replay && execName == "" {
		log.Error("When running replay of test without recording, -exec [TestName] must be set")
		panic("When running replay of test without recording, -exec [TestName] must be set")
	}

	_, _, err = toolchain.Run(mode, pathToAdvocate, progPath, "", record, analysis,
		replay, execName, progName, execName, -1, "", ignoreAtomics, recordTime,
		notExec, statistics, keepTraces, skipExisting, true, cont, 0, 0)
	if err != nil {
		log.Error("Failed to run toolchain: ", err.Error())
	}

	if statistics {
		// TODO: check if this
		err = stats.CreateStatsTotal(progPath, progName)
		if err != nil {
			log.Error("Failed to create stats total: ", err.Error())
		}
	}
}

// Parse the given analysis cases
//
// Parameter:
//   - cases string: The string of analysis cases to parse
//
// Returns:
//   - map[data.AnalysisCases]bool: A map of the analysis cases and if they are set
//   - error: An error if the cases could not be parsed
func parseAnalysisCases(cases string) (map[data.AnalysisCases]bool, error) {
	analysisCases := map[data.AnalysisCases]bool{
		data.All:              false, // all cases enabled
		data.SendOnClosed:     false,
		data.ReceiveOnClosed:  false,
		data.DoneBeforeAdd:    false,
		data.CloseOnClosed:    false,
		data.ConcurrentRecv:   false,
		data.Leak:             false,
		data.UnlockBeforeLock: false,
		data.MixedDeadlock:    false,
		data.ResourceDeadlock: false,
	}

	if cases == "-" {
		return analysisCases, nil
	}

	if cases == "" {
		for c := range analysisCases {
			analysisCases[c] = true
		}

		// remove when implemented
		analysisCases[data.MixedDeadlock] = false

		return analysisCases, nil
	}

	for _, c := range cases {
		switch c {
		case 's':
			analysisCases[data.SendOnClosed] = true
		case 'r':
			analysisCases[data.ReceiveOnClosed] = true
		case 'w':
			analysisCases[data.DoneBeforeAdd] = true
		case 'n':
			analysisCases[data.CloseOnClosed] = true
		case 'b':
			analysisCases[data.ConcurrentRecv] = true
		case 'l':
			analysisCases[data.Leak] = true
		case 'u':
			analysisCases[data.UnlockBeforeLock] = true
		case 'c':
			analysisCases[data.ResourceDeadlock] = true
		// case 'm':
		// analysisCases[data.MixedDeadlock] = true
		default:
			return nil, fmt.Errorf("Invalid analysis case: %c", c)
		}
	}

	all := true
	for key, val := range analysisCases {
		if key == data.All {
			continue
		}
		if !val {
			all = false
			break
		}
	}

	if all {
		analysisCases[data.All] = true
	}

	return analysisCases, nil
}

// getFolderTrace returns the path to the folder containing the trace, given the
// path to the trace
//
// Parameter:
//   - pathTrace string: path to the traces
//
// Returns:
//   - string: path to the folder containing the trace folder
//   - error
func getFolderTrace(pathTrace string) (string, error) {
	folderTrace, err := filepath.Abs(pathTrace)
	if err != nil {
		return "", err
	}

	// remove last folder from path
	return folderTrace[:strings.LastIndex(folderTrace, string(os.PathSeparator))+1], nil
}
