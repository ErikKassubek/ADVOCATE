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
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"advocate/fuzzing"
	"advocate/memory"
	"advocate/stats"
	"advocate/timer"
	"advocate/toolchain"
	"advocate/utils"
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

	noInfo bool

	cont bool

	noMemorySupervisor bool

	alwaysPanic bool

	settings string
	output   bool
)

// Main function
func main() {
	flag.BoolVar(&help, "h", false, "Print help")

	flag.StringVar(&progPath, "path", "", "Path to the program folder, for main: path to main file, for test: path to test folder")

	flag.StringVar(&progName, "prog", "", "Name of the program")
	flag.StringVar(&execName, "exec", "", "Name of the executable or test")

	flag.StringVar(&tracePath, "trace", "", "Path to the trace folder to replay or analyze")

	flag.IntVar(&timeoutRecording, "timeoutRec", 600, "Set the timeout in seconds for the recording. Default: 600s. To disable set to -1")
	flag.IntVar(&timeoutReplay, "timeoutRep", -1, "Set a timeout in seconds for the replay. If not set, it is set to 500 * recording time")

	flag.IntVar(&timeoutFuzzing, "timeoutFuz", 420, "Timeout of fuzzing per test/program in seconds. Default: 7 min. To Disable, set to -1")
	flag.IntVar(&maxFuzzingRun, "maxFuzzingRun", 100, "Maximum number of fuzzing runs per test/prog. Default: 100. To Disable, set to -1")

	flag.BoolVar(&recordTime, "time", false, "measure the runtime")

	flag.BoolVar(&noFifo, "noFifo", false, "Do not assume a FIFO ordering for buffered channels")
	flag.BoolVar(&ignoreCriticalSection, "ignCritSec", false, "Ignore happens before relations of critical sections (default false)")
	flag.BoolVar(&ignoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")

	flag.BoolVar(&rewriteAll, "replayAll", false, "Replay a bug even if it has already been confirmed")
	rewriteAll = true

	flag.BoolVar(&noRewrite, "noRewrite", false, "Do not rewrite the trace file (default false)")
	flag.BoolVar(&keepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")
	flag.BoolVar(&skipExisting, "skipExisting", false, "If set, all tests that already have a results folder will be skipped. Also skips failed tests.")

	flag.BoolVar(&notExec, "notExec", false, "Find never executed operations")
	flag.BoolVar(&statistics, "stats", false, "Create statistics")

	flag.BoolVar(&noWarning, "noWarning", false, "Only show critical bugs")

	flag.BoolVar(&cont, "cont", false, "Continue a partial analysis of tests")

	flag.BoolVar(&noMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.BoolVar(&noInfo, "noInfo", false, "Do not show infos in the terminal (will only show results, errors and important)")

	flag.BoolVar(&alwaysPanic, "panic", false, "Panic if the analysis panics")

	flag.BoolVar(&output, "output", false, "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

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

	flag.StringVar(&settings, "settings", "", "Settings")

	flag.Parse()

	var mode string
	if len(os.Args) >= 2 {
		mode = os.Args[1]
		flag.CommandLine.Parse(os.Args[2:])
		if help {
			printHelpMode(mode)
			return
		}
	} else {
		if help {
			printHelp()
			return
		}
		fmt.Println("No mode selected")
		fmt.Println("Select one mode from 'analysis', 'fuzzing' or 'recording'")
		printHelp()
		return
	}

	// If -main is set, the path needs to be the path to the main file
	// If the given path is to a folder, check if a main.go file exists in this folder
	// If so, fix the path. Otherwise return error and finish
	if modeMain {
		var err error
		progPath, err = utils.GetMainPath(progPath)
		if err != nil {
			utils.LogError("Could not find main file. If -main is set, -path should point to the main file.")
			utils.LogError(err)
			return
		}
	}

	utils.LogInit(noInfo, !alwaysPanic)

	utils.SetSettings(settings, maxFuzzingRun, fuzzingMode)

	progPathDir := utils.GetDirectory(progPath)
	timer.Init(recordTime, progPathDir)
	timer.Start(timer.Total)
	defer timer.Stop(timer.Total)

	execPath, _ := os.Executable()
	pathToAdvocate = filepath.Dir(filepath.Dir(execPath))

	advocatePathSplit := strings.Split(pathToAdvocate, string(os.PathSeparator))
	if advocatePathSplit[len(advocatePathSplit)-1] != "ADVOCATE" {
		utils.LogError("Could not determine ADVOCATE folder. Keep the toolchain and go-patch in the ADVOCATE folder. Do not rename the ADVOCATE folder.")
		return
	}

	if !noMemorySupervisor {
		go memory.Supervisor() // cancel analysis if not enough ram
	}

	// don't run any HB Analysis for direct GFuzz, GoPie and GoPie+
	if mode == "fuzzing" && (fuzzingMode == fuzzing.GFuzz ||
		fuzzingMode == fuzzing.GoPie || fuzzingMode == fuzzing.GoPiePlus) {
		scenarios = "-"
		onlyAPanicAndLeak = true
	}

	analysisCases, err := parseAnalysisCases(scenarios)
	if err != nil {
		utils.LogError("Could not read analysis cases: ", err)
		return
	}

	toolchain.SetFlags(noRewrite, analysisCases, ignoreAtomics,
		!noFifo, ignoreCriticalSection, rewriteAll, onlyAPanicAndLeak,
		timeoutRecording, timeoutReplay, rewriteAll, noWarning, tracePath, output)

	modeMainTest := "test"
	if modeMain {
		modeMainTest = "main"
	}

	checkGoMod()

	if modeMain && execName == "" {
		utils.LogError("Could not determine executable name from go.mod. Provide with -exec [ExecutableName]")
		panic(fmt.Errorf("Could not determine executable name"))
	}

	// TODO: mode for analysis of already existing trace

	switch mode {
	case "analysis":
		modeToolchain(modeMainTest, true, true, true)
	case "fuzzing":
		modeFuzzing()
	case "record":
		keepTraces = true
		modeToolchain(modeMainTest, true, false, false)
	case "replay":
		modeToolchain(modeMainTest, false, false, true)
	default:
		utils.LogErrorf("Unknown mode %s\n", os.Args[1])
		utils.LogError("Select one mode from  'analysis', 'fuzzing' or 'record'")
		printHelp()
	}

	numberResults, numberResultsConf, numberErr, numberTimeout := utils.GetLoggingNumbers()
	if numberErr == 0 {
		utils.LogInfo("Finished with 0 errors")
	} else {
		utils.LogErrorf("Finished with %d errors", numberErr)
	}
	if numberTimeout == 0 {
		utils.LogInfo("No timeouts occur")
	} else {
		utils.LogErrorf("%d timeouts occurred", numberTimeout)
	}
	if mode == "analysis" || mode == "fuzzing" {
		if numberResults == 0 {
			utils.LogInfo("No bugs have been found/indicated")
		} else {
			utils.LogResultf(false, "%d bugs have been indicated", numberResults)
			utils.LogResultf(false, "%d bugs have been confirmed", numberResultsConf)
		}
	}
	timer.UpdateTimeFileOverview(progName, "*Total*")
	utils.LogInfo("Total time: ", timer.GetTime(timer.Total))
}

// modeFuzzing starts the fuzzing
func modeFuzzing() {
	if progName == "" {
		utils.LogError("Provide a name for the analyzed program. Set with -prog [name]")
		return
	}

	progPath, err := utils.CheckPath(progPath)
	if err != nil {
		utils.LogError("Error on checking prog path: ", err)
		panic(err)
	}

	err = fuzzing.Fuzzing(modeMain, fuzzingMode, pathToAdvocate, progPath,
		progName, execName, ignoreAtomics, recordTime, notExec, statistics,
		keepTraces, cont, timeoutFuzzing, maxFuzzingRun)
	if err != nil {
		utils.LogError("Fuzzing Failed: ", err.Error())
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
	progPath, err := utils.CheckPath(progPath)
	if err != nil {
		utils.LogError("Error on checking prog path: ", err)
		panic(err)
	}

	if !record && (analysis || replay) {
		tracePath, err = utils.CheckPath(tracePath)
		if err != nil {
			utils.LogError("Error on checking trace path: ", err)
			panic(err)
		}
	}

	if mode == "test" && !record && replay && execName == "" {
		utils.LogError("When running replay of test without recording, -exec [TestName] must be set")
		panic("When running replay of test without recording, -exec [TestName] must be set")
	}

	err = toolchain.Run(mode, pathToAdvocate, progPath, "", record, analysis,
		replay, execName, progName, execName, -1, "", ignoreAtomics, recordTime,
		notExec, statistics, keepTraces, skipExisting, true, cont, 0, 0)
	if err != nil {
		utils.LogError("Failed to run toolchain: ", err.Error())
	}

	if statistics {
		// TODO: check if this
		err = stats.CreateStatsTotal(progPath, progName)
		if err != nil {
			utils.LogError("Failed to create stats total: ", err.Error())
		}
	}
}

// Parse the given analysis cases
//
// Parameter:
//   - cases string: The string of analysis cases to parse
//
// Returns:
//   - map[string]bool: A map of the analysis cases and if they are set
//   - error: An error if the cases could not be parsed
func parseAnalysisCases(cases string) (map[string]bool, error) {
	analysisCases := map[string]bool{
		"all":              false, // all cases enabled
		"sendOnClosed":     false,
		"receiveOnClosed":  false,
		"doneBeforeAdd":    false,
		"closeOnClosed":    false,
		"concurrentRecv":   false,
		"leak":             false,
		"unlockBeforeLock": false,
		"mixedDeadlock":    false,
		"resourceDeadlock": false,
	}

	if cases == "-" {
		return analysisCases, nil
	}

	if cases == "" {
		for c := range analysisCases {
			analysisCases[c] = true
		}

		// remove when implemented
		analysisCases["mixedDeadlock"] = false

		return analysisCases, nil
	}

	for _, c := range cases {
		switch c {
		case 's':
			analysisCases["sendOnClosed"] = true
		case 'r':
			analysisCases["receiveOnClosed"] = true
		case 'w':
			analysisCases["doneBeforeAdd"] = true
		case 'n':
			analysisCases["closeOnClosed"] = true
		case 'b':
			analysisCases["concurrentRecv"] = true
		case 'l':
			analysisCases["leak"] = true
		case 'u':
			analysisCases["unlockBeforeLock"] = true
		case 'c':
			analysisCases["resourceDeadlock"] = true
		// case 'm':
		// analysisCases["mixedDeadlock"] = true
		default:
			return nil, fmt.Errorf("Invalid analysis case: %c", c)
		}
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

// printHelp prints the usage help. Can be called with -h
func printHelp() {
	println("Usage: ./advocate [mode] [options]\n")
	println("There are different modes of operation:")
	println("1. Analysis")
	println("2. Fuzzing")
	println("3. Record")
	println("4. Replay")
	println("\n\n")
	printHelpMode("analysis")
	printHelpMode("fuzzing")
	printHelpMode("record")
	printHelpMode("replay")
}

// printHelpMode prints the help for one mode
//
// Parameter:
//   - mode string: the mode (analysis or fuzzing)
func printHelpMode(mode string) {
	switch mode {
	case "analysis":
		println("Mode: analysis")
		println("Analyze tests")
		println("This runs the analysis on tests")
		println("Usage: ./advocate analysis [options]")
		println("  -main                  Run on the main function instead on tests")
		println("  -path [path]           Path to the folder containing the program and tests, if main, path to the file containing the main function (required)")
		println("  -exec [name]           For tests, name of the test to run (do not set to run all tests). For main name of the executable (only required if fo.mod cannot be found)")
		println("  -prog [name]           Name of the program (required if -recordTime, -notExec or -stats is set)")
		println("  -timeoutRec [second]   Set a timeout in seconds for the recording")
		println("  -timeoutRep [second]   Set a timeout in seconds for the replay")
		println("  -ignoreAtomics         Set to ignore atomics in replay")
		println("  -recordTime            Set to record runtimes")
		println("  -notExec               Set to determine never executed operations")
		println("  -stats                 Set to create statistics")
		println("  -keepTrace             Do not delete the trace files after analysis finished")
	case "fuzzing":
		println("Mode: fuzzing")
		println("Run fuzzing")
		println("This creates and updates the information required for as well as executes the fuzzing runs and analysis")
		println("Usage: ./advocate fuzzing [options]")
		println("  -main                  Run on the main function instead on tests")
		println("  -path [folder]         If -main, path to the file containing the main function, otherwise path to the program folder")
		println("  -prog [name]           Name of the program")
		println("  -exec [name]           Name of the test to run (do not set to run all tests)")
		println("  -fuzzingMode [mode]    Mode of fuzzing:")
		println("\tGFuzz\n\tGFuzzHB\n\tGFuzzHBFlow\n\tFlow\n\tGoPie\n\tGoPie+\n\tGoPieHB")
		println("  -timeoutRec [second]   Set a timeout in seconds for the recording")
		println("  -timeoutRep [second]   Set a timeout in seconds for the replay")
		println("  -ignoreAtomics         Set to ignore atomics in replay")
		println("  -recordTime            Set to record runtimes")
		println("  -notExec               Set to determine never executed operations")
		println("  -stats                 Set to create statistics")
		println("  -keepTrace             Do not delete the trace files after analysis finished")
	case "record":
		println("Mode: record")
		println("Record traces")
		println("This records program or test traces without running any analysis")
		println("Usage: ./advocate record [options]")
		println("  -main                  Run on the main function instead on tests")
		println("  -path [folder]         If -main, path to the file containing the main function, otherwise path to the program folder")
		println("  -exec [name]           Name of the test to run (do not set to run all tests)")
		println("  -timeoutRec [second]   Set a timeout in seconds for the recording")
		println("  -recordTime            Set to record runtimes")
		println("  -notExec               Set to determine never executed operations")
		println("  -stats                 Set to create statistics")
	case "replay":
		println("Mode: replay")
		println("Replay the trace")
		println("This will take a prerecorded trace and replay it")
		println("Usage: ./advocate record [options]")
		println("  -main                  Run on the main function instead on tests")
		println("  -path [folder]         If -main, path to the file containing the main function, otherwise path to the program folder")
		println("  -trace [folder]        Path to the trace that should be executed")
		println("  -exec [name]           Name of the test to run (only when -main is not set)")
		println("  -timeoutRec [second]   Set a timeout in seconds for the recording")
		println("  -recordTime            Set to record runtimes")
		println("  -notExec               Set to determine never executed operations")
		println("  -stats                 Set to create statistics")
	default:
		println("Mode: unknown")
		printHelp()
	}
}

// checkGoMod checks the version of the program to be analyzed.
// Advocate is implemented in and for go1.24. It the analyzed program has another
// version, especially if the other version is also installed on the machine,
// this can lead to problems. checkGoMod therefore reads the version of the
// analyzed program and if its not 1.24, a warning and information is printed
// to the terminal
// Additionally it reads the module name from the go.mod file.
// If -main is set, but -exec is not set it will try to set the
// execname value. If no module value is found, the program will panic
func checkGoMod() {
	var goModPath string

	if progPath == "" {
		return
	}

	// Search for go.mod
	err := filepath.WalkDir(utils.GetDirectory(progPath), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == "go.mod" {
			goModPath = path
			return filepath.SkipAll // Stop searching after finding the first one
		}
		return nil
	})

	if goModPath == "" {
		utils.LogInfo("Could not find go.mod")
		return
	}

	// Open and read go.mod
	file, err := os.Open(goModPath)
	if err != nil {
		utils.LogInfo("Could not find go.mod")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// check for module name
		if modeMain && execName == "" && strings.HasPrefix(line, "module") {
			s := strings.Split(line, " ")
			if len(s) < 2 {
				continue
			}

			execName = s[1]
			continue
		}

		// check for version
		if strings.HasPrefix(line, "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "go "))

			versionSplit := strings.Split(version, ".")

			if len(versionSplit) < 2 {
				utils.LogError("Invalid go version")
			}

			if versionSplit[0] != "1" || versionSplit[1] != "24" {
				errString := "ADVOCATE is implemented for go version 1.24. "
				errString += fmt.Sprintf("Found version %s. ", version)
				errString += fmt.Sprintf("This may result in the analysis not working correctly, especially if go %s.%s is installed on the computer. ", versionSplit[0], versionSplit[1])
				errString += "The message 'package advocate is not in std' in the output.log file may indicate this."
				// errString += `'/home/.../go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/advocate' or 'package advocate is not in std' in the output files may indicate an incompatible go version.`
				utils.LogImportant(errString)
			}

			return
		}
	}

	utils.LogError("Could not determine go version")
}
