//
// File: headerUnitTests.go
// Brief: Functions to add and remove the ADVOCATE header into file containing
//    unit tests
//
// Created: 2024-10-29
//
// License: BSD-3-Clause

package toolchain

import (
	"io"
	"os"
	"strings"
)

// For a given run, check if it was terminated by a timeout
//
// Parameter:
//   - output (string): path to the output.log file
//
// Returns:
//   - true if an timeout occurred, false otherwise
func checkForTimeout(output string) bool {
	outFile, err := os.Open(output)
	if err != nil {
		return false
	}
	defer outFile.Close()

	content, err := io.ReadAll(outFile)
	if err != nil {
		return false
	}

	if strings.Contains(string(content), "panic: test timed out after") {
		return true
	}

	return false
}
