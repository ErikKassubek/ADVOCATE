// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: runFullWorkflowMain.go
// Brief: Function to run the whole ADVOCATE workflow, including running,
//    analysis and replay on all unit tests of a program
//
// Author: Erik Kassubek, Mario Occhinegro
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"analyzer/analysis"
	"analyzer/complete"
	"analyzer/stats"
	"analyzer/utils"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

/*
 * Run ADVOCATE for all given unit tests
 * Args:
 *    pathToAdvocate (string): pathToAdvocate
 *    dir (string): path to the folder containing the unit tests
 *    progName (string): name of the analyzed program
 *    measureTime (bool): if true, measure the time for all steps. This
 *      also runs the tests once without any recoding/replay to get a base value
 *    notExecuted (bool): if true, check for never executed operations
 *    createStats (bool): create a stats file
 *    fuzzing (int): -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
 *    keepTraces (bool): do not delete traces after analysis
 * 	firstRun (bool): this is the first run, only set to false for fuzzing (except for the first fuzzing)
 * Returns:
 *    error
 */
func runWorkflowUnit(pathToAdvocate, dir, progName string,
	measureTime, notExecuted, createStats bool, fuzzing int, keepTraces, firstRun bool) error {
	// Validate required inputs
	if pathToAdvocate == "" {
		return errors.New("Path to advocate is empty")
	}
	if dir == "" {
		return errors.New("Directory is empty")
	}

	// Change to the directory
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("Failed to change directory: %v", dir)
	}

	if firstRun {
		os.RemoveAll("advocateResult")
		if err := os.MkdirAll("advocateResult", os.ModePerm); err != nil {
			return fmt.Errorf("Failed to create advocateResult directory: %v", err)
		}

		// Remove possibly leftover traces from unexpected aborts that could interfere with replay
		removeTraces(dir)
	}

	// Find all _test.go files in the directory
	testFiles, err := FindTestFiles(dir)
	if err != nil {
		return fmt.Errorf("Failed to find test files: %v", err)
	}

	totalFiles := len(testFiles)
	attemptedTests, skippedTests, currentFile := 0, 0, 1

	resultPath := filepath.Join(dir, "advocateResult")

	ranTest := false
	// Process each test file
	for _, file := range testFiles {
		if testName == "" {
			fmt.Println("\n\n=================================================")
			fmt.Printf("Progress %s: %d/%d", progName, currentFile, totalFiles)
			fmt.Printf("\nProcessing file: %s\n", file)
			fmt.Println("=================================================\n\n")
		}

		packagePath := filepath.Dir(file)
		testFunctions, err := FindTestFunctions(file)
		if err != nil {
			utils.LogInfof("Could not find test functions in %s: %v\n", file, err)
			continue
		}

		for _, testFunc := range testFunctions {
			analysis.ClearTrace()
			analysis.ClearData()

			if testName != "" && testName != testFunc {
				continue
			}
			ranTest = true

			attemptedTests++
			packageName := filepath.Base(packagePath)
			fileName := filepath.Base(file)
			utils.LogInfof("Running full workflow for test: %s in package: %s in file: %s\n", testFunc, packageName, file)

			adjustedPackagePath := strings.TrimPrefix(packagePath, dir)
			if !strings.HasSuffix(adjustedPackagePath, string(filepath.Separator)) {
				adjustedPackagePath = adjustedPackagePath + string(filepath.Separator)
			}
			fileNameWithoutEnding := strings.TrimSuffix(fileName, ".go")
			directoryName := filepath.Join("advocateResult", fmt.Sprintf("file(%d)-test(%d)-%s-%s", currentFile, attemptedTests, fileNameWithoutEnding, testFunc))
			directoryPath := filepath.Join(dir, directoryName)
			if fuzzing < 1 {
				utils.LogInfo("Create ", directoryName)
				if err := os.MkdirAll(directoryName, os.ModePerm); err != nil {
					utils.LogErrorf("Failed to create directory %s: %v\n", directoryName, err)
					continue
				}
			}

			// Execute full workflow
			times, nrReplay, nrAnalyzer, err := unitTestFullWorkflow(pathToAdvocate, dir, testFunc, adjustedPackagePath, file, fuzzing)

			if measureTime {
				updateTimeFiles(progName, testFunc, resultPath, times, nrReplay, nrAnalyzer)
			}

			// Move logs and results to the appropriate directory
			total := fuzzing != -1
			collect(dir, packagePath, directoryPath, total)

			if err != nil {
				fmt.Printf("File %d with Test %d failed, check output.log for more information.\n", currentFile, attemptedTests)
				skippedTests++
			}

			generateBugReports(directoryPath, fuzzing)

			if createStats {
				// create statistics
				err := stats.CreateStats(directoryPath, progName, testFunc, fuzzing)
				if err != nil {
					utils.LogError("Could not create statistics: ", err.Error())
				}
			}

			if !keepTraces {
				removeTraces(dir)
				removeTraces(packagePath)
			}

			if total {
				removeLogs(resultPath)
			}
		}

		currentFile++
	}

	if testName != "" && !ranTest {
		return fmt.Errorf("could not find test function %s", testName)
	}

	// Check for untriggered selects
	if notExecuted && testName != "" {
		err := complete.Check(filepath.Join(dir, "advocateResult"), dir)
		if err != nil {
			fmt.Println("Could not run check for untriggered select and not executed progs")
		}
	}

	// Output test summary
	if testName == "" {
		utils.LogInfo("Finished full workflow for all tests")
		utils.LogInfof("Attempted tests: %d\n", attemptedTests)
		utils.LogInfof("Skipped tests: %d\n", skippedTests)
	} else {
		utils.LogInfof("Finished full work flow for %s\n", testName)
	}

	return nil
}

