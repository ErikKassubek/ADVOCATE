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
	"strings"
)

var partialDeadlocks = make([]string, 0)

// DetectBlockingGC runs a partial deadlock detection in the current execution
// Parameter:
//   - interval bool: interval to run the detector. Set to 0 to run only once
func DetectBlockingGC() int {
	runtime.AdvocatePDDetectionStopped = false

	if runtime.AdvocatePDDetectionStopped {
		return 0
	}

	req := AdvocateRequest("Check blocking bug")
	println("RES:", req)

	res := runtime.AdvocateDetectBlocking()
	containsChan := false
	for _, r := range res {
		fields := strings.Split(r, "@")
		if len(fields) == 0 {
			continue
		}
		elems := strings.Split(fields[len(fields)-1], ":")
		if len(elems) == 0 {
			continue
		}
		if elems[0] == "chan" {
			containsChan = true
			break
		}
	}
	if len(res) != 0 {
		partialDeadlocks = append(partialDeadlocks, res...)
		if containsChan {
			return runtime.ExitCodeMixedDeadlock
		}
		return runtime.ExitCodeCyclic
	}

	return 0
}

// StopPartialDeadlockDetection stops the partial deadlock detection before
// the start of a new run
// If the detection is currently not run in a loop, this has no effect
func StopPartialDeadlockDetection() {
	runtime.AdvocatePDDetectionStopped = true
}
