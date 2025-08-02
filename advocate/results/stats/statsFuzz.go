// Copyright (c) 2025 Sebastian Pohsner
//
// File: statsMisc.go
// Brief: Collect miscellaneous statistics about the advocate run
//
// Author: Erik Kassubek
// Created: 2025-02-25
//
// License: BSD-3-Clause

package stats

import (
	"advocate/analysis/data"
	"advocate/fuzzing/gopie"
)

var fuzzStats = []string{
	"TestName",
	"NrMut",
	"NrMutInvalid",
	"NrMutDouble",
	"ActiveReleased",
}

// Collect stats about each fuzzing run
//
// Parameter:
//   - dataPath string: path to the result folder
//
// Returns:
//   - map[string]int: map with the stats
//   - error
func statsFuzz(dataPath, testName string) (map[string]int, error) {
	stats := map[string]int{}

	stats["NrMut"] = gopie.NumberTotalMuts
	stats["NrMutInvalid"] = gopie.NumberInvalidMuts
	stats["NrMutDouble"] = gopie.NumberDoubleMuts
	stats["ActiveReleased"] = data.ActiveReleased

	return stats, nil
}