/*
 * Function to write the time information to a file
 * Args:
 *     progName (string): name of the program
 *     folderName (string): path to the destination of the file,
 *     time (map[string]time.Durations): runtimes
 *     numberReplay (int): number of replay
 */
func updateTimeFiles(progName string, testName string, folderName string, times map[string]time.Duration, numberReplay, numberAnalyzer int) {
	timeFilePath := filepath.Join(folderName, "times_"+progName+".csv")

	newFile := false
	_, err := os.Stat(timeFilePath)
	if os.IsNotExist(err) {
		newFile = true
	}

	file, err := os.OpenFile(timeFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.LogError("Error opening or creating file:", err)
		return
	}
	defer file.Close()

	if newFile {
		csvTitels := "TestName,ExecTime,ExecTimeWithTracing,AnalyzerTime,AnalysisTime,HBAnalysisTime,TimeToIdentifyLeaksPlusFindingPoentialPartners,TimeToIdentifyPanicBugs,ReplayTime,NumberReplay\n"
		if _, err := file.WriteString(csvTitels); err != nil {
			utils.LogError("Could not write time: ", err)
		}
	}

	timeInfo := fmt.Sprintf(
		"%s,%.5f,%.5f,%.5f,%.5f,%.5f,%.5f,%.5f,%.5f,%d\n", testName,
		times["run"].Seconds(), times["record"].Seconds(),
		times["analyzer"].Seconds(), times["analysis"].Seconds(),
		times["hb"].Seconds(), times["leak"].Seconds(), times["panic"].Seconds(), times["replay"].Seconds(),
		numberAnalyzer)

	// Write to the file
	if _, err := file.WriteString(timeInfo); err != nil {
		utils.LogError("Could not write time: ", err)
	}
}

/*
 * Function to find all _test.go files in the specified directory
 * Args:
 *    dir (string): folder to search in
 * Returns:
 *    []string: found files
 *    error
 */
func FindTestFiles(dir string) ([]string, error) {
	var testFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			testFiles = append(testFiles, path)
		}
		return nil
	})
	return testFiles, err
}

/*
 * Function to find all test function in the specified file
 * Args:
 *    file (string): file to search in
 * Returns:
 *    []string: functions
 *    error
 */
func FindTestFunctions(file string) ([]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var testFunctions []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "func Test") && strings.Contains(line, "*testing.T") {
			testFunc := strings.TrimSpace(strings.Split(line, "(")[0])
			testFunc = strings.TrimPrefix(testFunc, "func ")
			testFunctions = append(testFunctions, testFunc)
		}
	}
	return testFunctions, nil
}

