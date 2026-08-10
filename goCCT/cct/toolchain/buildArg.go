// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: header.go
// Brief: Functions to add and remove the GoCCT header into/from files containing
//    a main function
//
// Author: Erik Kassubek, Mario Occhinegro
//
// License: BSD-3-Clause

package toolchain

import (
	"bufio"
	"errors"
	"fmt"
	"gocct/utils/flags"
	"gocct/utils/log"
	"gocct/utils/paths"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getBuildArg(fileName string, replay bool, tracePath string,
	replayTimeout int, record bool, fuzzing int, fuzzingTrace string) (buildArg string) {

	buildArg = buildFlagsDefault + " "

	atomicReplayStr := "true"
	if flags.IgnoreAtomics {
		atomicReplayStr = "false"
	}

	if replay { // replay
		if record {
			buildArg += fmt.Sprintf("-gocctfuzzing -gocctpath=%s -goccttimeout=%d", tracePath, flags.Timeout)
		} else {
			buildArg += fmt.Sprintf("-gocctreplay -gocctpath=%s -goccttimeout=%d -gocctatomic=%s", tracePath, replayTimeout, atomicReplayStr)
		}
	} else if fuzzing > 0 {
		buildArg += fmt.Sprintf("-gocctfuzzing -gocctpath=%s -goccttimeout=%d", fuzzingTrace, replayTimeout)
	} else { // recording
		buildArg += fmt.Sprintf("-goccttrace  -goccttimeout=%d", replayTimeout)
	}

	// buildArg += "'"

	return
}

// ============================================
// MARK: Main
// ============================================

// Insert the header into a main function and return the build parameters
//
// Parameter:
//   - fileName string: path to the main file
//   - replay bool: true for replay, false for only recording
//   - replayNumber string: id of the trace to replay
//   - replayTimeout int: replay for timeout
//   - record bool: if both replay and record are set, the replay is rerecorded
//   - fuzzing int: fuzzing run, if no fuzzing: -1, for initial run: 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - static bool: set true for static dummy
//
// Returns:
//   - string: build parameters
//   - string: file where import was added
//   - int: line where import was added
//   - error
func importInsertMain(fileName string, replay bool, replayNumber string,
	replayTimeout int, record bool, fuzzing int, fuzzingTrace string, static bool) (string, string, int, error) {
	if fileName == "" {
		return "", "", -1, errors.New("Please provide a file  name")
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", "", -1, fmt.Errorf("File %s does not exist", fileName)
	}

	exists, err := mainMethodExists(fileName)
	if err != nil {
		return "", "", -1, err
	}

	if !exists {
		return "", "", -1, fmt.Errorf("Main Method not found in file")
	}

	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return "", "", -1, fmt.Errorf("Could not open main file to add header")
	}
	defer file.Close()

	if !static {
		fmt.Println("FileName: ", fileName)
		fmt.Println("TestName: Main")
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	importLine := -1
	currentLine := 0
	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		lines = append(lines, line)

		if strings.Contains(line, "package main") {
			if static {
				lines = append(lines, "")
			} else {
				lines = append(lines, "import _ \"gocct\"")
				fmt.Println("Import added at line:", currentLine)
			}
			importLine = currentLine
		} else if strings.Contains(line, "import \"") && importLine == -1 {
			if static {
				lines = append(lines, "")
			} else {
				lines = append(lines, "import _ \"gocct\"")
				fmt.Println("Import added at line:", currentLine)
			}
			importLine = currentLine
		} else if strings.Contains(line, "import (") && importLine == -1 {
			if static {
				lines = append(lines, "")
			} else {
				lines = append(lines, "\t _ \"gocct\"")
				fmt.Println("Import added at line:", currentLine)
			}
			importLine = currentLine
		}
	}

	file.Truncate(0)
	file.Seek(0, 0)
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()

	replayPath := ""
	if replayNumber != "" {
		replayPath = "rewrittenTrace_" + replayNumber
	} else if flags.TracePath != "" {
		replayPath = filepath.Base(flags.TracePath)
	} else {
		replayPath = "gocctTrace"
	}

	return getBuildArg(fileName, replay, replayPath, flags.Timeout, record, fuzzing, fuzzingTrace), fileName, importLine, nil
}

// Remove the header from a file with a header in a main function
//
// Parameter:
//   - fileName string: name of the file
//
// Returns:
//   - error
func importRemoveMain() error {
	if _, err := os.Stat(paths.Prog); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", paths.Prog)
	}

	file, err := os.Open(paths.Prog)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inImportBlock := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "import (") {
			inImportBlock = true
			lines = append(lines, line)
		} else if inImportBlock && strings.Contains(line, ")") {
			inImportBlock = false
			lines = append(lines, line)
		} else if inImportBlock && strings.Contains(line, "\"gocct\"") {
			continue
		} else if strings.Contains(line, "import _ \"gocct\"") {
			continue
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile(paths.Prog, []byte(strings.Join(lines, "\n")), 0644)
}

