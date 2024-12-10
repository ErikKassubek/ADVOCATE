// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_fuzzing.go
// Brief: Fuzzing
//
// Author: Erik Kassubek
// Created: 2024-12-10
//
// License: BSD-3-Clause


package runtime

const fuzzingSelectTimeoutSec = 4

var (
	advocateFuzzingEnabled = false
	fuzzingSelectData = make(map[string][]int)
	fuzzingSelectDataIndex = make(map[string]int)
)

func InitFuzzing(selectData map[string][]int) {
	fuzzingSelectData = selectData

	for key, _ := range fuzzingSelectData {
		fuzzingSelectDataIndex[key] = 0
	}

	advocateFuzzingEnabled = true
}

func isAdvocateFuzzingEnabled() {
	return advocateFuzzingEnabled
}

/*
 * Get the preferred case for the specified select
 * Args:
 *  skip for runtime.Caller
 * Returns:
 * 	bool: true if a preferred case exists, false otherwise
 * 	int: preferred case, -1 for default
 */
func AdvocateFuzzingGetPreferredCase(skip int) (bool, int) {
	if !advocateFuzzingEnabled {
		return false, 0
	}

	_, file, line, _ := Caller(skip)
	key = file + ":" + intToString(line)

	if val, ok := fuzzingSelectData[key]; ok {
		index = fuzzingSelectDataIndex[key]
		if index >= len(val) {
			return false, 0
		}
		fuzzingSelectDataIndex[key]++
		return val[index]
	}

	return false, 0
}