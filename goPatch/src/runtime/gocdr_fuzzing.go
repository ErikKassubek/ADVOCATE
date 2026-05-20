// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_fuzzing.go
// Brief: Fuzzing
//
// Author: Erik Kassubek
// Created: 2024-12-10
//
// License: BSD-3-Clause

package runtime

const (
	selectPreferredTimeoutSec int64   = 1
	flowSleepTimeSec          float64 = 3
)

var (
	gocdrFuzzingEnabled       = false
	gocdrFuzzingDelayEnabled  = false
	gocdrFuzzingReplayEnabled = false
	fuzzingSelectData         = make(map[string][]int)
	fuzzingSelectDataIndex    = make(map[string]int)
	fuzzingFlowData           = make(map[string][]int)
	fuzzingFlowCounter        = make(map[string]int)
	fuzzingFlowDataCounter    = make(map[string]int)

	finishFuzzingFunc func()
)

// Init fuzzing based on simple delay of flow and preferred select cases
//
// Parameter:
//   - selectData map[string][]int: preferred select cases (select pos -> []preferred case index)
//   - fuzzingFlow map[string][]int: operations to delay (file -> []lines)
//   - finishFuzzing func(): function that should be called if fuzzing has finished
func InitFuzzingDelay(selectData map[string][]int, fuzzingFlow map[string][]int, finishFuzzing func()) {
	finishFuzzingFunc = finishFuzzing
	fuzzingSelectData = selectData
	fuzzingFlowData = fuzzingFlow

	for key := range fuzzingSelectData {
		fuzzingSelectDataIndex[key] = 0
	}

	for key := range fuzzingFlowData {
		fuzzingFlowCounter[key] = 0
		fuzzingFlowDataCounter[key] = 0
	}

	gocdrFuzzingEnabled = true
	gocdrFuzzingDelayEnabled = true
}

// InitFuzzingReplay initializes fuzzing based on full replay
//
// Parameter:
//   - finishFuzzing func(): gocdr.FinishFuzzing function
func InitFuzzingReplay(finishFuzzing func()) {
	finishFuzzingFunc = finishFuzzing
	gocdrFuzzingEnabled = true
	gocdrFuzzingReplayEnabled = true
}

// Get if fuzzing is enables
//
// Returns:
//   - bool: true if fuzzing is enabled, false otherwise
func IsGocdrFuzzingEnabled() bool {
	return gocdrFuzzingEnabled
}

// Get the preferred case for the specified select
//
// Parameter:
//   - skip int: skip for runtime.Caller
//
// Returns:
//   - bool: true if a preferred case exists, false otherwise
//   - int: preferred case, -1 for default
func GocdrFuzzingGetPreferredCase(skip int) (bool, int) {
	if !gocdrFuzzingEnabled {
		return false, 0
	}

	routine := GetReplayRoutineID()

	_, file, line, _ := Caller(skip)
	if GocdrIgnore(file) {
		return false, 0
	}
	key := BuildReplayKey(routine, file, line)

	if val, ok := fuzzingSelectData[key]; ok {
		index := fuzzingSelectDataIndex[key]
		if index >= len(val) {
			return false, 0
		}
		fuzzingSelectDataIndex[key]++
		return true, val[index]
	}

	return false, 0
}

// FuzzingFlowWait is called by the operations to check if they should wait for
// the delay based fuzzing.
// Currently used in once.Do, chan.send, chan.recv, mutex.(Try)Lock, rwmutex.(Try)(R)Lock
//
// Parameter:
//   - skip int: skip for runtime.Caller
func FuzzingFlowWait(skip int) {
	if !gocdrFuzzingDelayEnabled {
		return
	}

	routine := GetReplayRoutineID()

	_, file, line, _ := Caller(skip)
	if GocdrIgnore(file) {
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