// Check if there is a main function in the given file
//
// Parameter:
//   - fileName string: name of the file
//
// Returns
//
//   - bool: true if the file contains a main function, false otherwise
//   - error
func mainMethodExists(fileName string) (bool, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	defer file.Close()

	regexStr := "func main\\(\\) {"
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if regex.MatchString(line) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// ============================================
// MARK: Test
// ============================================

// Add the header into a unit test and return the build parameter
//
// Parameter:
//   - fileName string: path to the file containing the the test
//   - testName string: name of the test
//   - replay bool: true for replay, false for only recording
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - replayInfo string: path of the fuzzing trace or if the replay trace
//   - record bool: true to rerecord the leaks
//
// Returns:
//   - string: build args
//   - error
func importInsertUnit(fileName, testName string, replay bool, fuzzing int, replayInfo string, record bool) (string, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s does not exist", fileName)
	}

	testExists, err := testExists(fileName, testName)
	if err != nil {
		return "", err
	}

	if !testExists {
		return "", errors.New("Test Method not found in file")
	}

	importAdded := false
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if replay && fuzzing >= 0 {
		return "", fmt.Errorf("Cannot add header for replay and fuzzing at the same time")
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	currentLine := 0

	fmt.Println("FileName: ", fileName)
	fmt.Println("TestName: ", testName)

	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		lines = append(lines, line)

		if strings.Contains(line, "import \"") && !importAdded {
			lines = append(lines, "import _ \"gocct\"")
			fmt.Println("Import added at line:", currentLine)
			importAdded = true
		} else if strings.Contains(line, "import (") && !importAdded {
			lines = append(lines, "\t _ \"gocct\"")
			fmt.Println("Import added at line:", currentLine)
			importAdded = true
		}
	}

	file.Truncate(0)
	file.Seek(0, 0)
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()

	replayPath := ""
	if replayInfo != "" {
		replayPath = "rewrittenTrace_" + replayInfo
	} else if flags.TracePath != "" {
		replayPath = filepath.Base(flags.TracePath)
	} else {
		replayPath = "gocctTrace"
	}

	return getBuildArg(fileName, replay, replayPath, flags.Timeout, record, fuzzing, replayInfo), nil
}

// Remove all headers from a unit test file
//
// Parameter:
//   - fileName string: path to the file containing the the test
//   - testName string: name of the test
//
// Returns:
//   - error
func importRemoverUnit(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("Please provide a file name")
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", fileName)
	}

	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inImports := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "import _ \"gocct\"") {
			continue
		}

		if strings.Contains(line, "import (") {
			inImports = true
		}

		if inImports && strings.Contains(line, "\"gocct\"") {
			continue
		}

		if strings.Contains(line, ")") {
			inImports = false
		}

		lines = append(lines, line)
	}

	file.Truncate(0)
	file.Seek(0, 0)
	writer := bufio.NewWriter(file)

	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}

	writer.Flush()

	return nil
}

// Check if a test exists
//
// Parameter:
//   - fileName string: path to the file
//   - testName string: name of the test
//
// Returns:
//   - bool: true if the test exists, false otherwise
//   - error
func testExists(fileName string, testName string) (bool, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "func "+testName) && strings.Contains(line, "testing.T") {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// ============================================
// MARK: Static
// ============================================

func ImportInsertStatic() (file string, line int) {
	if flags.ModeMain {
		var err error
		_, file, line, err = importInsertMain(paths.Prog, false, "1", flags.Timeout, false, 0, "", true)
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		panic("STATIC FOR TESTS NOT NOT IMPLEMENTED YET")
	}

	return
}

func ImportRemoveStatic(file string, line int) {
	if line < 1 {
		log.Errorf("ImportRemoveStatic failed: line number must be >= 1")
		return
	}

	data, err := os.ReadFile(file)
	if err != nil {
		log.Error("ImportRemoveStatic failed: ", err.Error())
		return
	}

	lines := strings.Split(string(data), "\n")

	if line > len(lines) {
		log.Errorf("ImportRemoveStatic failed: line %d out of range (file has %d lines)", line, len(lines))
		return
	}

	// Remove the requested line (1-based index)
	lines = append(lines[:line], lines[line+1:]...)

	output := strings.Join(lines, "\n")
	err = os.WriteFile(file, []byte(output), 0644)

	if err != nil {
		log.Error("ImportRemoveStatic failed: ", err.Error())
	}
}
