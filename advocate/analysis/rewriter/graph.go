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
	"advocate/results/bugs"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
)

// Create a new trace for a negative wait group counter (done before add)
//
// Parameter:
//   - trace *analysis.Trace: Trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - expectedErrorCode int: For wg exitNegativeWG, for unlock before lock: exitUnlockBeforeLock
func rewriteGraph(tr *trace.Trace, bug bugs.Bug, expectedErrorCode int) error {
	if bug.Type == helper.PNegWG {
		log.Info("Start rewriting trace for negative waitgroup counter...")
	} else if bug.Type == helper.PUnlockBeforeLock {
		log.Info("Start rewriting trace for unlock before lock...")
	}

	minTime := -1
	maxTime := -1

	for i := range bug.TraceElement2 {
		elem1 := bug.TraceElement1[i] // done/unlock

		tr.ShiftConcurrentOrAfterToAfter(elem1)

		if minTime == -1 || elem1.GetTPre() < minTime {
			minTime = elem1.GetTPre()
		}
		if maxTime == -1 || elem1.GetTPre() > maxTime {
			maxTime = elem1.GetTPre()
		}

	}

	// add start and end
	if !(minTime == -1 && maxTime == -1) {
		tr.AddTraceElementReplay(maxTime+1, expectedErrorCode)
	}

	return nil
}
