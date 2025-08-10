//
// File: main.go
// Brief: Main file and starting point for the toolchain
//
// Created: 2023-08-08
//
// License: BSD-3-Clause

package main

import (
	"flag"
	"fmt"
	"goCR/fuzzing"
	"goCR/fuzzing/gopie"
	"goCR/toolchain"
	"goCR/utils/control"
	"goCR/utils/helper"
	"goCR/utils/log"
	"goCR/utils/timer"
	"os"
	"path/filepath"

	fuzzingdata "goCR/fuzzing/data"
)

var (
	help bool

	pathToGoCR string

	tracePath string
	progPath  string

	progName string
	execName string

	timeoutExec    int
	timeoutFuzzing int
	recordTime     bool

	maxFuzzingRun int

	ignoreAtomics bool

	keepTraces bool

	statistics bool

	scenarios string

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

	flag.IntVar(&timeoutExec, "timeoutExec", 180, "Set a timeout in seconds for the execution of one run of a test. Default: 3min. To disable set to -1")
	flag.IntVar(&timeoutFuzzing, "timeoutProg", 420, "Timeout of fuzzing per test/program in seconds. Default: 7min. To Disable, set to -1")

	flag.IntVar(&maxFuzzingRun, "maxFuzzingRuns", -1, "Maximum number of fuzzing runs per test/prog. Default: -1. To Disable, set to -1")

	flag.BoolVar(&recordTime, "time", false, "measure the runtime")

	flag.BoolVar(&noMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.BoolVar(&ignoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")

	flag.BoolVar(&keepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")

	flag.BoolVar(&statistics, "stats", false, "Create statistics.")

	flag.BoolVar(&noWarning, "noWarning", false, "Only show critical bugs")

	flag.BoolVar(&cont, "cont", false, "Continue a partial analysis of tests")

	flag.BoolVar(&noInfo, "noInfo", false, "Do not show infos in the terminal (will only show results, errors, important and progress)")
	flag.BoolVar(&noProgress, "noProgress", false, "Do not show progress info")

	flag.BoolVar(&alwaysPanic, "panic", false, "Panic if the analysis panics")

	flag.BoolVar(&output, "output", false, "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

	flag.IntVar(&maxNumberElements, "maxNumberElements", 10000000, "Set the maximum number of elements in a trace. Traces with more elements will be skipped. To disable set -1. Default: 10000000")

	flag.StringVar(&fuzzingMode, "mode", "",
		"Mode for fuzzing. Possible values are:\n\tGoCR\n\tGFuzz\n\tGoPie")

	// partially implemented by may not work, therefore disables, enable again when fixed
	flag.BoolVar(&modeMain, "main", false, "set to run on main function")

	flag.StringVar(&settings, "settings", "", "Set some internal settings. For more info, see ../doc/usage.md")

	flag.BoolVar(&cancelTestIfBugFound, "cancelTestIfBugFound", false, "Skip further fuzzing runs of a test if one bug has been found")

	// for experiments
	flag.BoolVar(&sameElemTypeInSC, "sameElemTypeInSC", false, "Only allow elements of the same type in the same SC")
	flag.IntVar(&scSize, "scSize", -1, "max number of elements in SC")
	flag.BoolVar(&fuzzingWithoutReplay, "fuzzingWithoutReplay", false, "Disable replay before SC in fuzzing")
	flag.BoolVar(&fuzzingdata.FinishIfBugFound, "finishIfBugFound", false, "Finish fuzzing as soon as a bug was found")

	flag.Parse()

	if help {
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
	pathToGoCR = filepath.Dir(filepath.Dir(execPath))

	control.SetMaxNumberElem(maxNumberElements)
	if !noMemorySupervisor {
		go control.Supervisor() // cancel analysis if not enough ram
	}

	toolchain.SetFlags(ignoreAtomics,
		timeoutExec, timeoutExec, noWarning, tracePath, output)

	execName = helper.CheckGoMod(progPath, modeMain, execName)

	if modeMain && execName == "" {
		log.Error("Could not determine executable name from go.mod. Provide with -exec [ExecutableName]")
		panic(fmt.Errorf("Could not determine executable name"))
	}

	modeFuzzing()

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

	if numberTestWithRes == 0 {
		log.Info("No bugs have been found/indicated")
	} else {
		log.Resultf(false, false, "", "Tests with indicated bugs: %d", numberTestWithRes)
		log.Resultf(false, false, "", "Number indicated bugs:  %d", numberBugs)
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

	if scSize != -1 {
		helper.GoPieMaxSCLength = scSize
		helper.GoPieMaxSCLengthSet = true

	}
	gopie.SameElementTypeInSC = sameElemTypeInSC
	gopie.WithoutReplay = fuzzingWithoutReplay

	err = fuzzing.Fuzzing(modeMain, fuzzingMode, pathToGoCR, progPath,
		progName, execName, ignoreAtomics, recordTime, statistics,
		keepTraces, cont, timeoutFuzzing, maxFuzzingRun, cancelTestIfBugFound)
	if err != nil {
		log.Error("Fuzzing Failed: ", err.Error())
	}
}
