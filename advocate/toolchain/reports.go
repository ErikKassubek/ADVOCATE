// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: headerUnitTests.go
// Brief: Functions to generate bug reports
//
// Author: Erik Kassubek, Mario Occhinegro
// Created: 2024-09-18
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/results/explanation"
	"advocate/utils/log"
)

var movedTraces int = 0

// Generate the bug reports
//
// Parameter:
//   - folderName string: path to folder containing the results
//   - fuzzingRun int: number of fuzzing run, -1 for not fuzzing
func generateBugReports(folder string, fuzzing int) {
	err := explanation.CreateOverview(folder, true, fuzzing)
	if err != nil {
		log.Error("Error creating explanation: ", err.Error())
	}
}
