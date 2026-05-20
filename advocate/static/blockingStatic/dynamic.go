// Copyright (c) 2026 Erik Kassubek
//
// File: dynamic.go
// Brief: For a specific run check for concurrency bugs
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package blockingStatic

import (
	"advocate/utils/log"
	"strings"
)

// Run the blocking bug analysis given a specific program run
//
// Parameter:
//   - data string: message received from runtime in form
//     "posForkBlockedRoutine~posForkRefRoutine~posBlockedOp"
//
// Returns:
//   - string: "1" if refRout contains correct operation to release blockedRout, "0" if it cannot release it, message for error
//
// Todo:
//   - add last called function in ref routine
func RunDynamicBlockingAnalysis(data string) string {
	fields := strings.Split(data, "~")
	if len(fields) != 3 {
		return "ERROR: incorrect number of fields in data " + data
	}

	blockedRoutinePos := fields[0]
	refRoutinePos := fields[1]
	blockedOp := fields[2]

	log.Debug("blockedRoutinePos: ", blockedRoutinePos)
	log.Debug("refRoutinePos: ", refRoutinePos)
	log.Debug("blockedOp: ", blockedOp)

	// TODO: implement check if

	return "RESULT OF STATIC ANALYSIS"
}