/*
 * Run the full workflow for a given unit test
 * Args:
 *    pathToAdvocate (string): path to advocate
 *    dir (string): path to the package to test
 *    progName (string): name of the program
 *    testName (string): name of the test
 *    pkg (string): adjusted package path
 *    file (string): file with the test
 *    fuzzing (int): -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
 * Returns:
 *    map[string]time.Duration
 *    int: number of run replays
 *    int: number of analyzer runs
 *    error
 */
func unitTestFullWorkflow(pathToAdvocate, dir, testName, pkg, file string, fuzzing int) (map[string]time.Duration, int, int, error) {

	resTimes := make(map[string]time.Duration)

	output := "output.log"

	outFile, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return resTimes, 0, 0, fmt.Errorf("Failed to open log file: %v", err)
	}
	defer outFile.Close()

	// Redirect stdout and stderr to the file
	origStdout := os.Stdout
	origStderr := os.Stderr

	os.Stdout = outFile
	os.Stderr = outFile

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Validate required inputs
	if pathToAdvocate == "" {
		return resTimes, 0, 0, errors.New("Path to advocate is empty")
	}
	if dir == "" {
		return resTimes, 0, 0, errors.New("Directory is empty")
	}
	if testName == "" {
		return resTimes, 0, 0, errors.New("Test name is empty")
	}
	// if pkg == "" {
	// 	return 0, 0, 0, 0, errors.New("Package is empty")
	// }
	if file == "" {
		return resTimes, 0, 0, errors.New("Test file is empty")
	}

	pathToPatchedGoRuntime := filepath.Join(pathToAdvocate, "go-patch/bin/go")
	pathToGoRoot := filepath.Join(pathToAdvocate, "go-patch")
	pathToAnalyzer := filepath.Join(pathToAdvocate, "analyzer/analyzer")

	if runtime.GOOS == "windows" {
		pathToPatchedGoRuntime += ".exe"
	}

	// Change to the directory
	if err := os.Chdir(dir); err != nil {
		return resTimes, 0, 0, fmt.Errorf("Failed to change directory: %v", err)
	}

	pkg = strings.TrimPrefix(pkg, dir)

	unitTestRun(pkg, file, testName, resTimes)

	err = unitTestRecord(pathToGoRoot, pathToPatchedGoRuntime, pkg, file, testName, fuzzing, resTimes)
	if err != nil {
		utils.LogError("Failed record: ", err.Error())
		return resTimes, 0, 0, err
	}

	unitTestAnalyzer(pathToAnalyzer, dir, pkg, "advocateTrace", output, resTimes, "-1", fuzzing)

	lenRewTraces := unitTestReplay(pathToGoRoot, pathToPatchedGoRuntime, dir, pkg, file, testName, resTimes, false)

	return resTimes, lenRewTraces, 0, nil
}

func unitTestRun(pkg, file, testName string, resTimes map[string]time.Duration) {
	// run the tests without recording/replay
	resTimes["run"] = time.Duration(0)
	if measureTime {
		// Remove header just in case
		if err := headerRemoverUnit(file); err != nil {
			utils.LogError("Failed to remove header: ", err)
		}

		os.Unsetenv("GOROOT")

		timeStart := time.Now()
		utils.LogInfo("Run T0")
		packagePath := utils.MakePathLocal(pkg)
		var err error
		if timeoutRecording != -1 {
			timeoutRecString := fmt.Sprintf("%ds", timeoutRecording)
			err = runCommand("go", "test", "-v", "-timeout", timeoutRecString, "-count=1", "-run="+testName, packagePath)
		} else {
			err = runCommand("go", "test", "-v", "-count=1", "-run="+testName, packagePath)
		}
		if err != nil {
			utils.LogError("Test failed: ", err)
		}
		resTimes["run"] = time.Since(timeStart)
	}
}

