// Copyright (c) 2024 Erik Kassubek
//
// File: waitGroup.go
// Brief: Rewrite for negative wait group counter
//
// Author: Erik Kassubek
// Created: 2024-04-05
//
// License: BSD-3-Clause

package rewriter

import (
	"analyzer/analysis"
	"analyzer/bugs"
	"analyzer/utils"
)

// Create a new trace for a negative wait group counter (done before add)
//
// Parameter:
//   - trace *analysis.Trace: Trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - expectedErrorCode int: For wg exitNegativeWG, for unlock before lock: exitUnlockBeforeLock
func rewriteGraph(trace *analysis.Trace, bug bugs.Bug, expectedErrorCode int) error {
	if bug.Type == utils.PNegWG {
		utils.LogInfo("Start rewriting trace for negative waitgroup counter...")
	} else if bug.Type == utils.PUnlockBeforeLock {
		utils.LogInfo("Start rewriting trace for unlock before lock...")
	}

	minTime := -1
	maxTime := -1

	for i := range bug.TraceElement2 {
		elem1 := bug.TraceElement1[i] // done/unlock

		trace.ShiftConcurrentOrAfterToAfter(elem1)

		if minTime == -1 || elem1.GetTPre() < minTime {
			minTime = elem1.GetTPre()
		}
		if maxTime == -1 || elem1.GetTPre() > maxTime {
			maxTime = elem1.GetTPre()
		}

	}

	// add start and end
	if !(minTime == -1 && maxTime == -1) {
		trace.AddTraceElementReplay(maxTime+1, expectedErrorCode)
	}

	return nil
}
