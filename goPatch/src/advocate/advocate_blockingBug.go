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

// DetectBlockingGC runs a partial deadlock detection in the current execution
// Parameter:
//   - interval bool: interval to run the detector. Set to 0 to run only once
func DetectBlockingGC(interval int) {
	runtime.AdvocatePDDetectionStopped = false

	go func() {
		for {
			if runtime.AdvocatePDDetectionStopped {
				return
			}

			res := runtime.AdvocateDetectBlocking()
			if len(res) != 0 {
				partialDeadlocks = append(partialDeadlocks, res...)
			}

			if interval <= 0 {
				return
			}

			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
}

// StopPartialDeadlockDetection stops the partial deadlock detection before
// the start of a new run
// If the detection is currently not run in a loop, this has no effect
func StopPartialDeadlockDetection() {
	runtime.AdvocatePDDetectionStopped = true
}
