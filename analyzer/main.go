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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"analyzer/analysis"
	"analyzer/bugs"
	"analyzer/complete"
	"analyzer/explanation"
	"analyzer/fuzzing"
	"analyzer/io"
	"analyzer/results"
	"analyzer/rewriter"
	"analyzer/stats"
	timemeasurement "analyzer/timeMeasurement"
	"analyzer/toolchain"
	"analyzer/utils"

	"github.com/shirou/gopsutil/mem"
)

func main() {
	help := flag.Bool("h", false, "Print help")

	pathToAdvocate := flag.String("advocate", "", "Path to advocate")

	pathTrace := flag.String("trace", "", "Path to the trace folder to analyze or rewrite")
	programPath := flag.String("dir", "", "Path to the program folder, for toolMain: path to main file, for toolTest: path to test folder")

	progName := flag.String("prog", "", "Name of the program")
	testName := flag.String("test", "", "Name of the test")
	execName := flag.String("exec", "", "Name of the executable")

	timeoutAnalysis := flag.Int("timeout", -1, "Set a timeout in seconds for the analysis")
	timeoutReplay := flag.Int("timeoutReplay", -1, "Set a timeout in seconds for the replay")
	recordTime := flag.Bool("time", true, "measure the runtime")

	resultFolder := flag.String("out", "", "Path to where the result file should be saved.")
	resultFolderTool := flag.String("resultTool", "", "Path where the advocateResult folder created by the pipeline is located")
	outM := flag.String("outM", "results_machine", "Name for the result machine file")
	outR := flag.String("outR", "results_readable", "Name for the result readable file")
	outT := flag.String("outT", "rewritten_trace", "Name for the rewritten traces")

	fifo := flag.Bool("fifo", false, "Assume a FIFO ordering for buffered channels (default false)")
	ignoreCriticalSection := flag.Bool("ignCritSec", false, "Ignore happens before relations of critical sections (default false)")
	ignoreAtomics := flag.Bool("ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")
	ignoreRewrite := flag.String("ignoreRew", "", "Path to a result machine file. If a found bug is already in this file, it will not be rewritten")

	rewriteAll := flag.Bool("rewriteAll", false, "If a the same position is flagged multiple times, run the replay for each of them. "+
		"If not set, only the first occurence is rewritten")

	noRewrite := flag.Bool("noRewrite", false, "Do not rewrite the trace file (default false)")
	noWarning := flag.Bool("noWarning", false, "Do not print warnings (default false)")
	noPrint := flag.Bool("noPrint", false, "Do not print the results to the terminal (default false). Automatically set -noRewrite to true")
	keepTraces := flag.Bool("keepTrace", false, "If set, the traces are not deleted after analysis. Can result in very large output folders")

	notExec := flag.Bool("notExec", false, "Find never executed operations, *notExec, *stats")
	statistics := flag.Bool("stats", false, "Create statistics")

	scenarios := flag.String("scen", "", "Select which analysis scenario to run, e.g. -scen srd for the option s, r and d."+
		"If not set, all scenarios are run.\n"+
		"Options:\n"+
		"\ts: Send on closed channel\n"+
		"\tr: Receive on closed channel\n"+
		"\tw: Done before add on waitGroup\n"+
		"\tn: Close of closed channel\n"+
		"\tb: Concurrent receive on channel\n"+
		"\tl: Leaking routine\n"+
		"\tp: Select case without partner\n"+
		"\tu: Unlock of unlocked mutex\n",
	)
	// "\tc: Cyclic deadlock\n",
	// "\tm: Mixed deadlock\n"

	go memorySupervisor() // panic if not enough ram

	flag.Parse()

	var mode string
	if len(os.Args) >= 2 {
		mode = os.Args[1]
		flag.CommandLine.Parse(os.Args[2:])
		if *help {
			printHelpMode(mode)
		}
	} else {
		if *help {
			printHelp()
			return
		}
		fmt.Println("No mode selected")
		fmt.Println("Select one mode from 'run', 'stats', 'explain' or 'check'")
		printHelp()
		return
	}

	if *resultFolder == "" {
		*resultFolder = getFolderTrace(*pathTrace)
		if (*resultFolder)[len(*resultFolder)-1] != os.PathSeparator {
			*resultFolder += string(os.PathSeparator)
		}
	}

	outMachine := filepath.Join(*resultFolder, *outM) + ".log"
	outReadable := filepath.Join(*resultFolder, *outR) + ".log"
	newTrace := filepath.Join(*resultFolder, *outT)
	if *ignoreRewrite != "" {
		*ignoreRewrite = filepath.Join(*resultFolder, *ignoreRewrite)
	}

	// function injection to prevent circle import
	toolchain.InitFuncAnalyzer(modeAnalyzer)

	switch mode {
	case "toolMain", "toolmain":
		modeToolchain("main", *pathToAdvocate, *programPath, *execName, *progName, *testName, *timeoutAnalysis, *timeoutReplay, 0, *ignoreAtomics, *recordTime, *notExec, *statistics, *keepTraces)
	case "toolTest", "tooltest":
		modeToolchain("test", *pathToAdvocate, *programPath, "", *progName, *testName, *timeoutAnalysis, *timeoutReplay, 0, *ignoreAtomics, *recordTime, *notExec, *statistics, *keepTraces)
	case "stats":
		modeStats(*pathTrace, *progName, *testName)
	case "explain":
		modeExplain(*pathTrace, !*rewriteAll)
	case "check":
		modeCheck(resultFolderTool, programPath)
	case "run":
		modeAnalyzer(*pathTrace, *noPrint, *noRewrite, *scenarios, outReadable,
			outMachine, *ignoreAtomics, *fifo, *ignoreCriticalSection,
			*noWarning, *rewriteAll, newTrace, *timeoutAnalysis, *ignoreRewrite,
			-1)
	case "fuzzing":
		modeFuzzing(*pathToAdvocate, *programPath, *progName, *testName)
	default:
		fmt.Printf("Unknown mode %s\n", os.Args[1])
		fmt.Println("Select one mode from 'run', 'stats', 'explain' or 'check'")
		printHelp()
	}
}

