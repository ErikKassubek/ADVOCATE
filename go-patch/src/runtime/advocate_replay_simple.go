// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_replay_simple.go
// Brief: Functions for the wait based replay and replay with preferred select cases
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
	advocateReplaySimpleEnabled  = false
	advocateReplayDelayEnabled   = false
	advocateFuzzingReplayEnabled = false
	replaySelectData             = make(map[string][]int)
	replaySelectDataIndex        = make(map[string]int)
	replayFlowData               = make(map[string][]int)
	replayFlowCounter            = make(map[string]int)
	replayFlowDataCounter        = make(map[string]int)

	finishFuzzingFunc func()
)

// Init replay based on simple delay of flow and preferred select cases
//
// Parameter:
//   - selectData map[string][]int: preferred select cases (select pos -> []preferred case index)
//   - fuzzingFlow map[string][]int: operations to delay (file -> []lines)
func InitReplaySimple(selectData map[string][]int, fuzzingFlow map[string][]int) {
	currentReplayMode = replayModeSimple

	replaySelectData = selectData
	replayFlowData = fuzzingFlow

	for key := range replaySelectData {
		replaySelectDataIndex[key] = 0
	}

	for key := range replayFlowData {
		replayFlowCounter[key] = 0
		replayFlowDataCounter[key] = 0
	}

	advocateReplaySimpleEnabled = true
	advocateReplayDelayEnabled = true
}

// Get if fuzzing is enables
//
// Returns:
//   - bool: true if fuzzing is enabled, false otherwise
func IsAdvocateFuzzingEnabled() bool {
	return advocateReplaySimpleEnabled
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
func AdvocateGetPreferredCase(skip int) (bool, int, int64) {
	if !advocateReplaySimpleEnabled {
		return false, 0, selectPreferredTimeoutSec
	}

	routine := GetRoutineID()

	_, file, line, _ := Caller(skip)
	if AdvocateIgnore(file) {
		return false, 0, selectPreferredTimeoutSec
	}
	key := buildReplayKey(routine, file, line)

	if val, ok := replaySelectData[key]; ok {
		index := replaySelectDataIndex[key]
		if index >= len(val) {
			return false, 0, selectPreferredTimeoutSec
		}
		replaySelectDataIndex[key]++
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
	if !advocateReplayDelayEnabled {
		return
	}

	routine := GetRoutineID()

	_, file, line, _ := Caller(skip)
	if AdvocateIgnore(file) {
		return
	}

	pos := buildReplayKey(routine, file, line)

	if countList, ok := replayFlowData[pos]; ok {
		dataCounter := replayFlowDataCounter[pos]
		if dataCounter >= len(countList) {
			return
		}

		count := countList[dataCounter]

		newCount := replayFlowCounter[pos] + 1

		if count == newCount {
			replayFlowDataCounter[pos] = replayFlowDataCounter[pos] + 1
			sleep(flowSleepTimeSec)
		}

		replayFlowCounter[pos] = newCount
	}
}
