// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: headerMain.go
// Brief: Functions to add and remove the ADVOCATE header into/from files containing
//    a main function
//
// Author: Erik Kassubek, Mario Occhinegro
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Insert the header into a main function
//
// Parameter:
//   - fileName string: path to the main file
//   - replay bool: true for replay, false for only recording
//   - replayNumber string: id of the trace to replay
//   - replayTimeout int: replay for timeout
//   - record bool: if both replay and record are set, the replay is rerecorded
//   - fuzzing int: fuzzing run, if no fuzzing: -1, for initial run: 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//
// Returns:
//   - error
func headerInserterMain(fileName string, replay bool, replayNumber string,
	replayTimeout int, record bool, fuzzing int, fuzzingTrace string) error {
	if fileName == "" {
		return errors.New("Please provide a file  name")
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("File %s does not exist", fileName)
	}

	return addMainHeader(fileName, replay, replayNumber, replayTimeout, record, fuzzing, fuzzingTrace)
}

// Remove the header from a file with a header in a main function
//
// Parameter:
//   - fileName string: name of the file
//
// Returns:
//   - error
func headerRemoverMain(fileName string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", fileName)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inImportBlock := false
	numberOfLinesToSkip := 0

	for scanner.Scan() {
		line := scanner.Text()

		if numberOfLinesToSkip > 0 {
			numberOfLinesToSkip--
			continue
		} else if strings.Contains(line, "// ======= Preamble Start =======") {
			numberOfLinesToSkip = 3
			continue
		} else if strings.Contains(line, "import (") {
			inImportBlock = true
			lines = append(lines, line)
		} else if inImportBlock && strings.Contains(line, ")") {
			inImportBlock = false
			lines = append(lines, line)
		} else if inImportBlock && strings.Contains(line, "\"advocate\"") {
			continue
		} else if strings.Contains(line, "import \"advocate\"") {
			continue
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile(fileName, []byte(strings.Join(lines, "\n")), 0644)
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

// Add the header into the main file
//
// Parameter:
//   - fileName string: name of the file containing the main routine
//   - replay bool: true for replay, false for just recording
//   - replayNumber int: id of the trace to replay
//   - replayTimeout int: replay for timeout
//   - record bool: if both replay and record are set, the replay is rerecorded
//   - fuzzing int: fuzzing run, if no fuzzing: -1, for initial run: 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//
// Returns:
//   - error
func addMainHeader(fileName string, replay bool, replayNumber string,
	replayTimeout int, record bool, fuzzing int, fuzzingTrace string) error {
	exists, err := mainMethodExists(fileName)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("Main Method not found in file")
	}

	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("Could not open main file to add header")
	}
	defer file.Close()

	atomicReplayStr := "false"
	if replayAtomic {
		atomicReplayStr = "true"
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	importAdded := false
	currentLine := 0
	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		lines = append(lines, line)

		if strings.Contains(line, "package main") {
			lines = append(lines, "import \"advocate\"")
			importAdded = true
		} else if strings.Contains(line, "import \"") && !importAdded {
			lines = append(lines, "import \"advocate\"")
			importAdded = true
		} else if strings.Contains(line, "import (") && !importAdded {
			lines = append(lines, "\t\"advocate\"")
			importAdded = true
		}

		if strings.Contains(line, "func main() {") {
			if replay { // replay
				replayPath := ""
				if replayNumber != "" {
					replayPath = "rewrittenTrace_" + replayNumber
				} else if tracePathFlag != "" {
					replayPath = filepath.Base(tracePathFlag)
				} else {
					replayPath = "advocateTrace"
				}
				if record {
					lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  advocate.InitReplayTracing("%s", false, %d, %s)
  defer advocate.FinishReplayTracing()
  // ======= Preamble End =======`, replayPath, replayTimeout, atomicReplayStr))
				} else {
					lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  advocate.InitReplay("%s", %d, %s)
  defer advocate.FinishReplay()
  // ======= Preamble End =======`, replayPath, replayTimeout, atomicReplayStr))
				}
			} else if fuzzing > 0 {
				lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  advocate.InitFuzzing("%s", %d)
  defer advocate.FinishTracing()
  // ======= Preamble End =======`, fuzzingTrace, timeoutRecording))
			} else { // recording
				lines = append(lines, `	// ======= Preamble Start =======
  advocate.InitTracing()
  defer advocate.FinishTracing()
  // ======= Preamble End =======`)
			}
			fmt.Println("Header added at line:", currentLine)
			fmt.Printf("Header added at file: %s\n", fileName)
		}
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