func modeFuzzing(advocate, testPath, progName, testName string) {
	if progName == "" {
		fmt.Println("Provide a name for the analyzed program. Set with -prog [name]")
		return
	}

	if testName == "" {
		fmt.Println("Provide a name for the analyzed test. Set with -test [name]")
		return
	}

	err := fuzzing.Fuzzing(advocate, testPath, progName, testName)
	if err != nil {
		fmt.Println("Fuzzing Failed: ", err.Error())
	}
}

func modeToolchain(mode, advocate, file, execName, progName, test string,
	timeoutA, timeoutR, numRerecorded int,
	replayAt, meaTime, notExec, stats, keepTraces bool) {
	err := toolchain.Run(mode, advocate, file, execName, progName, test,
		timeoutA, timeoutR, numRerecorded,
		-1, replayAt, meaTime, notExec, stats, keepTraces)
	if err != nil {
		fmt.Println("Failed to run toolchain")
		fmt.Println(err.Error())
	}
}

func modeStats(pathFolder string, progName string, testName string) {
	// instead of the normal program, create statistics for the trace
	if pathFolder == "" {
		fmt.Println("Provide the path to the folder containing the results_machine file. Set with -trace [path]")
		return
	}

	if progName == "" {
		fmt.Println("Provide a name for the analyzed program. Set with -prog [name]")
		return
	}

	if testName == "" {
		testName = progName
	}

	stats.CreateStats(pathFolder, progName, testName)
}

func modeCheck(resultFolderTool, programPath *string) {
	if *resultFolderTool == "" {
		fmt.Println("Please provide the path to the advocateResult folder created by the pipeline. Set with -resultTool [folder]")
		return
	}

	if *programPath == "" {
		fmt.Println("Please provide the path to the program folder. Set with -dir [folder]")
		return
	}

	err := complete.Check(*resultFolderTool, *programPath)

	if err != nil {
		panic(err.Error())
	}
}

func modeExplain(pathTrace string, ignoreDouble bool) {
	if pathTrace == "" {
		fmt.Println("Please provide a path to the trace files for the explanation. Set with -trace [file]")
		return
	}

	err := explanation.CreateOverview(pathTrace, ignoreDouble)
	if err != nil {
		log.Println("Error creating explanation: ", err.Error())
	}
}

