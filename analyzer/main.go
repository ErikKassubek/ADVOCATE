// Copyrigth (c) 2024 Erik Kassubek
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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"analyzer/analysis"
	"analyzer/bugs"
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

	fifo                  bool
	ignoreCriticalSection bool
	ignoreAtomics         bool
	ignoreRewrite         string

	rewriteAll bool

	noRewrite  bool
	keepTraces bool

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

const (
	HBFuzzHBAna = 0
	FuzzHBAna   = 1
	HBFuzzNoAna = 2
	FuzzNoAna   = 3
)

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

	flag.BoolVar(&fifo, "fifo", false, "Assume a FIFO ordering for buffered channels (default false)")
	flag.BoolVar(&ignoreCriticalSection, "ignCritSec", false, "Ignore happens before relations of critical sections (default false)")
	flag.BoolVar(&ignoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")
	flag.StringVar(&ignoreRewrite, "ignoreRew", "", "Path to a result machine file. If a found bug is already in this file, it will not be rewritten")

	flag.BoolVar(&rewriteAll, "rewriteAll", false, "If a the same position is flagged multiple times, run the replay for each of them. "+
		"If not set, only the first occurence is rewritten")

	flag.BoolVar(&noRewrite, "noRewrite", false, "Do not rewrite the trace file (default false)")
	flag.BoolVar(&keepTraces, "keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")

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
		"Mode for fuzzing. Possible values are:\n\tGoFuzz\n\tGoFuzzHB\n\tGFuzzFlow\n\tFlow\n\tGoPie")

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
		go memory.MemorySupervisor() // cancel analysis if not enough ram
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
		fifo, ignoreCriticalSection, rewriteAll, ignoreRewrite, onlyAPanicAndLeak,
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

func modeFuzzing() {
	if progName == "" {
		utils.LogError("Provide a name for the analyzed program. Set with -prog [name]")
		return
	}

	checkVersion()

	err := fuzzing.Fuzzing(modeMain, fuzzingMode, pathToAdvocate, progPath, progName, execName,
		ignoreAtomics, recordTime, notExec, statistics,
		keepTraces, cont)
	if err != nil {
		utils.LogError("Fuzzing Failed: ", err.Error())
	}
}

func modeToolchain(mode string, numRerecorded int) {
	checkVersion()
	err := toolchain.Run(mode, pathToAdvocate, progPath, "", execName, progName, execName,
		numRerecorded, -1, "", ignoreAtomics, recordTime, notExec, statistics, keepTraces, true, cont, 0, 0)
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

func getFolderTrace(pathTrace string) (string, error) {
	folderTrace, err := filepath.Abs(pathTrace)
	if err != nil {
		return "", err
	}

	// remove last folder from path
	return folderTrace[:strings.LastIndex(folderTrace, string(os.PathSeparator))+1], nil
}

func modeAnalyzer(pathTrace string, noRewrite bool,
	analysisCases map[string]bool, outReadable string, outMachine string,
	ignoreAtomics bool, fifo bool, ignoreCriticalSection bool,
	rewriteAll bool, newTrace string, ignoreRewrite string,
	fuzzingRun int, onlyAPanicAndLeak bool) error {
	// printHeader()

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
		return fmt.Errorf("Trace does not contain any elem")
	} else {
		utils.LogInfof("Read trace with %d elements in %d routines", numberElems, numberOfRoutines)
	}

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

	if canceled, ram := memory.CheckCanceled(); canceled {
		// analysis.LogSizes()
		analysis.Clear()
		if ram {
			return fmt.Errorf("Analysis was canceled due to insufficient small RAM")
		} else {
			return fmt.Errorf("Analysis was canceled due to unexpected panic")
		}
	} else {
		utils.LogInfo("Analysis finished")
	}

	numberOfResults, err := results.CreateResultFiles(noWarning, true)
	if err != nil {
		utils.LogError("Error in printing summary: ", err.Error())
	}

	// collect the required data to decide whether run is interesting
	// and to create the mutations
	if fuzzingRun >= 0 {
		fuzzing.ParseTrace(analysis.GetTraces())
	}

	if noRewrite {
		utils.LogInfo("Skip rewrite")
		return nil
	}

	numberRewrittenTrace := 0
	failedRewrites := 0
	notNeededRewrites := 0
	utils.LogInfo("Start rewriting")
	originalTrace, err := analysis.CopyCurrentTrace()

	if err != nil {
		utils.LogError("Failed to rewrite: ", err)
		return nil
	}

	if memory.WasCanceled() {
		utils.LogError("Could not run rewrite: Not enough RAM")
	}

	rewrittenBugs := make(map[bugs.ResultType][]string) // bugtype -> paths string

	addAlreadyProcessed(rewrittenBugs, ignoreRewrite)

	file := filepath.Base(pathTrace)
	rewriteNr := "0"
	spl := strings.Split(file, "_")
	if len(spl) > 1 {
		rewriteNr = spl[len(spl)-1]
	}

	for resultIndex := 0; resultIndex < numberOfResults; resultIndex++ {
		needed, double, err := rewriteTrace(outMachine,
			newTrace+"_"+strconv.Itoa(resultIndex+1)+"/", resultIndex, numberOfRoutines, &rewrittenBugs, !rewriteAll)

		if !needed {
			notNeededRewrites++
			if double {
				fmt.Printf("Bugreport info: %s_%d,double\n", rewriteNr, resultIndex+1)
			} else {
				fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
			}
		} else if err != nil {
			failedRewrites++
			fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
		} else { // needed && err == nil
			numberRewrittenTrace++
			fmt.Printf("Bugreport info: %s_%d,suc\n", rewriteNr, resultIndex+1)
		}
		analysis.SetTrace(originalTrace)

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

/*
 * Parse the given analysis cases
 * Args:
 *   cases (string): The string of analysis cases to parse
 * Returns:
 *   map[string]bool: A map of the analysis cases and if they are set
 *   error: An error if the cases could not be parsed
 */
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
		"cyclicDeadlock":       false, // only for comparison with resource deadlock
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

		// remove to run old cyclic deadlock detection
		analysisCases["cyclicDeadlock"] = false

		// takes to long, only take out for tests
		analysisCases["resourceDeadlock"] = false

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
			// enable to run old cyclic deadlock detection
			// analysisCases["cyclicDeadlock"] = true
			analysisCases["resourceDeadlock"] = true
		// case 'm':
		// analysisCases["mixedDeadlock"] = true
		default:
			return nil, fmt.Errorf("Invalid analysis case: %c", c)
		}
	}
	return analysisCases, nil
}

