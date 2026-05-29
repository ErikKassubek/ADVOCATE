// Copyright (c) 2024 Erik Kassubek
//
// File: cl.go
// Brief: command line entries
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package app

import (
	"advocate/fuzzing/baseF"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"flag"
	"os"
	"strings"
)

var (
	help bool
)

func CommandLine() string {
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
	flag.BoolVar(&flags.CreateStatistics, "stats", false, "Create statistics.")
	flag.BoolVar(&flags.NotExecuted, "notExec", false, "Find never executed operations")

	flag.BoolVar(&flags.IgnoreCriticalSection, "ignoreCritSec", false, "Ignore happens before relations of critical sections (default false)")
	flag.BoolVar(&flags.IgnoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")
	flag.BoolVar(&flags.OnlyAPanicAndLeak, "onlyActual", false, "only test for actual bugs leading to panic and actual leaks. This will overwrite `scen`")

	flag.BoolVar(&flags.NoSkipRewrite, "replayAll", false, "Replay a bug even if it has already been confirmed")
	flag.BoolVar(&flags.NoRewrite, "noRewrite", false, "Do not rewrite the trace file (default false)")
	flag.BoolVar(&flags.KeepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")
	flag.BoolVar(&flags.SkipExisting, "skipExisting", false, "If set, all tests that already have a results folder will be skipped. Also skips failed tests.")

	flag.BoolVar(&flags.Continue, "cont", false, "Continue a partial analysis of tests")

	flag.BoolVar(&flags.NoWarning, "noWarning", false, "Only show critical bugs")
	flag.BoolVar(&flags.NoInfo, "noInfo", false, "Do not show infos in the terminal (will only show results, errors, important and progress)")
	flag.BoolVar(&flags.NoProgress, "noProgress", false, "Do not show progress info")
	flag.BoolVar(&flags.Output, "output", false, "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

	flag.BoolVar(&flags.AlwaysPanic, "panic", false, "Panic if the analysis panics")
	flag.BoolVar(&flags.NoMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.StringVar(&flags.Scenarios, "scen", "", "Select which analysis scenario to run, e.g. -scen srd for the option s, r and d."+
		"If not set, all scenarios are run.\n"+
		"Options:\n"+
		"\ts: Send on closed channel\n"+
		// "\tr: Receive on closed channel\n"+
		"\tw: Done before add on waitGroup\n"+
		"\tn: Close of closed channel\n"+
		// "\tb: Concurrent receive on channel\n"+
		"\tl: Leaking routine\n"+
		"\tu: Unlock of unlocked mutex\n"+
		"\tc: Cyclic deadlock\n"+
		"\tm: Mixed deadlock\n",
	)

	flag.StringVar(&flags.FuzzingMode, "mode", "",
		"Mode for fuzzing. Possible values are:\n\tGuided\n\tGFuzz\n\tGFuzzHB\n\tGFuzzHBFlow\n\tFlow\n\tGoPie\n\tGoCR\n\tGoCRHB\n\tDefault: Guided")

	flag.BoolVar(&flags.ModeMain, "main", false, "set to run on main function")

	flag.StringVar(&flags.Settings, "settings", "", "Set some internal settings. For more info, see ../doc/usage.md")

	flag.BoolVar(&flags.CancelTestIfBugFound, "cancelTestIfBugFound", false, "Skip further fuzzing runs of a test if one bug has been found")

	// for experiments
	flag.BoolVar(&baseF.FinishIfBugFound, "finishIfBugFound", false, "Finish fuzzing as soon as a bug was found")

	flag.Parse()

	var mode string
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		mode = os.Args[1]
		flag.CommandLine.Parse(os.Args[2:])
		if help {
			helper.PrintHelpMode(mode)
			return ""
		}
	} else {
		if help {
			helper.PrintHelp()
			return ""
		}
		helper.PrintHelp()
		return ""
	}

	return mode
}
