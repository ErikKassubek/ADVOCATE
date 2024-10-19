// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: headerUnitTests.go
// Brief: Functions to add and remove the ADVOCATE header into file containing
//    unit tests
//
// Author: Erik Kassubek, Mario Occhinegro
// Created: 2024-09-18
// Last Changed 2024-09-18
//
// License: BSD-3-Clause

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

/*
 * Add the header into a unit test
 * Args:
 *    fileName (string): path to the file containing the the test
 *    testName (string): name of the test
 *    replay (bool): true for replay, false for only recording
 *    replayNumber (string): id of the trace to replay
 *    timeoutReplay (int): timeout for replay
 * Returns:
 *    error
 */
func headerInserterUnit(fileName string, testName string, replay bool, replayNumber string, timeoutReplay int) error {
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

	return addHeaderUnit(fileName, testName, replay, replayNumber, timeoutReplay)
}

/*
 * Remove all headers from a unit test file
 * Args:
 *    fileName (string): path to the file containing the the test
 *    testName (string): name of the test
 * Returns:
 *    error
 */
func headerRemoverUnit(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("Please provide a file name")
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", fileName)
	}

	return removeHeaderUnit(fileName)
}

/*
 * Check if a test exists
 * Args:
 *    fileName (string): path to the file
 *    testName (string): name of the test
 * Returns:
 *    bool: true if the test exists, false otherwise
 *    error
 */
func testExists(fileName string, testName string) (bool, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	defer file.Close()

	regexStr := "func " + testName + "\\(*t \\*testing.T*\\) {"
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

/*
 * Add the header into the unit tests. Do not call directly.
 * Call via headerInserterUnit. This functions assumes, that the
 * test exists.
 * Args:
 *    fileName (string): path to the file
 *    testName (string): name of the test
 *    replay (bool): true for replay, false for only recording
 *    replayNumber (string): id of the trace to replay
 *    timeoutReplay (int): timeout for replay
 * Returns:
 *    error
 */
func addHeaderUnit(fileName string, testName string, replay bool, replayNumber string, timeoutReplay int) error {
	importAdded := false
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		line := scanner.Text()
		lines = append(lines, line)

		if strings.Contains(line, "import \"") && !importAdded {
			lines = append(lines, "import \"advocate\"")
			fmt.Println("Import added at line:", currentLine)
			importAdded = true
		} else if strings.Contains(line, "import (") && !importAdded {
			lines = append(lines, "\t\"advocate\"")
			fmt.Println("Import added at line:", currentLine)
			importAdded = true
		}

		if strings.Contains(line, "func "+testName) {
			if replay {
				lines = append(lines, fmt.Sprintf(`	// ======= Preamble Start =======
  advocate.EnableReplay(%s, true, %d)
  defer advocate.WaitForReplayFinish()
  // ======= Preamble End =======`, replayNumber, timeoutReplay))
			} else {
				lines = append(lines, `	// ======= Preamble Start =======
  advocate.InitTracing()
  defer advocate.Finish()
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

/*
 * Remove the header from the unit test. Do not call directly.
 * Call via headerRemoverUnit. This functions assumes, that the
 * test exists.
 * Args:
 *    fileName (string): path to the file
 * Returns:
 *    error
 */
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

		if strings.Contains(line, "import \"advocate\"") {
			continue
		}

		if strings.Contains(line, "import (") {
			inImports = true
		}

		if inImports && strings.Contains(line, "\"advocate\"") {
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
