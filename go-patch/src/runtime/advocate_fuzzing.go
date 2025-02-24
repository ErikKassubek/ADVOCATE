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

const (
	selectPreferredTimeoutSec int64   = 2
	flowSleepTimeSec          float64 = 3
)

var (
	advocateFuzzingEnabled = false
	fuzzingSelectData      = make(map[string][]int)
	fuzzingSelectDataIndex = make(map[string]int)
	fuzzingFlowData        = make(map[string]int)
	fuzzingFLowCounter     = make(map[string]int)
)

func InitFuzzing(selectData map[string][]int, fuzzingFlow map[string]int) {
	fuzzingSelectData = selectData
	fuzzingFlowData = fuzzingFlow

	for key := range fuzzingSelectData {
		fuzzingSelectDataIndex[key] = 0
	}

	for key := range fuzzingFlowData {
		fuzzingFLowCounter[key] = 0
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
		return false, 0, selectPreferredTimeoutSec
	}

	_, file, line, _ := Caller(skip)
	if AdvocateIgnore(file) {
		return false, 0, selectPreferredTimeoutSec
	}
	key := file + ":" + intToString(line)

	if val, ok := fuzzingSelectData[key]; ok {
		index := fuzzingSelectDataIndex[key]
		if index >= len(val) {
			return false, 0, selectPreferredTimeoutSec
		}
		fuzzingSelectDataIndex[key]++
		return true, val[index], selectPreferredTimeoutSec
	}

	return false, 0, selectPreferredTimeoutSec
}

// currently used in once.Do, chan.send, chan.recv, mutex.(Try)Lock, rwmutex.(Try)(R)Lock
func FuzzingFlowWait(skip int) {
	if !advocateFuzzingEnabled {
		return
	}

	_, file, line, _ := Caller(skip)
	if AdvocateIgnore(file) {
		return
	}

	pos := file + ":" + intToString(line)

	if count, ok := fuzzingFlowData[pos]; ok {
		newCount := fuzzingFLowCounter[pos] + 1

		if count == newCount {
			sleep(flowSleepTimeSec)
		}

		fuzzingFLowCounter[pos] = newCount
	}
}
