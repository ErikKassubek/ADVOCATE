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

const selectPreferredTimeout int64 = 2

var (
	advocateFuzzingEnabled = false
	fuzzingSelectData      = make(map[string][]int)
	fuzzingSelectDataIndex = make(map[string]int)
	fuzzingFlowData        = make(map[string]int)
)

func InitFuzzing(selectData map[string][]int, fuzzingFlow map[string]int) {
	fuzzingSelectData = selectData
	fuzzingFlowData = fuzzingFlow

	for key := range fuzzingSelectData {
		fuzzingSelectDataIndex[key] = 0
	}

	advocateFuzzingEnabled = true
}

func isAdvocateFuzzingEnabled() bool {
	return advocateFuzzingEnabled
}

/*
 * Get the preferred case for the specified select
 * Args:
 *  skip for runtime.Caller
 * Returns:
 * 	bool: true if a preferred case exists, false otherwise
 * 	int: preferred case, -1 for default
 * 	int64: fuzzing timeout in seconds
 */
func AdvocateFuzzingGetPreferredCase(skip int) (bool, int, int64) {
	if !advocateFuzzingEnabled {
		return false, 0, selectPreferredTimeout
	}

	_, file, line, _ := Caller(skip)
	if AdvocateIgnore(file) {
		return false, 0, selectPreferredTimeout
	}
	key := file + ":" + intToString(line)

	if val, ok := fuzzingSelectData[key]; ok {
		index := fuzzingSelectDataIndex[key]
		if index >= len(val) {
			return false, 0, selectPreferredTimeout
		}
		fuzzingSelectDataIndex[key]++
		return true, val[index], selectPreferredTimeout
	}

	return false, 0, selectPreferredTimeout
}