func getFolderTrace(pathTrace string) string {
	folderTrace, err := filepath.Abs(pathTrace)
	if err != nil {
		panic(err)
	}

	// remove last folder from path
	return folderTrace[:strings.LastIndex(folderTrace, string(os.PathSeparator))+1]
}

func modeAnalyzer(pathTrace string, noPrint bool, noRewrite bool,
	scenarios string, outReadable string, outMachine string,
	ignoreAtomics bool, fifo bool, ignoreCriticalSection bool,
	noWarning bool, rewriteAll bool, newTrace string, timeout int, ignoreRewrite string,
	fuzzingRun int) {
	// printHeader()

	if pathTrace == "" {
		fmt.Println("Please provide a path to the trace files. Set with -trace [folder]")
		return
	}

	if noPrint {
		noRewrite = true
	}

	// set timeout
	if timeout > 0 {
		go func() {
			<-time.After(time.Duration(timeout) * time.Second)
			os.Exit(1)
		}()
	}

	analysisCases, err := parseAnalysisCases(scenarios)
	if err != nil {
		panic(err)
	}

	// clean data in case of fuzzing
	if analysis.DataUsed {
		analysis.ClearData()
		analysis.ClearTrace()
		analysis.DataUsed = true
	}

	// run the analysis and, if requested, create a reordered trace file
	// based on the analysis results

	results.InitResults(outReadable, outMachine)

	// done and separate routine to implement timeout
	done := make(chan bool)
	numberOfRoutines := 0
	containsElems := false
	go func() {
		defer func() { done <- true }()

		numberOfRoutines, containsElems, err = io.CreateTraceFromFiles(pathTrace, ignoreAtomics)
		if err != nil {
			panic(err)
		}

		log.Println("Trace size: ", len(analysis.GetTraces()))

		if !containsElems {
			fmt.Println("Trace does not contain any elem")
			fmt.Println("Skip analysis")
			return
		}

		analysis.SetNumberOfRoutines(numberOfRoutines)

		if analysisCases["all"] {
			fmt.Println("Start Analysis for all scenarios")
		} else {
			fmt.Println("Start Analysis for the following scenarios:")
			for key, value := range analysisCases {
				if value {
					fmt.Println("\t", key)
				}
			}
		}

		timemeasurement.Start("analysis")
		analysis.RunAnalysis(fifo, ignoreCriticalSection, analysisCases, fuzzingRun >= 0)
		timemeasurement.End("analysis")

		timemeasurement.Print()
	}()

	if timeout > 0 {
		select {
		case <-done:
			fmt.Print("Analysis finished\n\n")
		case <-time.After(time.Duration(timeout) * time.Second):
			fmt.Printf("Analysis ended by timeout after %d seconds\n\n", timeout)
		}
	} else {
		<-done
	}

	numberOfResults := results.PrintSummary(noWarning, noPrint)

	// collect the required data to decide whether run is interesting
	// and to create the mutations
	if fuzzingRun >= 0 {
		fuzzing.ParseTrace(analysis.GetTraces())
	}

	if !noRewrite {
		numberRewrittenTrace := 0
		failedRewrites := 0
		notNeededRewrites := 0
		println("\n\nStart rewriting trace file ", pathTrace)
		originalTrace := analysis.CopyCurrentTrace()

		analysis.ClearData()

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
				println("Trace can not be rewritten.")
				notNeededRewrites++
				if double {
					fmt.Printf("Bugreport info: %s_%d,double", rewriteNr, resultIndex+1)
				} else {
					fmt.Printf("Bugreport info: %s_%d,fail", rewriteNr, resultIndex+1)
				}
			} else if err != nil {
				println("Failed to rewrite trace: ", err.Error())
				failedRewrites++
				analysis.SetTrace(originalTrace)
				fmt.Printf("Bugreport info: %s_%d,fail", rewriteNr, resultIndex+1)
			} else { // needed && err == nil
				numberRewrittenTrace++
				analysis.SetTrace(originalTrace)
				fmt.Printf("Bugreport info: %s_%d,suc", rewriteNr, resultIndex+1)
			}

			print("\n\n")
		}

		println("Finished Rewrite")
		println("\n\n\tNumber Results: ", numberOfResults)
		println("\tSuccessfully rewrites: ", numberRewrittenTrace)
		println("\tNo need/not possible to rewrite: ", notNeededRewrites)
		if failedRewrites > 0 {
			println("\tFailed rewrites: ", failedRewrites)
		} else {
			println("\tFailed rewrites: ", failedRewrites)
		}
	}

	print("\n\n\n")
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
		"cyclicDeadlock":       false,
		"mixedDeadlock":        false,
	}

	if cases == "" {
		analysisCases["all"] = true
		analysisCases["sendOnClosed"] = true
		analysisCases["receiveOnClosed"] = true
		analysisCases["doneBeforeAdd"] = true
		analysisCases["closeOnClosed"] = true
		analysisCases["concurrentRecv"] = true
		analysisCases["leak"] = true
		analysisCases["selectWithoutPartner"] = true
		analysisCases["unlockBeforeLock"] = true
		// analysisCases["cyclicDeadlock"] = true
		// analysisCases["mixedDeadlock"] = true

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
		// case 'c':
		// 	analysisCases["cyclicDeadlock"] = true
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

	actual, bug, err := io.ReadAnalysisResults(outMachine, resultIndex)
	if err != nil {
		return false, false, err
	}

	if actual {
		return false, false, nil
	}

	rewriteNeeded, skip, code, err := rewriter.RewriteTrace(bug, 0, *rewrittenTrace, rewriteOnce)

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

func memorySupervisor() {
	thresholdRAM := uint64(1 * 1024 * 1024 * 1024) // 1GB
	thresholdSwap := uint64(200 * 1024 * 1024)     // 200mb
	for {
		// Get the memory stats
		v, err := mem.VirtualMemory()
		if err != nil {
			log.Fatalf("Error getting memory info: %v", err)
		}

		// Get the swap stats
		s, err := mem.SwapMemory()
		if err != nil {
			log.Fatalf("Error getting swap info: %v", err)
		}

		// fmt.Printf("Available RAM: %v MB, Available Swap: %v MB\n", v.Available/1024/1024, s.Free/1024/1024)

		// Panic if available RAM or swap is below the threshold
		if v.Available < thresholdRAM {
			log.Panicf("Available RAM is below threshold! Available: %v MB, Threshold: %v MB", v.Available/1024/1024, thresholdRAM/1024/1024)
		}

		if s.Free < thresholdSwap {
			log.Panicf("Available Swap is below threshold! Available: %v MB, Threshold: %v MB", s.Free/1024/1024, thresholdSwap/1024/1024)
		}

		// Sleep for a while before checking again
		time.Sleep(5 * time.Second)
	}
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
	println("There are different modes of operation:")
	println("1. Analyze a trace file and create a reordered trace file based on the analysis results (Default)")
	println("2. Create an explanation for a found bug")
	println("3. Check if all concurrency elements of the program have been executed at least once")
	println("4. Create statistics about a program")
	println("5. Run the toolchain on tests")
	println("6. Run the toolchain on a main function")
	println("7. Create new runs for fuzzing\n\n")
	println("1. Analyze a trace file and create a reordered trace file based on the analysis results (Default)")
	println("This mode is the default mode and analyzes a trace file and creates a reordered trace file based on the analysis results.")
	println("Usage: ./analyzer run [options]")
	println("It has the following options:")
	println("  -trace [file]          Path to the trace folder to analyze or rewrite (required)")
	println("  -fifo                  Assume a FIFO ordering for buffered channels (default false)")
	println("  -ignCritSec            Ignore happens before relations of critical sections (default false)")
	println("  -noRewrite             Do not rewrite the trace file (default false)")
	println("  -noWarning             Do not print warnings (default false)")
	println("  -noPrint               Do not print the results to the terminal (default false). Automatically set -noRewrite to true")
	println("  -keepTrace             Do not delete the trace files after analysis finished")
	println("  -out [folder]          Path to where the result file should be saved. (default parallel to -t)")
	println("  -ignoreAtomics         Ignore atomic operations (default false). Use to reduce memory header for large traces.")
	println("  -rewriteAll            If the same bug is detected multiple times, run the replay for each of them. If not set, only the first occurence is rewritten")
	println("  -timeout [second]      Set a timeout in seconds for the analysis")
	println("  -scen [cases]          Select which analysis scenario to run, e.g. -scen srd for the option s, r and d.")
	println("                         If it is not set, all scenarios are run")
	println("                         Options:")
	println("                             s: Send on closed channel")
	println("                             r: Receive on closed channel")
	println("                             w: Done before add on waitGroup")
	println("                             n: Close of closed channel")
	println("                             b: Concurrent receive on channel")
	println("                             l: Leaking routine")
	println("                             u: Select case without partner")
	// println("                             c: Cyclic deadlock")
	// println("                             m: Mixed deadlock")
	println("\n\n")
	println("2. Create an explanation for a found bug")
	println("Usage: ./analyzer explain [options]")
	println("This mode creates an explanation for a found bug in the trace file.")
	println("It has the following options:")
	println("  -trace [file]          Path to the folder containing the machine readable result file (required)")
	println("\n\n")
	println("3. Check if all concurrency elements of the program have been executed at least once")
	println("Usage: ./analyzer check [options]")
	println("This mode checks if all concurrency elements of the program have been executed at least once.")
	println("It has the following options:")
	println("  -resultTool [folder]   Path where the advocateResult folder created by the pipeline is located (required)")
	println("  -dir [folder]          Path to the program folder (required)")
	println("\n\n")
	println("4. Create statistics about a program")
	println("This creates some statistics about the program and the trace")
	println("Usage: ./analyzer stats [options]")
	// println("  -dir [folder] Path to the program folder (required)")
	println("  -trace [file]          Path to the folder containing the results_machine file (required)")
	println("  -prog [name]           Name of the program")
	println("  -test [name]           Name of the test")
	println("\n\n")
	println("5. Run the toolchain on tests")
	println("This runs the toolchain on a given main function")
	println("Usage: ./analyzer toolTest [options]")
	println("  -advocate [path]       Path to advocate")
	println("  -dir [path]            Path to the folder containing the program and tests")
	println("  -test [name]           Name of the test to run. If not set, all tests are run")
	println("  -prog [name]           Name of the program (used for statistics)")
	println("  -timeout [sec]         Timeout for the analysis")
	println("  -timeoutRelay [sec]    Timeout for the replay")
	println("  -ignoreAtomics         Set to ignore atomics in replay")
	println("  -recordTime            Set to record runtimes")
	println("  -notExec               Set to determine never executed operations")
	println("  -stats                 Set to create statistics")
	println("  -keepTrace             Do not delete the trace files after analysis finished")
	println("\n\n")
	println("6. Run the toolchain on a main function")
	println("This runs the toolchain on a given main function")
	println("Usage: ./analyzer toolMain [options]")
	println("  -advocate [path]       Path to advocate")
	println("  -dir [path]            Path to the file containing the main function")
	println("  -exec [name]           Name of the executable")
	println("  -prog [name]           Name of the program (used for statistics)")
	println("  -timeout [sec]         Timeout for the analysis")
	println("  -timeoutRelay [sec]    Timeout for the replay")
	println("  -ignoreAtomics         Set to ignore atomics in replay")
	println("  -recordTime            Set to record runtimes")
	println("  -notExec               Set to determine never executed operations")
	println("  -stats                 Set to create statistics")
	println("  -keepTrace             Do not delete the trace files after analysis finished")
	println("\n\n")
	println("7. Create runs for fuzzing")
	println("This creates and updates the information required for the fuzzing runs")
	println("Usage: ./analyzer fuzzing [options]")
	println("  -advocate [path]       Path to advocate")
	println("  -dir [folder]          Path to the program folder")
	println("  -prog [name]           Name of the program")
	println("  -test [name]           Name of the test (only if used on tests)")
}

func printHelpMode(mode string) {
	switch mode {
	case "run":
		println("Mode: run")
		println("Analyze a trace file and create a reordered trace file based on the analysis results (Default)")
		println("This mode is the default mode and analyzes a trace file and creates a reordered trace file based on the analysis results.")
		println("Usage: ./analyzer run [options]")
		println("It has the following options:")
		println("  -trace [file]          Path to the trace folder to analyze or rewrite (required)")
		println("  -fifo                  Assume a FIFO ordering for buffered channels (default false)")
		println("  -ignCritSec            Ignore happens before relations of critical sections (default false)")
		println("  -noRewrite             Do not rewrite the trace file (default false)")
		println("  -noWarning             Do not print warnings (default false)")
		println("  -noPrint               Do not print the results to the terminal (default false). Automatically set -noRewrite to true")
		println("  -keepTrace             Do not delete the trace files after analysis finished")
		println("  -out [folder]          Path to where the result file should be saved. (default parallel to -t)")
		println("  -ignoreAtomics         Ignore atomic operations (default false). Use to reduce memory header for large traces.")
		println("  -rewriteAll            If the same bug is detected multiple times, run the replay for each of them. If not set, only the first occurence is rewritten")
		println("  -timeout [second]      Set a timeout in seconds for the analysis")
		println("  -scen [cases]          Select which analysis scenario to run, e.g. -scen srd for the option s, r and d.")
		println("                         If it is not set, all scenarios are run")
		println("                         Options:")
		println("                             s: Send on closed channel")
		println("                             r: Receive on closed channel")
		println("                             w: Done before add on waitGroup")
		println("                             n: Close of closed channel")
		println("                             b: Concurrent receive on channel")
		println("                             l: Leaking routine")
		println("                             u: Select case without partner")
		// println("                             c: Cyclic deadlock")
		// println("                             m: Mixed deadlock")
	case "explain":
		println("Mode: explain")
		println("Create an explanation for a found bug")
		println("Usage: ./analyzer explain [options]")
		println("This mode creates an explanation for a found bug in the trace file.")
		println("It has the following options:")
		println("  -trace [file]          Path to the folder containing the machine readable result file (required)")
	case "check":
		println("Mode: check")
		println("Check if all concurrency elements of the program have been executed at least once")
		println("Usage: ./analyzer check [options]")
		println("This mode checks if all concurrency elements of the program have been executed at least once.")
		println("It has the following options:")
		println("  -resultTool [folder]   Path where the advocateResult folder created by the pipeline is located (required)")
		println("  -dir [folder]          Path to the program folder (required)")
	case "stats":
		println("Mode: stats")
		println("Create statistics about a program")
		println("This creates some statistics about the program and the trace")
		println("Usage: ./analyzer stats [options]")
		// println("  -dir [folder] Path to the program folder (required)")
		println("  -trace [file]          Path to the folder containing the results_machine file (required)")
		println("  -prog [name]           Name of the program")
		println("  -test [name]           Name of the test")
	case "toolTest", "tooltest":
		println("Mode: toolTest")
		println("Run the toolchain on tests")
		println("This runs the toolchain on a given main function")
		println("Usage: ./analyzer toolTest [options]")
		println("  -advocate [path]       Path to advocate")
		println("  -dir [path]            Path to the folder containing the program and tests")
		println("  -test [name]           Name of the test to run. If not set, all tests are run")
		println("  -prog [name]           Name of the program (used for statistics)")
		println("  -timeout [sec]         Timeout for the analysis")
		println("  -timeoutRelay [sec]    Timeout for the replay")
		println("  -ignoreAtomics         Set to ignore atomics in replay")
		println("  -recordTime            Set to record runtimes")
		println("  -notExec               Set to determine never executed operations")
		println("  -stats                 Set to create statistics")
		println("  -keepTrace             Do not delete the trace files after analysis finished")
	case "toolMain", "toolmain":
		println("Mode: toolMain")
		println("Run the toolchain on a main function")
		println("This runs the toolchain on a given main function")
		println("Usage: ./analyzer toolMain [options]")
		println("  -advocate [path]       Path to advocate")
		println("  -dir [path]            Path to the file containing the main function")
		println("  -exec [name]           Name of the executable")
		println("  -prog [name]           Name of the program (used for statistics)")
		println("  -timeout [sec]         Timeout for the analysis")
		println("  -timeoutRelay [sec]    Timeout for the replay")
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
		println("  -advocate [path]       Path to advocate")
		println("  -dir [folder]          Path to the program folder")
		println("  -prog [name]           Name of the program")
		println("  -test [name]           Name of the test (only if used on tests)")
	default:
		println("Mode: unknown")
		printHelp()
	}
}
