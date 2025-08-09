// File: flags.go
// Brief: Flags
//
// Created: 2025-08-09
//
// License: BSD-3-Clause

package helper

import "flag"

// Flags
var (
	PathToGoCR string

	TracePath string
	ProgPath  string

	ProgName string
	ExecName string

	TimeoutRecording int
	TimeoutReplay    int
	TimeoutFuzzing   int
	RecordTime       bool

	MaxFuzzingRun int

	IgnoreAtomics bool

	KeepTraces bool

	Statistics bool

	FuzzingMode string

	ModeMain bool

	NoWarning bool

	NoInfo     bool
	NoProgress bool

	AlwaysPanic bool

	Settings string
	Output   bool

	CancelTestIfBugFound bool

	NoMemorySupervisor bool

	MaxNumberElements int

	SameElemTypeInSC     bool
	ScSize               int
	FuzzingWithoutReplay bool
	FinishIfBugFound     = false

	help bool
)

// ReadFlags reads the command line flags
//
// Returns:
//   - bool: true if help is requested
func ReadFlags() bool {
	flag.BoolVar(&help, "h", false, "Print help")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.StringVar(&ProgPath, "path", "", "Path to the program folder, for main: path to main file, for test: path to test folder")

	flag.StringVar(&ProgName, "prog", "", "Name of the program")
	flag.StringVar(&ExecName, "exec", "", "Name of the executable or test")

	flag.StringVar(&TracePath, "trace", "", "Path to the trace folder to replay")

	flag.IntVar(&TimeoutRecording, "timeoutRec", 600, "Set the timeout in seconds for the recording. Default: 600s. To disable set to -1")
	flag.IntVar(&TimeoutReplay, "timeoutRep", 900, "Set a timeout in seconds for the replay. Default: 600s. To disable set to -1")

	flag.IntVar(&TimeoutFuzzing, "timeoutFuz", 420, "Timeout of fuzzing per test/program in seconds. Default: 7min. To Disable, set to -1")
	flag.IntVar(&MaxFuzzingRun, "maxFuzzingRuns", -1, "Maximum number of fuzzing runs per test/prog. Default: -1. To Disable, set to -1")

	flag.BoolVar(&RecordTime, "time", false, "measure the runtime")

	flag.BoolVar(&NoMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.BoolVar(&IgnoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")

	flag.BoolVar(&KeepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")

	flag.BoolVar(&Statistics, "stats", false, "Create statistics.")

	flag.BoolVar(&NoWarning, "noWarning", false, "Only show critical bugs")

	flag.BoolVar(&NoInfo, "noInfo", false, "Do not show infos in the terminal (will only show results, errors, important and progress)")
	flag.BoolVar(&NoProgress, "noProgress", false, "Do not show progress info")

	flag.BoolVar(&AlwaysPanic, "panic", false, "Panic if the analysis panics")

	flag.BoolVar(&Output, "output", false, "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

	flag.IntVar(&MaxNumberElements, "maxNumberElements", 10000000, "Set the maximum number of elements in a trace. Traces with more elements will be skipped. To disable set -1. Default: 10000000")

	flag.StringVar(&FuzzingMode, "fuzzingMode", "",
		"Mode for fuzzing. Possible values are:\n\tGFuzz\n\tGFuzzHB\n\tGFuzzHBFlow\n\tFlow\n\tGoPie\n\tGoCR\n\tGoCRHB")

	// partially implemented by may not work, therefore disables, enable again when fixed
	flag.BoolVar(&ModeMain, "main", false, "set to run on main function")

	flag.StringVar(&Settings, "settings", "", "Set some internal settings. For more info, see ../doc/usage.md")

	flag.BoolVar(&CancelTestIfBugFound, "cancelTestIfBugFound", false, "Skip further fuzzing runs of a test if one bug has been found")

	// for experiments
	flag.BoolVar(&SameElemTypeInSC, "sameElemTypeInSC", false, "Only allow elements of the same type in the same SC")
	flag.IntVar(&ScSize, "scSize", -1, "max number of elements in SC")
	flag.BoolVar(&FuzzingWithoutReplay, "fuzzingWithoutReplay", false, "Disable replay before SC in fuzzing")
	flag.BoolVar(&FinishIfBugFound, "finishIfBugFound", false, "Finish fuzzing as soon as a bug was found")

	flag.Parse()

	TracePath = CleanPathHome(TracePath)
	ProgPath = CleanPathHome(ProgPath)

	return help
}