func unitTestRecord(pathToGoRoot, pathToPatchedGoRuntime, pkg, file, testName string, fuzzing int, resTimes map[string]time.Duration) error {
	isFuzzing := (fuzzing > 0)

	// Remove header just in case
	if err := headerRemoverUnit(file); err != nil {
		fmt.Printf("Error in removing header: %v\n", err)
		return fmt.Errorf("Failed to remove header: %v", err)
	}

	// Add header
	if err := headerInserterUnit(file, testName, false, fuzzing, "0", timeoutReplay, false); err != nil {
		utils.LogErrorf("Error in adding header: %v\n", err)
		return fmt.Errorf("Error in adding header: %v", err)
	}

	// Run the test
	utils.LogInfo("Run Recording")

	// Set GOROOT
	os.Setenv("GOROOT", pathToGoRoot)

	timeoutRecString := fmt.Sprintf("%ds", timeoutRecording)

	timeStart := time.Now()
	pkgPath := utils.MakePathLocal(pkg)
	err := runCommand(pathToPatchedGoRuntime, "test", "-v", "-timeout", timeoutRecString, "-count=1", "-run="+testName, pkgPath)
	if err != nil {
		if isFuzzing {
			utils.LogError("Failed to run recording: ", err)
		} else {
			utils.LogError("Failed to run fuzzing recording: ", err)
		}
	}
	resTimes["record"] = time.Since(timeStart)

	os.Unsetenv("GOROOT")

	// Remove header after the test
	headerRemoverUnit(file)

	return nil
}

func unitTestAnalyzer(pathToAnalyzer, dir, pkg, traceName, output string,
	resTimes map[string]time.Duration, resultID string, fuzzing int) {
	// Apply analyzer

	tracePath := filepath.Join(dir, pkg, traceName)

	utils.LogInfof("Run the analyzer for %s", tracePath)

	startTime := time.Now()
	var err error
	if resultID == "-1" {
		runAnalyzer(tracePath, noRewriteFlag, analyisCasesFlag, "results_readable.log",
			"results_machine.log", ignoreAtomicsFlag, fifoFlag, ignoreCriticalSectionFlag, rewriteAllFlag, "rewritten_trace",
			timeoutAnalysis, ignoreRewriteFlag, fuzzing, onlyAPanicAndLeakFlag)
	} else {
		outM := fmt.Sprintf("results_machine_%s", resultID)
		outR := fmt.Sprintf("results_readable_%s", resultID)
		outT := fmt.Sprintf("rewritten_trace_%s", resultID)
		runAnalyzer(tracePath, noRewriteFlag, analyisCasesFlag, outR,
			outM, ignoreAtomicsFlag, fifoFlag, ignoreCriticalSectionFlag, rewriteAllFlag, outT,
			timeoutAnalysis, ignoreRewriteFlag, fuzzing, onlyAPanicAndLeakFlag)
	}
	if err != nil {
		utils.LogError("Analyzer failed: ", err)
	}
	resTimes["analyzer"] += time.Since(startTime)

	fileOuputRead, err := os.OpenFile(output, os.O_RDONLY, 0644)
	if err != nil {
		utils.LogError("Could not open file: ", err)
	}
	outFileContent, err := io.ReadAll(fileOuputRead)
	if err != nil {
		utils.LogError("Could not read file: ", err)
	}
	lines := strings.Split(string(outFileContent), "\n")

	durationOther := time.Duration(0)
	for _, line := range lines {
		if strings.HasPrefix(line, "AdvocateAnalysisTimes:") {
			line = strings.TrimPrefix(line, "AdvocateAnalysisTimes:")
			elems := strings.Split(line, "#")

			timeAnaFloat, _ := strconv.ParseFloat(elems[0], 64)
			timeLeakFloat, _ := strconv.ParseFloat(elems[1], 64)
			timePanicFloat, _ := strconv.ParseFloat(elems[2], 64)
			timeOtherFloat, _ := strconv.ParseFloat(elems[3], 64)

			resTimes["analysis"] += time.Duration(timeAnaFloat * float64(time.Second))
			resTimes["leak"] += time.Duration(timeLeakFloat * float64(time.Second))
			resTimes["panic"] += time.Duration(timePanicFloat * float64(time.Second))
			durationOther += time.Duration(timeOtherFloat * float64(time.Second))
		}
	}
	fileOuputRead.Close()
	resTimes["hb"] += resTimes["analysis"] - resTimes["leak"] - resTimes["panic"] - durationOther
	if resTimes["hb"] < 0 {
		resTimes["hb"] = 0
	}

	utils.LogInfo("Finished Analyzer")
}

