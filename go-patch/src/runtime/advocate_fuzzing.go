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
	advocateFuzzingEnabled       = false
	advocateFuzzingDelayEnabled  = false
	advocateFuzzingReplayEnabled = false
	fuzzingSelectData            = make(map[string][]int)
	fuzzingSelectDataIndex       = make(map[string]int)
	fuzzingFlowData              = make(map[string][]int)
	fuzzingFlowCounter           = make(map[string]int)
	fuzzingFlowDataCounter       = make(map[string]int)

	finishFuzzingFunc func()
)

// Init fuzzing based on simple delay of flow and preferred select cases
//
// Parameter:
//   - selectData map[string][]int: preferred select cases (select pos -> []preferred case index)
//   - fuzzingFlow map[string][]int: operations to delay (file -> []lines)
func InitFuzzingDelay(selectData map[string][]int, fuzzingFlow map[string][]int) {
	fuzzingSelectData = selectData
	fuzzingFlowData = fuzzingFlow

	for key := range fuzzingSelectData {
		fuzzingSelectDataIndex[key] = 0
	}

	for key := range fuzzingFlowData {
		fuzzingFlowCounter[key] = 0
		fuzzingFlowDataCounter[key] = 0
	}

	advocateFuzzingEnabled = true
	advocateFuzzingDelayEnabled = true
}

// InitFuzzingReplay initializes fuzzing based on full replay
//
// Parameter:
//   - finishFuzzing func(): advocate.FinishFuzzing function
func InitFuzzingReplay(finishFuzzing func()) {
	finishFuzzingFunc = finishFuzzing
	advocateFuzzingEnabled = true
	advocateFuzzingReplayEnabled = true
}

// Get if fuzzing is enables
//
// Returns:
//   - bool: true if fuzzing is enabled, false otherwise
func IsAdvocateFuzzingEnabled() bool {
	return advocateFuzzingEnabled
}

// Get the preferred case for the specified select
//
// Parameter:
//   - skip int: skip for runtime.Caller
//
// Returns:
//   - bool: true if a preferred case exists, false otherwise
//   - int: preferred case, -1 for default
//   - int64: fuzzing timeout in seconds
func AdvocateFuzzingGetPreferredCase(skip int) (bool, int, int64) {
	if !advocateFuzzingEnabled {
		return false, 0, selectPreferredTimeoutSec
	}

	routine := GetRoutineID()

	_, file, line, _ := Caller(skip)
	if AdvocateIgnore(file) {
		return false, 0, selectPreferredTimeoutSec
	}
	key := BuildReplayKey(routine, file, line)

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

// FuzzingFlowWait is called by the operations to check if they should wait for
// the delay based fuzzing.
// Currently used in once.Do, chan.send, chan.recv, mutex.(Try)Lock, rwmutex.(Try)(R)Lock
//
// Parameter:
//   - skip int: skip for runtime.Caller
func FuzzingFlowWait(skip int) {
	if !advocateFuzzingDelayEnabled {
		return
	}

	routine := GetRoutineID()

	_, file, line, _ := Caller(skip)
	if AdvocateIgnore(file) {
		return
	}

	pos := BuildReplayKey(routine, file, line)

	if countList, ok := fuzzingFlowData[pos]; ok {
		dataCounter := fuzzingFlowDataCounter[pos]
		if dataCounter >= len(countList) {
			return
		}

		count := countList[dataCounter]

		newCount := fuzzingFlowCounter[pos] + 1

		if count == newCount {
			fuzzingFlowDataCounter[pos] = fuzzingFlowDataCounter[pos] + 1
			sleep(flowSleepTimeSec)
		}

		fuzzingFlowCounter[pos] = newCount
	}
}
