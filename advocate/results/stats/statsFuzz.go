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
	"advocate/analysis/baseA"
	"advocate/fuzzing/gopie"
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
// Returns:
//   - map[string]int: map with the stats
//   - error
func statsFuzz() (map[string]int, error) {
	stats := map[string]int{}

	stats["NrMut"] = gopie.NumberTotalMuts
	stats["NrMutInvalid"] = gopie.NumberInvalidMuts
	stats["ActiveReleased"] = baseA.ActiveReleased
	stats["AllActiveReleased"] = baseA.AllActiveReleased

	return stats, nil
}