func unitTestReplay(pathToGoRoot, pathToPatchedGoRuntime, dir, pkg, file, testName string, resTimes map[string]time.Duration, rerecorded bool) int {
	pathPkg := filepath.Join(dir, pkg)
	var rewrittenTraces = make([]string, 0)
	if rerecorded {
		rewrittenTraces, _ = filepath.Glob(filepath.Join(pathPkg, "rewritten_trace_*_*"))
	} else {
		rewrittenTraces, _ = filepath.Glob(filepath.Join(pathPkg, "rewritten_trace_*"))
	}
	utils.LogInfof("Found %d rewritten traces\n", len(rewrittenTraces))

	timeoutRepl := time.Duration(0)
	if timeoutReplay == -1 {
		timeoutRepl = 500 * resTimes["record"]
		timeoutRepl = min(timeoutRepl, 15*time.Minute)
	} else {
		timeoutRepl = time.Duration(timeoutReplay) * time.Second
	}
	timeoutReplString := fmt.Sprintf("%ds", int(timeoutRepl.Seconds()))

	rerecordCounter := 0
	for i, trace := range rewrittenTraces {
		traceNum := extractTraceNumber(trace)
		// record := getRerecord(trace)
		record := false

		// limit the number of rerecordings
		if numberRerecord != -1 {
			if record {
				rerecordCounter++
				if rerecordCounter > numberRerecord {
					continue
				}
			}
		}

		headerInserterUnit(file, testName, true, -1, traceNum, int(timeoutRepl.Seconds()), record)

		os.Setenv("GOROOT", pathToGoRoot)

		utils.LogInfof("Run replay %d/%d\n", i+1, len(rewrittenTraces))
		startTime := time.Now()
		pkgPath := utils.MakePathLocal(pkg)
		runCommand(pathToPatchedGoRuntime, "test", "-v", "-count=1", "-timeout", timeoutReplString, "-run="+testName, pkgPath)
		resTimes["replay"] += time.Since(startTime)
		utils.LogInfof("Finished replay %d/%d\n", i+1, len(rewrittenTraces))

		os.Unsetenv("GOROOT")

		// Remove reorder header
		headerRemoverUnit(file)
	}

	return len(rewrittenTraces)
}

// func unitTestReanalyzeLeaks(pathToGoRoot, pathToPatchedGoRuntime, pathToAnalyzer, dir, pkg, file, testName, output string, resTimes map[string]time.Duration) (int, int) {
// 	pathPkg := filepath.Join(dir, pkg)
// 	rerecordedTraces, _ := filepath.Glob(filepath.Join(pathPkg, "advocateTraceReplay_*"))
// 	fmt.Printf("\nFound %d rerecorded traces\n\n", len(rerecordedTraces))

// 	for _, trace := range rerecordedTraces {
// 		number := extractTraceNumber(trace)
// 		traceName := filepath.Base(trace)
// 		unitTestAnalyzer(pathToAnalyzer, dir, pkg, traceName, output, resTimes, number)
// 	}
// 	numberRerecord = 0
// 	recorded := false // for now do not rerecord
// 	nrRewTrace := unitTestReplay(pathToGoRoot, pathToPatchedGoRuntime, dir, pkg, file, testName, resTimes, recorded)

// 	return nrRewTrace, len(rerecordedTraces)
// }

func getRerecord(trace string) bool {
	data, err := os.ReadFile(filepath.Join(trace, "rewrite_info.log"))
	if err != nil {
		return false
	}

	elems := strings.Split(string(data), "#")
	if len(elems) != 3 {
		return false
	}

	if len(elems[1]) == 0 {
		return false
	}

	return string([]rune(elems[1])[0]) == "S"
	// return string([]rune(elems[1])[0]) == "L"

}
