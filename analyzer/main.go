// Copyright (c) 2024 Erik Kassubek
//
// File: main.go
// Brief: Main file and starting point for the analyzer
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"analyzer/analysis"
	"analyzer/fuzzing"
	"analyzer/io"
	"analyzer/memory"
	"analyzer/results"
	"analyzer/rewriter"
	"analyzer/stats"
	"analyzer/timer"
	"analyzer/toolchain"
	"analyzer/utils"
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
	recordTime       bool

	resultFolder     string
	resultFolderTool string
	outM             string
	outR             string
	outT             string

	noFifo                bool
	ignoreCriticalSection bool
	ignoreAtomics         bool
	ignoreRewrite         string

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

	cont bool

	noMemorySupervisor bool
)

// Main function
func main() {
	flag.BoolVar(&help, "h", false, "Print help")

	flag.StringVar(&tracePath, "trace", "", "Path to the trace folder to analyze or rewrite")
	flag.StringVar(&progPath, "path", "", "Path to the program folder, for main: path to main file, for test: path to test folder")

	flag.StringVar(&progName, "prog", "", "Name of the program")
	flag.StringVar(&execName, "exec", "", "Name of the executable or test")

	flag.IntVar(&timeoutRecording, "timeoutRec", 600, "Set the timeout in seconds for the recording. Default: 600s. To disable set to -1")
	flag.IntVar(&timeoutReplay, "timeoutRep", -1, "Set a timeout in seconds for the replay. If not set, it is set to 500 * recording time")
	flag.BoolVar(&recordTime, "time", false, "measure the runtime")

	flag.StringVar(&resultFolder, "out", "", "Path to where the result file should be saved.")
	flag.StringVar(&resultFolderTool, "resultTool", "", "Path where the advocateResult folder created by the pipeline is located")
	flag.StringVar(&outM, "outM", "results_machine", "Name for the result machine file")
	flag.StringVar(&outR, "outR", "results_readable", "Name for the result readable file")
	flag.StringVar(&outT, "outT", "rewrittenTrace", "Name for the rewritten traces")

	flag.BoolVar(&noFifo, "noFifo", false, "Do not assume a FIFO ordering for buffered channels")
	flag.BoolVar(&ignoreCriticalSection, "ignCritSec", false, "Ignore happens before relations of critical sections (default false)")
	flag.BoolVar(&ignoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")


	flag.BoolVar(&rewriteAll, "replayAll", false, "If the instance of a bug was already confirmed, a later occurrence will normally not be replayed. Setting replayAll will force the replay for
	all occurrences of a bug, even if it has been confirmed before")
	rewriteAll = true

	flag.BoolVar(&noRewrite, "noRewrite", false, "Do not rewrite the trace file (default false)")
	flag.BoolVar(&keepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")
	flag.BoolVar(&skipExisting, "skipExisting", false, "If set, all tests that already have a results folder will be skipped. Also skips failed tests.")

	flag.BoolVar(&notExec, "notExec", false, "Find never executed operations, *notExec, *stats")
	flag.BoolVar(&statistics, "stats", false, "Create statistics")

	flag.BoolVar(&noWarning, "noWarning", false, "Only show critical bugs")

	flag.BoolVar(&cont, "cont", false, "Continue a partial analysis of tests")

	flag.BoolVar(&noMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.StringVar(&scenarios, "scen", "", "Select which analysis scenario to run, e.g. -scen srd for the option s, r and d."+
		"If not set, all scenarios are run.\n"+
		"Options:\n"+
		"\ts: Send on closed channel\n"+
		"\tr: Receive on closed channel\n"+
		"\tw: Done before add on waitGroup\n"+
		"\tn: Close of closed channel\n"+
		"\tb: Concurrent receive on channel\n"+
		"\tl: Leaking routine\n"+
		"\tp: Select case without partner\n"+
		"\tu: Unlock of unlocked mutex\n"+
		"\tc: Cyclic deadlock\n",
	)
	// "\tm: Mixed deadlock\n"

	flag.BoolVar(&onlyAPanicAndLeak, "onlyActual", false, "only test for actual bugs leading to panic and actual leaks. This will overwrite `scen`")

	flag.StringVar(&fuzzingMode, "fuzzingMode", "",
		"Mode for fuzzing. Possible values are:\n\tGFuzz\n\tGFuzzHB\n\tGFuzzHBFlow\n\tFlow\n\tGoPie\n\tGoPieHB")

	// partially implemented by may not work, therefore disables, enable again when fixed
	// flag.BoolVar(&modeMain, "main", false, "set to run on main function")

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
		fmt.Println("Select one mode from 'run', 'stats', 'explain' or 'check'")
		printHelp()
		return
	}

	timer.Init(recordTime, progPath)
	timer.Start(timer.Total)
	defer timer.Stop(timer.Total)

	execPath, _ := os.Executable()
	pathToAdvocate = filepath.Dir(filepath.Dir(execPath))

	advocatePathSplit := strings.Split(pathToAdvocate, string(os.PathSeparator))
	if advocatePathSplit[len(advocatePathSplit)-1] != "ADVOCATE" {
		utils.LogError("Could not determine ADVOCATE folder. Keep the analyzer and go-patch in the ADVOCATE folder. Do not rename the ADVOCATE folder.")
		return
	}

	if resultFolder == "" {
		resultFolder, err := getFolderTrace(tracePath)
		if err != nil {
			utils.LogError("Could not get folder trace: ", err)
			return
		}

		if (resultFolder)[len(resultFolder)-1] != os.PathSeparator {
			resultFolder += string(os.PathSeparator)
		}
	}

	if !noMemorySupervisor {
		go memory.Supervisor() // cancel analysis if not enough ram
	}

	// outMachine := filepath.Join(resultFolder, outM) + ".log"
	// outReadable := filepath.Join(resultFolder, outR) + ".log"
	// newTrace := filepath.Join(resultFolder, outT)
	if ignoreRewrite != "" {
		ignoreRewrite = filepath.Join(resultFolder, ignoreRewrite)
	}

	// don't run any HB Analysis for direct GFuzz and GoPie
	if mode == "fuzzing" && (fuzzingMode == fuzzing.GFuzz || fuzzingMode == fuzzing.GoPie) {
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
		timeoutRecording, timeoutReplay)

	// function injection to prevent circle import
	toolchain.InitFuncAnalyzer(modeAnalyzer)

	switch mode {
	case "analysis":
		if modeMain {
			modeToolchain("main", 0)
		} else {
			modeToolchain("test", 0)
		}
	// case "analyze":
	// 	// here the parameter need to stay, because the function is used in the
	// 	// toolchain package via function injection
	// 	modeAnalyzer(tracePath, noRewrite, analysisCases, outReadable,
	// 		outMachine, ignoreAtomics, fifo, ignoreCriticalSection,
	// 		rewriteAll, newTrace, ignoreRewrite,
	// 		-1, onlyAPanicAndLeak)
	case "fuzzing":
		modeFuzzing()
	default:
		utils.LogErrorf("Unknown mode %s\n", os.Args[1])
		utils.LogError("Select one mode from 'run', 'stats', 'explain' or 'check'")
		printHelp()
	}

	numberErr, numberTimeout := utils.GetNumberErr()
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
	timer.UpdateTimeFileOverview(progName, "*Total*")
	utils.LogInfo("Total time: ", timer.GetTime(timer.Total))
}

// Starting point for fuzzing
func modeFuzzing() {
	if progName == "" {
		utils.LogError("Provide a name for the analyzed program. Set with -prog [name]")
		return
	}

	checkProgPath()
	checkVersion()

	err := fuzzing.Fuzzing(modeMain, fuzzingMode, pathToAdvocate, progPath, progName, execName,
		ignoreAtomics, recordTime, notExec, statistics,
		keepTraces, cont)
	if err != nil {
		utils.LogError("Fuzzing Failed: ", err.Error())
	}
}

// Start point for the toolchain
// This will run, analyze and replay a given program or test
func modeToolchain(mode string, numRerecorded int) {
	checkProgPath()
	checkVersion()
	err := toolchain.Run(mode, pathToAdvocate, progPath, "", execName, progName, execName,
		numRerecorded, -1, "", ignoreAtomics, recordTime, notExec, statistics, keepTraces, true, skipExisting, cont, 0, 0)
	if err != nil {
		utils.LogError("Failed to run toolchain: ", err.Error())
	}

	if statistics {
		err = stats.CreateStatsTotal(progPath, progName)
		if err != nil {
			utils.LogError("Failed to create stats total: ", err.Error())
		}
	}
}

// modeAnalyzer is the starting point to the analyzer.
// This function will read the trace at a stored path, analyze it and,
// if needed, rewrite the trace.
//
// Parameter:
//   - pathTrace string: path to the trace to be analyzed
//   - noRewrite bool: if set, rewrite is disabled
//   - analysisCases map[string]bool: map of analysis cases to run
//   - outReadable string: path to the readable result file
//   - outMachine string: path to the machine result file
//   - ignoreAtomics bool: if true, atomics are ignored for replay
//   - fifo bool: assume, that the channels work as a fifo queue
//   - ignoreCriticalSection bool: ignore the ordering of lock/unlock for the hb analysis
//   - rewriteAll bool: rewrite bugs that have been rewritten before
//   - newTrace string: path to where the rewritten trace should be created
//   - fuzzingRun int: number of fuzzing run (0 for recording, then always add 1)
//   - onlyAPanicAndLeak bool: only check for actual leaks and panics, do not calculate HB information
//
// Returns:
//   - error
func modeAnalyzer(pathTrace string, noRewrite bool,
	analysisCases map[string]bool, outReadable string, outMachine string,
	ignoreAtomics bool, fifo bool, ignoreCriticalSection bool,
	rewriteAll bool, newTrace string, fuzzingRun int, onlyAPanicAndLeak bool) error {

	if pathTrace == "" {
		return fmt.Errorf("Please provide a path to the trace files. Set with -trace [folder]")
	}

	// run the analysis and, if requested, create a reordered trace file
	// based on the analysis results

	results.InitResults(outReadable, outMachine)

	numberOfRoutines, numberElems, err := io.CreateTraceFromFiles(pathTrace, ignoreAtomics)
	if err != nil {
		fmt.Println("Could not open trace: ", err.Error())
	}

	if numberElems == 0 {
		utils.LogInfof("Trace at %s does not contain any elements", pathTrace)
		return nil
	}

	utils.LogInfof("Read trace with %d elements in %d routines", numberElems, numberOfRoutines)

	analysis.SetNoRoutines(numberOfRoutines)

	if onlyAPanicAndLeak {
		utils.LogInfo("Start Analysis for actual panics and leaks")
	} else if analysisCases["all"] {
		utils.LogInfo("Start Analysis for all scenarios")
	} else {
		info := "Start Analysis for the following scenarios: "
		for key, value := range analysisCases {
			if value {
				info += (key + ",")
			}
		}
		utils.LogInfo(info)
	}

	analysis.RunAnalysis(fifo, ignoreCriticalSection, analysisCases, fuzzingRun >= 0, onlyAPanicAndLeak)

	if memory.WasCanceled() {
		// analysis.LogSizes()
		analysis.Clear()
		if memory.WasCanceledRAM() {
			return fmt.Errorf("Analysis was canceled due to insufficient small RAM")
		}
		return fmt.Errorf("Analysis was canceled due to unexpected panic")
	}
	utils.LogInfo("Analysis finished")

	numberOfResults, err := results.CreateResultFiles(noWarning, true)
	if err != nil {
		utils.LogError("Error in printing summary: ", err.Error())
	}

	// collect the required data to decide whether run is interesting
	// and to create the mutations
	if fuzzingRun >= 0 {
		fuzzing.ParseTrace(&analysis.MainTrace)
	}

	if noRewrite {
		utils.LogInfo("Skip rewrite")
		return nil
	}

	numberRewrittenTrace := 0
	failedRewrites := 0
	notNeededRewrites := 0
	utils.LogInfo("Start rewriting")

	if err != nil {
		utils.LogError("Failed to rewrite: ", err)
		return nil
	}

	if memory.WasCanceled() {
		utils.LogError("Could not run rewrite: Not enough RAM")
	}

	rewrittenBugs := make(map[utils.ResultType][]string) // bugtype -> paths string

	file := filepath.Base(pathTrace)
	rewriteNr := "0"
	spl := strings.Split(file, "_")
	if len(spl) > 1 {
		rewriteNr = spl[len(spl)-1]
	}

	for resultIndex := 0; resultIndex < numberOfResults; resultIndex++ {
		needed, err := rewriteTrace(outMachine,
			newTrace+"_"+strconv.Itoa(resultIndex+1)+"/", resultIndex, numberOfRoutines, &rewrittenBugs)

		if !needed {
			notNeededRewrites++
			fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
		} else if err != nil {
			failedRewrites++
			fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
		} else { // needed && err == nil
			numberRewrittenTrace++
			fmt.Printf("Bugreport info: %s_%d,suc\n", rewriteNr, resultIndex+1)
		}

		if memory.WasCanceled() {
			failedRewrites += max(0, numberOfResults-resultIndex-1)
			break
		}
	}
	if memory.WasCanceledRAM() {
		utils.LogError("Rewrite Canceled: Not enough RAM")
	} else {
		utils.LogInfo("Finished Rewrite")
	}
	utils.LogInfo("Number Results: ", numberOfResults)
	utils.LogInfo("Successfully rewrites: ", numberRewrittenTrace)
	utils.LogInfo("No need/not possible to rewrite: ", notNeededRewrites)
	if failedRewrites > 0 {
		utils.LogInfo("Failed rewrites: ", failedRewrites)
	} else {
		utils.LogInfo("Failed rewrites: ", failedRewrites)
	}

	return nil
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
		"all":                  false, // all cases enabled
		"sendOnClosed":         false,
		"receiveOnClosed":      false,
		"doneBeforeAdd":        false,
		"closeOnClosed":        false,
		"concurrentRecv":       false,
		"leak":                 false,
		"selectWithoutPartner": false,
		"unlockBeforeLock":     false,
		"mixedDeadlock":        false,
		"resourceDeadlock":     false,
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
		case 'p':
			analysisCases["selectWithoutPartner"] = true
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

// Rewrite the trace file based on given analysis results
//
// Parameter:
//   - outMachine string: The path to the analysis result file
//   - newTrace string: The path where the new traces folder will be created
//   - resultIndex int: The index of the result to use for the reordered trace file
//   - numberOfRoutines int: The number of routines in the trace
//   - rewrittenTrace *map[utils.ResultType][]string: set of bugs that have been already rewritten
//
// Returns:
//   - bool: true, if a rewrite was necessary, false if not (e.g. actual bug, warning)
//   - error: An error if the trace file could not be created
func rewriteTrace(outMachine string, newTrace string, resultIndex int,
	numberOfRoutines int, rewrittenTrace *map[utils.ResultType][]string) (bool, error) {
	timer.Start(timer.Rewrite)
	defer timer.Stop(timer.Rewrite)

	actual, bug, err := io.ReadAnalysisResults(outMachine, resultIndex)
	if err != nil {
		return false, err
	}

	if actual {
		return false, nil
	}

	traceCopy := analysis.CopyMainTrace()

	rewriteNeeded, code, err := rewriter.RewriteTrace(&traceCopy, bug, *rewrittenTrace)

	if err != nil {
		return rewriteNeeded, err
	}

	err = io.WriteTrace(&traceCopy, newTrace, true)
	if err != nil {
		return rewriteNeeded, err
	}

	err = io.WriteRewriteInfoFile(newTrace, bug, code, resultIndex)
	if err != nil {
		return rewriteNeeded, err
	}

	return rewriteNeeded, nil
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
	println("Usage: ./analyzer [mode] [options]\n")
	println("There are two different modes of operation:")
	println("1. Analysis")
	println("2. Fuzzing")
	println("\n\n")
	printHelpMode("analysis")
	printHelpMode("fuzzing")
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
		println("Usage: ./analyzer analysis [options]")
		// println("  -main                  Run on the main function instead on tests")
		println("  -path [path]           Path to the folder containing the program and tests, if main, path to the file containing the main function")
		println("  -exec [name]           Name of the test to run (do not set to run all tests)")
		println("  -prog [name]           Name of the program (used for statistics)")
		println("  -timeoutRec [second]   Set a timeout in seconds for the recording")
		println("  -timeoutRep [second]   Set a timeout in seconds for the replay")
		println("  -ignoreAtomics         Set to ignore atomics in replay")
		println("  -recordTime            Set to record runtimes")
		println("  -notExec               Set to determine never executed operations")
		println("  -stats                 Set to create statistics")
		println("  -keepTrace             Do not delete the trace files after analysis finished")
	case "fuzzing":
		println("Mode: fuzzing")
		println("Create runs for fuzzing")
		println("This creates and updates the information required for the fuzzing runs as well as executes the fuzzing runs")
		println("Usage: ./analyzer fuzzing [options]")
		// println("  -main                  Run on the main function instead on tests")
		println("  -path [folder]         If -main, path to the file containing the main function, otherwise path to the program folder")
		println("  -prog [name]           Name of the program")
		println("  -exec [name]           Name of the test to run (do not set to run all tests)")
		println("  -fuzzingMode [mode]    Mode of fuzzing:")
		println("\tGFuzz\n\tGFuzzHB\n\tGFuzzHBFlow\n\tFlow\n\tGoPie\n\tGoPieHB")

		println("Additionally, the tags from mode analysis can be used")
	default:
		println("Mode: unknown")
		printHelp()
	}
}

// checkProgPath checks if the provided path to the program that should
// be run/analyzed exists. If not, it panics.
func checkProgPath() {
	_, err := os.Stat(progPath)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		utils.LogErrorf("Could not find path %s", progPath)
		panic(err)
	}
}

// checkVersion checks the version of the program to be analyzed.
// Advocate is implemented in and for go1.24. It the analyzed program has another
// version, especially if the other version is also installed on the machine,
// this can lead to problems. checkVersion therefore reads the version of the
// analyzed program and if its not 1.24, a warning and information is printed
// to the terminal
func checkVersion() {
	var goModPath string

	if progPath == "" {
		return
	}

	// Search for go.mod
	err := filepath.WalkDir(progPath, func(path string, d os.DirEntry, err error) error {
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
		utils.LogError("Could not find go.mod")
		return
	}

	// Open and read go.mod
	file, err := os.Open(goModPath)
	if err != nil {
		utils.LogError("Could not find go.mod")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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
