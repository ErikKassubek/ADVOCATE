// Copyright (c) 2025 Erik Kassubek
//
// File: timeMeasurement.go
// Brief: The actual timer
//
// Author: Erik Kassubek
// Created: 2025-02-25
//
// License: BSD-3-Clause

package timer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const numberTimer = 20

const (
	Total                int = iota // total runtime of everything
	TotalTest                       // total runtime for each test
	Run                             // runtime for base run without recording or replay
	Recording                       // runtime for recording
	Io                              // time spend reading and writing files (not including the time file)
	Analysis                        // total runtime for analysis
	AnaHb                           // runtime to update vc
	AnaExitCode                     // runtime for analyzing the exit code of the run
	AnaLeak                         // runtime to check for leaks
	AnaClose                        // runtime or checking send/recv on closed channel
	AnaConcurrent                   // runtime for checking concurrent recv
	AnaResource                     // runtime for finding cyclic deadlocks
	AnaSelWithoutPartner            // runtime for select cases without partner
	AnaUnlock                       // runtime for unlock before lock
	AnaWait                         // runtime for negative wait group counter
	AnaMixed                        // runtime for mixed deadlock
	FuzzingAna                      // runtime for getting info for fuzzing
	FuzzingMut                      // runtime for creation of mutations
	Rewrite                         // runtime for rewrite
	Replay                          // runtime for replay
)

var (
	timer       = make([]Timer, numberTimer)
	measureTime = false
)

// Init time measurement
//
// Parameter:
//   - mt (bool): if true, time is print into a time file
//   - progPath (string):  path to the result folder
func Init(mt bool, progPath string) {
	measureTime = mt

	home, _ := os.UserHomeDir()
	resultFolder = filepath.Join(strings.Replace(progPath, "~", home, -1), "advocateResult")

	for i := range numberTimer {
		timer[i] = Timer{}
	}
}

// Start a specified timer
//
// Parameter:
//   - t int: the timer to start
func Start(t int) {
	timer[t].Start()
}

// Stop a specified timer
//
// Parameter:
//   - t int: the timer to stop
func Stop(t int) {
	timer[t].Stop()
}

// Get the current time from a specified counter
//
// Parameter:
//   - t int: the timer to start
//
// Returns:
//   - time.Duration: the current time of the specified counter
func GetTime(t int) time.Duration {
	return timer[t].GetTime()
}

// Reset all counter to zero
func ResetAll() {
	for i := range numberTimer {
		timer[i].Reset()
	}
}

// Reset all counter to zero that are test specific
func ResetTest() {
	for i := 1; i < numberTimer; i++ {
		timer[i].Reset()
	}
}

// Reset all counter to zero that are specific for each fuzzing run
func ResetFuzzing() {
	for i := 3; i < numberTimer; i++ {
		timer[i].Reset()
	}
}

// Get a string representation of test specific timer values
//
// Returns:
//   - string: representation of test specific timer values
func ToString() string {
	res := ""

	for i := 2; i < numberTimer; i++ {
		if res != "" {
			res += ","
		}
		res += fmt.Sprintf("%.5f", timer[i].GetTime().Seconds())
	}

	return res
}
