// File: statsMisc.go
// Brief: Collect miscellaneous statistics about the goCR run
//
// Created: 2025-02-25
//
// License: BSD-3-Clause

package stats

import (
	"goCR/analysis/data"
	"goCR/fuzzing/gopie"
)

var fuzzStats = []string{
	"TestName",
	"NrMut",
	"NrMutInvalid",
	"ActiveReleased",
	"AllActiveReleased",
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
	stats["ActiveReleased"] = data.ActiveReleased
	stats["AllActiveReleased"] = data.AllActiveReleased

	return stats, nil
}
