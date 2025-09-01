// Copyright (c) 2025 Erik Kassubek
//
// File: help.go
// Brief: Function to print help/how to use
//
// Author: Erik Kassubek
// Created: 2025-05-19
//
// License: BSD-3-Clause

package helper

import (
	"fmt"
)

var (
	// help
	help1 = newFlagVal("h", "false", "", "Print help")
	help2 = newFlagVal("help", "false", "", "Print help")

	// submodes
	runMain      = newFlagVal("main", "false", "", "Set to run on main function. If not set, the unit tests are run")
	fuzzingModes = newFlagVal("mode", "", "", "Mode for fuzzing. Possible values are:", "\tGFuzz", "\tGFuzzHB", "\tGFuzzHBFlow", "\tFlow", "\tGoPie", "\tGoCR", "\tGoCRHB")

	// paths
	path  = newFlagVal("path", "", "", "Path to the program folder, for main: path to main file, for test: path to test folder")
	prog  = newFlagVal("prog", "", "-stat/-time/-notExec", "Name of the program")
	prog2 = newFlagVal("prog", "", "", "Name of the program")
	exec1 = newFlagVal("exec", "", "-main", "Name of the executable or test. If set for test, only this test will be executed, otherwise all tests will be run")
	exec2 = newFlagVal("exec", "", "", "Name of the executable or test")
	trace = newFlagVal("trace", "", "", "Path to the trace folder to replay")

	// scenarios
	scenarios = newFlagVal("scen", "", "", "Select which analysis scenario to run, e.g. -scen srd for the option s, r and d",
		"If not set, all scenarios are run.",
		"Options:",
		"\ts: Send on closed channel",
		"\tr: Receive on closed channel",
		"\tw: Done before add on waitGroup",
		"\tn: Close of closed channel",
		"\tb: Concurrent receive on channel",
		"\tl: Leaking routine",
		"\tu: Unlock of unlocked mutex",
		"\tc: Cyclic deadlock")
	noWarning = newFlagVal("noWarning", "false", "", "Only show critical bugs")
	onlyA     = newFlagVal("onlyActual", "false", "", "only test for actual bugs leading to panic and actual leaks. This will overwrite `scen`")

	// timeout
	timeoutRec    = newFlagVal("timeoutRec", "600", "", "Set the timeout in seconds for the recording. To disable set to -1")
	timeoutRep    = newFlagVal("timeoutRep", "900", "", "Set a timeout in seconds for the replay. To disable set to -1")
	timeoutFuz    = newFlagVal("timeoutFuz", "420", "", "Timeout of fuzzing per test/program in seconds. To Disable, set to -1")
	maxFuzzingRun = newFlagVal("maxFuzzingRuns", "-1", "", "Maximum number of fuzzing runs per test/prog. To Disable, set to -1")

	// statistics
	time    = newFlagVal("time", "false", "", "Measure the execution times of programs/tests and analysis")
	notExec = newFlagVal("notExec", "false", "", "Find never executed operations")
	stats   = newFlagVal("stats", "false", "", "Create statistics")

	// logging and output
	noInfo     = newFlagVal("noInfo", "false", "", "Do not show infos in the terminal (will only show results, errors, important and progress)")
	noProgress = newFlagVal("noProgress", "false", "", "Do not show progress info")
	output     = newFlagVal("output", "false", "", "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

	// continue
	cont         = newFlagVal("cont", "false", "", "Continue a partial analysis of tests")
	skipExisting = newFlagVal("skipExisting", "false", "", "If set, all tests that already have a results folder will be skipped. Also skips failed tests.")

	// memory
	noMemorySupervisor = newFlagVal("noMemorySupervisor", "false", "", "Disable the memory supervisor")
	maxNumberElem      = newFlagVal("maxNumberElements", "10000000", "Set the maximum number of elements in a trace. Traces with more elements will be skipped. To disable set to -1")

	// panic
	alwaysPanic = newFlagVal("panic", "false", "", "Panic if the analysis panics")

	// settings
	noFifo                = newFlagVal("ignoreFifo", "false", "", "Do not assume a FIFO ordering for buffered channels")
	ignoreCriticalSection = newFlagVal("ignoreCritSec", "false", "", "Ignore happens before relations of critical sections")
	ignoreAtomics         = newFlagVal("ignoreAtomics", "false", "", "Ignore atomic operations. Use to reduce memory required for large traces")
	replayAll             = newFlagVal("replayAll", "false", "", "Replay a bug even if it has already been confirmed")
	noRewrite             = newFlagVal("noRewrite", "true", "", "Do not rewrite/replay the trace file")
	keepTrace             = newFlagVal("keepTrace", "false", "", "If set, the traces are not deleted after analysis. Can result in very large output folders")
	settings              = newFlagVal("settings", "", "", "Set some internal settings. For more info, see ../doc/usage.md")
	cancelTestIfFound     = newFlagVal("cancelTestIfBugFound", "", "false", "Skip further fuzzing runs of a test if one bug has been found. Mostly used for benchmarks")
)

// flagValue is a struct to store one flag value and its description
type flagValue struct {
	name string   // name without -
	desc []string // description
	def  string   // default
	req  string   // required
}

// newFlagVal returns a new flag value
//
// Parameter
//   - name string: name of the flag
//   - def string: default value
//   - req string: additional info for required
//   - desc ...string: description of the flag (each value in new line)
func newFlagVal(name, def string, req string, desc ...string) flagValue {
	return flagValue{name, desc, def, req}
}

// get the string representation of a flag value
//
// Parameter:
//   - req bool: true if required
//
// Returns:
//   - string representation of fv
func (fv *flagValue) toString(req bool) string {
	res := fmt.Sprintf("-%-20s ", fv.name)

	res += fmt.Sprintf("%-10s", fv.def)

	if req {
		res += "req"
	} else {
		res += "opt"
	}

	if fv.req != "" {
		res += fmt.Sprintf(", req if %-24s", fv.req)
	} else {
		res += fmt.Sprintf("%-33s", fv.req)
	}

	if len(fv.desc) != 0 {
		res += fv.desc[0]
		for _, line := range fv.desc[1:] {
			res += fmt.Sprintf("\n%-68s%s", "", line)
		}
	}
	return res
}

// print the flag table description
func printFlagHeader() {
	fmt.Printf("%-22s%-10s%-36s%s\n\n", "flag", "default", "required/optional", "description")
}

// PrintHelp prints the main help header
func PrintHelp() {
	fmt.Println("Welcome to ADVOCATE")
	fmt.Println("")
	fmt.Println("AdvocateGo is an analysis tool for concurrent Go programs. It tries to detects concurrency bugs and gives diagnostic insight.")
	fmt.Println("")
	printHeader()
}

// PrintHelpMode prints the help for a specific mode
//
// Parameter:
//   - mode string: the mode
func PrintHelpMode(mode string) {
	switch mode {
	case "analysis":
		printHelpAnalysis()
	case "fuzzing":
		printHelpFuzzing()
	case "record", "recording":
		printHelpRecord()
	case "replay":
		printHelpReplay()
	default:
		fmt.Printf("Unknown mode '%s'\n\n", mode)
		printHeader()
	}

}

// printHeader prints the main help header
func printHeader() {
	fmt.Println("Usage: ./advocate [mode] [args]")
	fmt.Println("")
	fmt.Println("Advocate contains four different mode. These are:")
	fmt.Println("\trecord")
	fmt.Println("\treplay")
	fmt.Println("\tanalyzer")
	fmt.Println("\tfuzzing")
	fmt.Println("")
	fmt.Println("With 'record', the execution of a program or test can be recorded into a trace.")
	fmt.Println("With 'replay', a program or test can be forced to follow the execution schedule specified in a trace.")
	fmt.Println("With 'analyzer', a program or test can be recorded and then analyzed to find potential bugs. For some bugs, a rewrite and replay mechanism has been implemented to confirm the potential bugs.")
	fmt.Println("With 'fuzzing', different fuzzing approaches can be run on a program or test.")
	fmt.Print("\n\n")
	fmt.Println("For more information about the mode and there functionality, see the doc folder in the repository.")
	fmt.Println("For information on how to prepare the required runtime, see the usage file linked in the README")
	fmt.Println("")
	fmt.Println("To get information about one of the modes, including the required and optional tags, run\n\t./advocate [mode] -help.")
}

// print help for record mode
func printHelpRecord() {
	fmt.Println("Mode: record")
	fmt.Println("")

	printFlagHeader()

	// help
	fmt.Println(help1.toString(false))
	fmt.Println(help2.toString(false))

	// submodes
	fmt.Println(runMain.toString(false))

	// paths
	fmt.Println(path.toString(true))
	fmt.Println(prog.toString(false))
	fmt.Println(exec1.toString(false))

	// timeout
	fmt.Println(timeoutRec.toString(false))

	// statistics
	fmt.Println(time.toString(false))
	fmt.Println(notExec.toString(false))
	fmt.Println(stats.toString(false))

	// logging and output
	fmt.Println(noInfo.toString(false))
	fmt.Println(noProgress.toString(false))
	fmt.Println(output.toString(false))

	// continue
	fmt.Println(cont.toString(false))
	fmt.Println(skipExisting.toString(false))

	// memory
	fmt.Println(maxNumberElem.toString(false))
	fmt.Println(noMemorySupervisor.toString(false))

	// panic
	fmt.Println(alwaysPanic.toString(false))

	// settings
	fmt.Println(ignoreAtomics.toString(false))

}

// print help for REPLAY mode
func printHelpReplay() {
	fmt.Println("Mode: replay")
	fmt.Println("")

	printFlagHeader()

	// help
	fmt.Println(help1.toString(false))
	fmt.Println(help2.toString(false))

	// submodes
	fmt.Println(runMain.toString(false))

	// paths
	fmt.Println(path.toString(true))
	fmt.Println(exec2.toString(true))
	fmt.Println(trace.toString(true))

	// timeout
	fmt.Println(timeoutRep.toString(false))

	fmt.Println(output.toString(false))

	// memory
	fmt.Println(maxNumberElem.toString(false))
	fmt.Println(noMemorySupervisor.toString(false))

	// panic
	fmt.Println(alwaysPanic.toString(false))

	// settings
	fmt.Println(ignoreAtomics.toString(false))
}

// print help for analysis mode
func printHelpAnalysis() {
	fmt.Println("Mode: analysis")
	fmt.Println("")

	printFlagHeader()

	// help
	fmt.Println(help1.toString(false))
	fmt.Println(help2.toString(false))

	// submodes
	fmt.Println(runMain.toString(false))

	// paths
	fmt.Println(path.toString(true))
	fmt.Println(prog.toString(false))
	fmt.Println(exec1.toString(false))

	// scenarios
	fmt.Println(scenarios.toString(false))
	fmt.Println(noWarning.toString(false))
	fmt.Println(onlyA.toString(false))

	// timeout
	fmt.Println(timeoutRec.toString(false))
	fmt.Println(timeoutRep.toString(false))

	// statistics
	fmt.Println(time.toString(false))
	fmt.Println(notExec.toString(false))
	fmt.Println(stats.toString(false))

	// logging and output
	fmt.Println(noInfo.toString(false))
	fmt.Println(noProgress.toString(false))
	fmt.Println(output.toString(false))

	// continue
	fmt.Println(cont.toString(false))
	fmt.Println(skipExisting.toString(false))

	// memory
	fmt.Println(maxNumberElem.toString(false))
	fmt.Println(noMemorySupervisor.toString(false))

	// panic
	fmt.Println(alwaysPanic.toString(false))

	// settings
	fmt.Println(noFifo.toString(false))
	fmt.Println(ignoreCriticalSection.toString(false))
	fmt.Println(ignoreAtomics.toString(false))
	fmt.Println(replayAll.toString(false))
	fmt.Println(noRewrite.toString(false))
	fmt.Println(keepTrace.toString(false))
}

// print help for fuzzing mode
func printHelpFuzzing() {
	fmt.Println("Mode: fuzzing")
	fmt.Println("")

	printFlagHeader()

	// help
	fmt.Println(help1.toString(false))
	fmt.Println(help2.toString(false))

	// submodes
	fmt.Println(runMain.toString(false))
	fmt.Println(fuzzingModes.toString(true))

	// paths
	fmt.Println(path.toString(true))
	fmt.Println(prog2.toString(true))
	fmt.Println(exec1.toString(false))

	// scenarios
	fmt.Println(scenarios.toString(false))
	fmt.Println(noWarning.toString(false))
	fmt.Println(onlyA.toString(false))

	// timeout
	fmt.Println(timeoutRec.toString(false))
	fmt.Println(timeoutRep.toString(false))
	fmt.Println(timeoutFuz.toString(false))
	fmt.Println(maxFuzzingRun.toString(false))

	// statistics
	fmt.Println(time.toString(false))
	fmt.Println(notExec.toString(false))
	fmt.Println(stats.toString(false))

	// logging and output
	fmt.Println(noInfo.toString(false))
	fmt.Println(noProgress.toString(false))
	fmt.Println(output.toString(false))

	// memory
	fmt.Println(maxNumberElem.toString(false))
	fmt.Println(maxNumberElem.toString(false))

	// panic
	fmt.Println(alwaysPanic.toString(false))

	// settings
	fmt.Println(noFifo.toString(false))
	fmt.Println(ignoreCriticalSection.toString(false))
	fmt.Println(ignoreAtomics.toString(false))
	fmt.Println(replayAll.toString(false))
	fmt.Println(noRewrite.toString(false))
	fmt.Println(keepTrace.toString(false))
	fmt.Println(settings.toString(false))
	fmt.Println(cancelTestIfFound.toString(false))
}
