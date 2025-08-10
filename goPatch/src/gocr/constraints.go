// File: constraints.go
// Brief: Read constraints
//
// Created: 2025-07-14
//
// License: BSD-3-Clause

package gocr

import (
	"bufio"
	"os"
	"runtime"
	"strings"
)

// Read the file containing the constraints
//
// Parameter:
//   - path string: path to the file containing the constraints
//
// Returns:
//   - error
func readFuzzingFile(pathSelect string) error {
	file, err := os.Open(pathSelect)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		elems := strings.Split(line, ";")
		if len(elems) != 4 {
			continue
		}

		positive := (elems[0] == "+") // positive constraint
		t := elems[1]                 // type, e.g. "C" for channel

		elem1 := strings.Split(elems[2], ",")[1] // ignore routine
		elem2 := strings.Split(elems[3], ",")[1] // ignore routine

		runtime.AddConstraint(positive, t, elem1, elem2)
	}

	return nil
}
