// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: headerUnitTests.go
// Brief: Functions to add and remove the gocdr header into file containing
//    unit tests
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
	"gocdr/utils/flags"
	"os"
	"strings"

	"gocdr/utils/log"
)

// Add the header into a unit test
//
// Parameter:
//   - fileName string: path to the file containing the the test
//   - testName string: name of the test
//   - replay bool: true for replay, false for only recording
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - replayInfo string: path of the fuzzing trace or if the replay trace
//   - record bool: true to rerecord the leaks
//   - output *os.File: output file
//
// Returns:
//   - error
func headerInserterUnit(fileName, testName string, replay bool, fuzzing int, replayInfo string, record bool, output *os.File) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", fileName)
	}

	testExists, err := testExists(fileName, testName)
	if err != nil {
		return err
	}

	if !testExists {
		return errors.New("Test Method not found in file")
	}

	return addHeaderUnit(fileName, testName, replay, fuzzing, replayInfo, record, output)
}

// Remove all headers from a unit test file
//
// Parameter:
//   - fileName string: path to the file containing the the test
//   - testName string: name of the test
//
// Returns:
//   - error
func headerRemoverUnit(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("Please provide a file name")
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", fileName)
	}

	return removeHeaderUnit(fileName)
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

// Add the header into the unit tests. Do not call directly.
// Call via headerInserterUnit. This functions assumes, that the
// test exists.
//
// Parameter:
//   - fileName string: path to the file
//   - testName string: name of the test
//   - replay bool: true for replay, false for only recording
//   - replayInfo string: path of the fuzzing trace or if the replay trace
//   - record bool: true to rerecord the trace
//   - output *os.File: output file
//
// Returns:
//   - error
func addHeaderUnit(fileName string, testName string, replay bool, fuzzing int, replayInfo string, record bool, output *os.File) error {
	importAdded := false
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if replay && fuzzing >= 0 {
		return fmt.Errorf("Cannot add header for replay and fuzzing at the same time")
	}

	atomicReplayStr := "true"
	if flags.IgnoreAtomics {
		atomicReplayStr = "false"
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	currentLine := 0

	fmt.Fprintln(output, "FileName: ", fileName)
	fmt.Fprintln(output, "TestName: ", testName)

	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		lines = append(lines, line)

		if strings.Contains(line, "import \"") && !importAdded {
			lines = append(lines, "import \"gocdr\"")
			fmt.Fprintln(output, "Import added at line:", currentLine)
			importAdded = true
		} else if strings.Contains(line, "import (") && !importAdded {
			lines = append(lines, "\t\"gocdr\"")
			fmt.Fprintln(output, "Import added at line:", currentLine)
			importAdded = true
		}

		if strings.Contains(line, "func "+testName) {
			if replay { // replay
				replayPath := ""
				log.Debug(replayInfo)
				if replayInfo != "" {
					replayPath = "rewrittenTrace_" + replayInfo
				} else if flags.TracePath != "" {
					replayPath = flags.TracePath
					log.Debug("SET")
				} else {
					replayPath = "gocdrTrace"
				}
				if record {
					lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  gocdr.InitFuzzing("%s", %d)
  defer gocdr.FinishFuzzing()
  // ======= Preamble End =======`, replayPath, flags.TimeoutReplay))
				} else {
					log.Debug(replayPath)
					lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  gocdr.InitReplay("%s", %d, %s)
  defer gocdr.FinishReplay()
  // ======= Preamble End =======`, replayPath, flags.TimeoutReplay, atomicReplayStr))
				}
			} else if fuzzing > 0 {
				lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  gocdr.InitFuzzing("%s", %d)
  defer gocdr.FinishFuzzing()
  // ======= Preamble End =======`, replayInfo, flags.TimeoutRecording))
			} else { // recording
				lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  gocdr.InitTracing(%d)
  defer gocdr.FinishTracing()
  // ======= Preamble End =======`, flags.TimeoutRecording))
			}
			fmt.Fprintln(output, "Header added at line:", currentLine)
			fmt.Fprintf(output, "Header added at file: %s\n", fileName)
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

// Remove the header from the unit test. Do not call directly.
// Call via headerRemoverUnit. This functions assumes, that the
// test exists.
//
// Parameter:
//   - fileName string: path to the file
//
// Returns:
//   - error
func removeHeaderUnit(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inPreamble := false
	inImports := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "// ======= Preamble Start =======") {
			inPreamble = true
			continue
		}

		if strings.Contains(line, "// ======= Preamble End =======") {
			inPreamble = false
			continue
		}

		if inPreamble {
			continue
		}

		if strings.Contains(line, "import \"gocdr\"") {
			continue
		}

		if strings.Contains(line, "import (") {
			inImports = true
		}

		if inImports && strings.Contains(line, "\"gocdr\"") {
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