func addAlreadyProcessed(alreadyProcessed map[bugs.ResultType][]string, ignoreRewrite string) {
	if ignoreRewrite == "" {
		return
	}

	data, err := os.ReadFile(ignoreRewrite)
	if err != nil {
		return
	}
	for _, bugStr := range strings.Split(string(data), "\n") {
		_, bug, err := bugs.ProcessBug(bugStr)
		if err != nil {
			continue
		}

		if _, ok := alreadyProcessed[bug.Type]; !ok {
			alreadyProcessed[bug.Type] = make([]string, 0)
		} else {
			if utils.ContainsString(alreadyProcessed[bug.Type], bugStr) {
				continue
			}
		}
		alreadyProcessed[bug.Type] = append(alreadyProcessed[bug.Type], bugStr)
	}
}

/*
 * Rewrite the trace file based on given analysis results
 * Args:
 *   outMachine (string): The path to the analysis result file
 *   newTrace (string): The path where the new traces folder will be created
 *   resultIndex (int): The index of the result to use for the reordered trace file
 *   numberOfRoutines (int): The number of routines in the trace
 *   rewrittenTrace (*map[string][]string): set of bugs that have been already rewritten
 * Returns:
 *   bool: true, if a rewrite was nessesary, false if not (e.g. actual bug, warning)
 *   bool: true if rewrite was skipped because of double
 *   error: An error if the trace file could not be created
 */
func rewriteTrace(outMachine string, newTrace string, resultIndex int,
	numberOfRoutines int, rewrittenTrace *map[bugs.ResultType][]string, rewriteOnce bool) (bool, bool, error) {
	timer.Start(timer.Rewrite)
	defer timer.Stop(timer.Rewrite)

	actual, bug, err := io.ReadAnalysisResults(outMachine, resultIndex)
	if err != nil {
		return false, false, err
	}

	if actual {
		return false, false, nil
	}

	rewriteNeeded, skip, code, err := rewriter.RewriteTrace(bug, *rewrittenTrace, rewriteOnce)

	if err != nil {
		return rewriteNeeded, false, err
	}

	if skip {
		return rewriteNeeded, skip, err
	}

	err = io.WriteTrace(newTrace, numberOfRoutines)
	if err != nil {
		return rewriteNeeded, false, err
	}

	err = io.WriteRewriteInfoFile(newTrace, string(bug.Type), code, resultIndex)
	if err != nil {
		return rewriteNeeded, false, err
	}

	return rewriteNeeded, false, nil
}

