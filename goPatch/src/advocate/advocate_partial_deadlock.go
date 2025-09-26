// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_replay.go
// Brief: Entry point for partial deadlock detection
//
// Author: Erik Kassubek
// Created: 2025-07-23
//
// License: BSD-3-Clause

package advocate

import (
	"runtime"
	"time"
)

var partialDeadlocks = make([]string, 0)

// DetectPartialDeadlock runs a partial deadlock detection in the current execution
// Parameter:
//   - loop bool: if true, run a
func DetectPartialDeadlock(interval int) {
	runtime.AdvocatePDDetectionStopped = false

	go func() {
		for {
			if runtime.AdvocatePDDetectionStopped {
				return
			}

			res := runtime.AdvocateDetectPD()
			if len(res) != 0 {
				partialDeadlocks = append(partialDeadlocks, res...)
			}

			if interval <= 0 {
				return
			}

			time.Sleep(1 * time.Millisecond)
		}
	}()
}

// StopPartialDeadlockDetection stops the partial deadlock detection before
// the start of a new run
// If the detection is currently not run in a loop, this has no effect
func StopPartialDeadlockDetection() {
	runtime.AdvocatePDDetectionStopped = true
}
