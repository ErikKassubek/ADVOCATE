// Copyright (c) 2025 Erik Kassubek
//
// File: vc.go
// Brief: Data required for how the analysis is run
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package data

var (
	Fifo          bool
	ModeIsFuzzing bool

	// analysis cases to run
	AnalysisCases       = make(map[string]bool)
	AnalysisFuzzingFlow = false

	// exit code info
	ExitCode int
	ExitPos  string

	BugWasFound = false

	// replay timeout info
	replayTimeoutOldest   int
	replayTimeoutDisabled int
	replayTimeoutAck      int

	durationInSeconds = -1 // the duration of the recording in seconds
)

// SetExitInfo stores the exit code and exit position of a run
//
// Parameter:
//   - code int: the exit code
//   - pos string: the exit position
func SetExitInfo(code int, pos string) {
	ExitCode = code
	ExitPos = pos
}

// SetReplayTimeoutInfo stores information about wether a run that was guided
// by replay (especially in GoPie fuzzing) had a timeout
//
// Parameter:
//
//   - oldest int: the timer when the the replay released the oldest waiting
//
//     or the current next for the first time, if never it should be 0
//
//   - disabled int: the timer when the the replay was so stuck, that the
//     replay had to be disabled for the first time, if never it should be 0
//
//   - ack int: the timer when the the replay timed out on an acknowledgement,
//     if never it should be 0
func SetReplayTimeoutInfo(oldest, disabled, ack int) {
	replayTimeoutOldest = oldest
	replayTimeoutDisabled = disabled
	replayTimeoutAck = ack
}

// GetTimeoutHappened return if any kind of timeout happened
// A timeout happened if at least one of the three timeout var is not 0
//
// Returns:
//   - - bool: true if a timeout happened, false otherwise
func GetTimeoutHappened() bool {
	return (replayTimeoutOldest + replayTimeoutDisabled + replayTimeoutAck) != 0
}

// SetRuntimeDurationSec is a setter for durationInSeconds
//
// Parameter:
//   - sec int: the runtime duration of a run in second
func SetRuntimeDurationSec(sec int) {
	durationInSeconds = sec
}

// GetRuntimeDurationInSec is a getter for durationInSeconds
//
// Returns:
//   - int: the runtime duration of a run in second
func GetRuntimeDurationInSec() int {
	return durationInSeconds
}
