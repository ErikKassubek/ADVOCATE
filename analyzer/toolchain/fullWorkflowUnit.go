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
	"analyzer/timer"
	"analyzer/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
 *    firstRun (bool): this is the first run, only set to false for fuzzing (except for the first fuzzing)
 *    cont (bool): continue an already started run
 * Returns:
 *    error
 */
func runWorkflowUnit(pathToAdvocate, dir, progName string,
	measureTime, notExecuted, createStats bool, fuzzing int, keepTraces, firstRun, cont bool, fileNumber, testNumber int) error {
	// Validate required inputs
	if pathToAdvocate == "" {
		return errors.New("Path to advocate is empty")
	}
	if dir == "" {
		return errors.New("Directory is empty")
	}

	isFuzzing := (fuzzing != -1)

	// Change to the directory
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("Failed to change directory: %v", dir)
	}

	if firstRun && !cont {
		os.RemoveAll("advocateResult")
		if err := os.MkdirAll("advocateResult", os.ModePerm); err != nil {
			return fmt.Errorf("Failed to create advocateResult directory: %v", err)
		}
	}

	// Find all _test.go files in the directory
	testFiles, _, totalFiles, err := FindTestFiles(dir, cont && testName == "")
	if err != nil {
		return fmt.Errorf("Failed to find test files: %v", err)
	}

	attemptedTests, skippedTests, currentFile := 0, 0, fileNumber

	// resultPath := filepath.Join(dir, "advocateResult")

	ranTest := false
	// Process each test file
	for _, file := range testFiles {
		if testName == "" {
			utils.LogInfof("Progress %s: %d/%d", progName, currentFile, totalFiles)
			utils.LogInfof("Processing file: %s", file)
		}

		packagePath := filepath.Dir(file)
		testFunctions, err := FindTestFunctions(file)
		if err != nil {
			utils.LogInfof("Could not find test functions in %s: %v", file, err)
			continue
		}

		for _, testFunc := range testFunctions {
			if testName != "" && testName != testFunc {
				continue
			}

			analysis.Clear()

			if !isFuzzing {
				timer.ResetTest()
				timer.Start(timer.TotalTest)
			}

			ranTest = true

			attemptedTests++
			packageName := filepath.Base(packagePath)
			fileName := filepath.Base(file)
			utils.LogInfof("Running full workflow for test %s in package %s in file %s", testFunc, packageName, file)

			adjustedPackagePath := strings.TrimPrefix(packagePath, dir)
			if !strings.HasSuffix(adjustedPackagePath, string(filepath.Separator)) {
				adjustedPackagePath = adjustedPackagePath + string(filepath.Separator)
			}
			fileNameWithoutEnding := strings.TrimSuffix(fileName, ".go")
			directoryName := filepath.Join("advocateResult", fmt.Sprintf("file(%d)-test(%d)-%s-%s", currentFile, attemptedTests, fileNameWithoutEnding, testFunc))
			if cont && fileNumber != 0 {
				directoryName = filepath.Join("advocateResult", fmt.Sprintf("file(%d)-test(%d)-%s-%s", fileNumber, testNumber, fileNameWithoutEnding, testFunc))
			}
			directoryPath := filepath.Join(dir, directoryName)
			if fuzzing < 1 {
				utils.LogInfo("Create ", directoryName)
				if err := os.MkdirAll(directoryName, os.ModePerm); err != nil {
					utils.LogErrorf("Failed to create directory %s: %v", directoryName, err)
					if !isFuzzing {
						timer.Stop(timer.TotalTest)
					}
					continue
				}
			}

			// Execute full workflow
			nrReplay, anaPassed, err := unitTestFullWorkflow(pathToAdvocate, dir, testFunc, adjustedPackagePath, file, fuzzing)

			timer.UpdateTimeFileDetail(progName, testFunc, nrReplay)

			if !isFuzzing {
				timer.ResetTest()
				timer.UpdateTimeFileOverview(progName, testFunc)
			}

			// Move logs and results to the appropriate directory
			total := fuzzing != -1
			collect(dir, packagePath, directoryPath, total)

			if err != nil {
				utils.LogErrorf(err.Error())
				skippedTests++
			}

			if anaPassed {
				generateBugReports(directoryPath, fuzzing)
				if createStats {
					// create statistics
					err := stats.CreateStats(directoryPath, progName, testFunc, movedTraces, fuzzing)
					if err != nil {
						utils.LogError("Could not create statistics: ", err.Error())
					}
				}
			}

			if !keepTraces {
				removeTraces(dir)
			}

			if total {
				removeLogs(dir)
			}

			if !isFuzzing {
				timer.Stop(timer.TotalTest)
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
		utils.LogInfof("Attempted tests: %d", attemptedTests)
		utils.LogInfof("Skipped tests: %d", skippedTests)
	} else {
		utils.LogInfof("Finished full work flow for %s", testName)
	}

	return nil
}

/*
 * Function to find all _test.go files in the specified directory
 * Args:
 *    dir (string): folder to search in
 *    cont (bool): only return test files not already in the advocateResult
 * Returns:
 *    []string: found files
 *    int: min file num, only if cont, otherwise 0
 *    int: total number of files
 *    error
 */
func FindTestFiles(dir string, cont bool) ([]string, int, int, error) {
	var testFiles []string

	alreadyProcessed, maxFileNum := make(map[string]struct{}), 0
	var err error

	alreadyProcessed, maxFileNum, err = getFilesInResult(dir, cont)
	if err != nil {
		utils.LogError(err)
		return testFiles, 0, 0, err
	}

	totalNumFiles := 0
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if strings.HasSuffix(name, "_test.go") {
			totalNumFiles++
			if _, ok := alreadyProcessed[name]; !cont || !ok {
				testFiles = append(testFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		utils.LogError(err)
	}
	return testFiles, maxFileNum, totalNumFiles, err
}

func getFilesInResult(dir string, cont bool) (map[string]struct{}, int, error) {
	res := make(map[string]struct{})

	path := filepath.Join(dir, "advocateResult")

	patternPrefix := `file\([0-9]+\)-test\([0-9]+\)-`
	patternFileNum := `^file\((\d+)\)-test\(\d+\)-.+$`
	rePrefix := regexp.MustCompile(patternPrefix)
	reNum := regexp.MustCompile(patternFileNum)

	files, err := os.ReadDir(path)
	if err != nil {
		return res, 0, err
	}

	maxFileNum := -1
	maxKey := ""
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		name := file.Name()
		nameClean := rePrefix.ReplaceAllString(name, "")
		lastIndex := strings.LastIndex(nameClean, "-")
		if lastIndex != -1 {
			nameClean = nameClean[:lastIndex] // Keep everything before the last separator
		}

		numbers := reNum.FindStringSubmatch(name)

		if len(numbers) > 1 {
			numberInt, err := strconv.Atoi(numbers[1])
			if err != nil {
				return res, 0, err
			}
			if numberInt > maxFileNum {
				maxKey = nameClean + ".go"
				maxFileNum = numberInt
			}
		}

		res[nameClean+".go"] = struct{}{}
	}

	// remove all folders created by the last file and remove the file name from the processed
	if cont && maxFileNum != -1 {
		for _, file := range files {
			if !file.IsDir() || !strings.Contains(file.Name(), fmt.Sprintf("file(%d)", maxFileNum)) {
				continue
			}

			_ = os.RemoveAll(filepath.Join(path, file.Name()))
		}
		utils.LogError()
		delete(res, maxKey)
		maxFileNum = maxFileNum - 1
	}

	return res, maxFileNum, nil
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
 *    int: number of run replays
 *    bool: true if analysis passed without error
 *    error
 */
func unitTestFullWorkflow(pathToAdvocate, dir, testName, pkg, file string, fuzzing int) (int, bool, error) {
	output := "output.log"

	outFile, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, false, fmt.Errorf("Failed to open log file: %v", err)
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
		return 0, false, errors.New("Path to advocate is empty")
	}
	if dir == "" {
		return 0, false, errors.New("Directory is empty")
	}
	if testName == "" {
		return 0, false, errors.New("Test name is empty")
	}
	// if pkg == "" {
	// 	return 0, errors.New("Package is empty")
	// }
	if file == "" {
		return 0, false, errors.New("Test file is empty")
	}

	pathToPatchedGoRuntime := filepath.Join(pathToAdvocate, "go-patch/bin/go")
	pathToGoRoot := filepath.Join(pathToAdvocate, "go-patch")
	pathToAnalyzer := filepath.Join(pathToAdvocate, "analyzer/analyzer")

	if runtime.GOOS == "windows" {
		pathToPatchedGoRuntime += ".exe"
	}

	// Change to the directory
	if err := os.Chdir(dir); err != nil {
		return 0, false, fmt.Errorf("Failed to change directory: %v", err)
	}

	pkg = strings.TrimPrefix(pkg, dir)

	if measureTime && fuzzing < 1 {
		err := unitTestRun(pkg, file, testName)
		if err != nil {
			if err != nil {
				if checkForTimeout(output) {
					utils.LogTimeout("Running T0 timed out")
				}
			}
		}
	}

	err = unitTestRecord(pathToGoRoot, pathToPatchedGoRuntime, pkg, file, testName, fuzzing, output)
	if err != nil {
		utils.LogError("Recording failed: ", err.Error())
	}

	err = unitTestAnalyzer(pathToAnalyzer, dir, pkg, "advocateTrace", output, fuzzing)
	if err != nil {
		return 0, false, err
	}

	numberReplay := unitTestReplay(pathToGoRoot, pathToPatchedGoRuntime, dir, pkg, file, testName)

	return numberReplay, true, nil
}

// run the tests without recording/replay
func unitTestRun(pkg, file, testName string) error {
	timer.Start(timer.Run)
	defer timer.Stop(timer.Run)

	// Remove header just in case
	if err := headerRemoverUnit(file); err != nil {
		utils.LogError("Failed to remove header: ", err)
	}

	os.Unsetenv("GOROOT")

	utils.LogInfo("Run T0")
	packagePath := utils.MakePathLocal(pkg)
	var err error
	if timeoutRecording != -1 {
		timeoutRecString := fmt.Sprintf("%ds", timeoutRecording)
		err = runCommand("go", "test", "-v", "-timeout", timeoutRecString, "-count=1", "-run="+testName, packagePath)
	} else {
		err = runCommand("go", "test", "-v", "-count=1", "-run="+testName, packagePath)
	}

	return err
}

func unitTestRecord(pathToGoRoot, pathToPatchedGoRuntime, pkg, file, testName string, fuzzing int, output string) error {
	timer.Start(timer.Recording)
	defer timer.Stop(timer.Recording)

	isFuzzing := (fuzzing > 0)

	// Remove header just in case
	if err := headerRemoverUnit(file); err != nil {
		return fmt.Errorf("Failed to remove header: %v", err)
	}

	// Add header
	if err := headerInserterUnit(file, testName, false, fuzzing, "0", timeoutReplay, false); err != nil {
		return fmt.Errorf("Error in adding header: %v", err)
	}

	// Run the test
	utils.LogInfo("Run Recording")

	// Set GOROOT
	os.Setenv("GOROOT", pathToGoRoot)

	timeoutRecString := fmt.Sprintf("%ds", timeoutRecording)

	pkgPath := utils.MakePathLocal(pkg)
	err := runCommand(pathToPatchedGoRuntime, "test", "-v", "-timeout", timeoutRecString, "-count=1", "-run="+testName, pkgPath)
	if err != nil {
		if isFuzzing {
			if checkForTimeout(output) {
				utils.LogTimeout("Recording timed out")
			}
		} else {
			if checkForTimeout(output) {
				utils.LogTimeout("Fuzzing recording timed out")
			}
		}
	}

	err = os.Unsetenv("GOROOT")

	if err != nil {
		utils.LogErrorf("Failed to unset GOROOT: ", err.Error())
	}

	// Remove header after the test
	err = headerRemoverUnit(file)

	return err
}

// Apply analyzer
func unitTestAnalyzer(pathToAnalyzer, dir, pkg, traceName, output string, fuzzing int) error {
	pkgPath := filepath.Join(dir, pkg)
	tracePath := filepath.Join(pkgPath, traceName)

	utils.LogInfof("Run the analyzer for %s", tracePath)

	outM := filepath.Join(pkgPath, "results_machine.log")
	outR := filepath.Join(pkgPath, "results_readable.log")
	outT := filepath.Join(pkgPath, "rewritten_trace")
	err := runAnalyzer(tracePath, noRewriteFlag, analyisCasesFlag, outR,
		outM, ignoreAtomicsFlag, fifoFlag, ignoreCriticalSectionFlag, rewriteAllFlag,
		outT, ignoreRewriteFlag, fuzzing, onlyAPanicAndLeakFlag)

	if err != nil {
		return err
	}

	utils.LogInfo("Finished Analyzer")
	return nil
}

func unitTestReplay(pathToGoRoot, pathToPatchedGoRuntime, dir, pkg, file, testName string) int {
	timer.Start(timer.Replay)
	defer timer.Stop(timer.Replay)

	utils.LogInfo("Start Replay")

	pathPkg := filepath.Join(dir, pkg)

	rewrittenTraces, _ := filepath.Glob(filepath.Join(pathPkg, "rewritten_trace_*"))

	utils.LogInfof("Found %d rewritten traces", len(rewrittenTraces))

	timeoutRepl := time.Duration(0)
	if timeoutReplay == -1 {
		timeoutRepl = 500 * timer.GetTime(timer.Recording)
		timeoutRepl = max(min(timeoutRepl, 10*time.Minute), time.Duration(timeoutRecording)*time.Second*2)
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

		utils.LogInfof("Run replay %d/%d", i+1, len(rewrittenTraces))
		pkgPath := utils.MakePathLocal(pkg)
		runCommand(pathToPatchedGoRuntime, "test", "-v", "-count=1", "-timeout", timeoutReplString, "-run="+testName, pkgPath)
		utils.LogInfof("Finished replay %d/%d", i+1, len(rewrittenTraces))

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