func printHeader() {
	fmt.Print("\n")
	fmt.Println(" $$$$$$\\  $$$$$$$\\  $$\\    $$\\  $$$$$$\\   $$$$$$\\   $$$$$$\\ $$$$$$$$\\ $$$$$$$$\\ ")
	fmt.Println("$$  __$$\\ $$  __$$\\ $$ |   $$ |$$  __$$\\ $$  __$$\\ $$  __$$\\\\__$$  __|$$  _____|")
	fmt.Println("$$ /  $$ |$$ |  $$ |$$ |   $$ |$$ /  $$ |$$ /  \\__|$$ /  $$ |  $$ |   $$ |      ")
	fmt.Println("$$$$$$$$ |$$ |  $$ |\\$$\\  $$  |$$ |  $$ |$$ |      $$$$$$$$ |  $$ |   $$$$$\\    ")
	fmt.Println("$$  __$$ |$$ |  $$ | \\$$\\$$  / $$ |  $$ |$$ |      $$  __$$ |  $$ |   $$  __|   ")
	fmt.Println("$$ |  $$ |$$ |  $$ |  \\$$$  /  $$ |  $$ |$$ |  $$\\ $$ |  $$ |  $$ |   $$ |      ")
	fmt.Println("$$ |  $$ |$$$$$$$  |   \\$  /    $$$$$$  |\\$$$$$$  |$$ |  $$ |  $$ |   $$$$$$$$\\ ")
	fmt.Println("\\__|  \\__|\\_______/     \\_/     \\______/  \\______/ \\__|  \\__|  \\__|   \\________|")

	headerInfo := "\n\n\n" +
		"Welcome to the trace analyzer and rewriter.\n" +
		"This program analyzes a trace file and detects common concurrency bugs in Go programs.\n" +
		"It can also create a reordered trace file based on the analysis results.\n" +
		"Be aware, that the analysis is based on the trace file and may not be complete.\n" +
		"Be aware, that the analysis may contain false positives and false negatives.\n" +
		"\n" +
		"If the rewrite of a trace file does not create the expected result, it can help to run the\n" +
		"analyzer with the -ignCritSecflag to ignore the happens before relations of critical sections (mutex lock/unlock operations).\n" +
		"For the first analysis this is not recommended, because it increases the likelihood of false positives." +
		"\n\n\n"

	fmt.Print(headerInfo)
}

func printHelp() {
	println("Usage: ./analyzer [mode] [options]\n")
	println("There are two different modes of operation:")
	println("1. Analysis")
	println("2. Fuzzing")
	println("\n\n")
	printHelpMode("analysis")
	printHelpMode("fuzzing")
}

func printHelpMode(mode string) {
	switch mode {
	case "analysis":
		println("Mode: analysis")
		println("Analyze a test or tool chain")
		println("This runs the analysis on tests or the main function")
		println("Usage: ./analyzer analysis [options]")
		// println("  -main                  Run on the main function instead on tests")
		println("  -path [path]           Path to the folder containing the program and tests, if main, path to the file containing the main function")
		println("  -exec [name]           Name of the test to run (do not set to run all tests)")
		println("  -prog [name]           Name of the program (used for statistics)")
		println("  -timeoutRec [second]   Set a timeout in seconds for the recording")
		println("  -timeoutRepl [second]  Set a timeout in seconds for the replay")
		println("  -ignoreAtomics         Set to ignore atomics in replay")
		println("  -recordTime            Set to record runtimes")
		println("  -notExec               Set to determine never executed operations")
		println("  -stats                 Set to create statistics")
		println("  -keepTrace             Do not delete the trace files after analysis finished")
	case "fuzzing":
		println("Mode: fuzzing")
		println("Create runs for fuzzing")
		println("This creates and updates the information required for the fuzzing runs")
		println("Usage: ./analyzer fuzzing [options]")
		// println("  -main                  Run on the main function instead on tests")
		println("  -path [folder]         If -main, path to the file containing the main function, otherwise path to the program folder")
		println("  -prog [name]           Name of the program")
		println("  -exec [name]           Name of the test to run (do not set to run all tests)")
		println("  -noWarning             Only show critical bugs")
		println("  -fuzzingMode [mode]    Mode of fuzzing:")
		println("     GoFuzz\n\tGoFuzzHB\n\tGFuzzFlow\n\tFlow\n\tGoPie")
		println("     GoFuzzHB")
		println("     GoFuzzFlow")
		println("     Flow")
		println("     GoPie")
		println("Additionally, the tags from mode tool can be used")
	default:
		println("Mode: unknown")
		printHelp()
	}
}

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
