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

var fuzzStats = []statsType{
	testName,
	nrMut,
	nrMutInvalid,
	activeReleased,
	allActiveReleased,
}

var fuzzStatsStr = []string{
	string(testName),
	string(nrMut),
	string(nrMutInvalid),
	string(activeReleased),
	string(allActiveReleased),
}

// Collect stats about each fuzzing run
//
// Returns:
//   - map[string]int: map with the stats
//   - error
func statsFuzz() (map[statsType]int, error) {
	stats := map[statsType]int{}

	stats[nrMut] = gopie.NumberTotalMuts
	stats[nrMutInvalid] = gopie.NumberInvalidMuts
	stats[activeReleased] = baseA.ActiveReleased
	stats[allActiveReleased] = baseA.AllActiveReleased

	return stats, nil
}
