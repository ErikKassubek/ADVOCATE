// Copyright (c) 2025 Erik Kassubek
//
// File: exitCodes.go
// Brief: Check exit codes
//
// Author: Erik Kassubek
// Created: 2025-07-21
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/data"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// runAnalysisOnExitCodes checks the exit codes for the recording for actual bugs
//
// Parameter:
//   - all bool: If true, check for all, else only check for the once, that are not detected by the full analysis
func RunAnalysisOnExitCodes(all bool) {
	timer.Start(timer.AnaExitCode)
	defer timer.Stop(timer.AnaExitCode)

	switch data.ExitCode {
	case helper.ExitCodeCloseClose: // close on closed
		file, line, err := trace.PosFromPosString(data.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}

		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "CC",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, helper.ACloseOnClosed,
			"close", []results.ResultElem{arg1}, "", []results.ResultElem{})
		data.BugWasFound = true
	case helper.ExitCodeCloseNil: // close on nil
		file, line, err := trace.PosFromPosString(data.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "CC",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, helper.ACloseOnNilChannel,
			"close", []results.ResultElem{arg1}, "", []results.ResultElem{})
		data.BugWasFound = true
	case helper.ExitCodeNegativeWG: // negative wg counter
		file, line, err := trace.PosFromPosString(data.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "WD",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, helper.ANegWG,
			"done", []results.ResultElem{arg1}, "", []results.ResultElem{})
		data.BugWasFound = true
	case helper.ExitCodeUnlockBeforeLock: // unlock of not locked mutex
		file, line, err := trace.PosFromPosString(data.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "ML",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, helper.AUnlockOfNotLockedMutex,
			"done", []results.ResultElem{arg1}, "", []results.ResultElem{})
		data.BugWasFound = true
	case helper.ExitCodePanic: // unknown panic
		file, line, err := trace.PosFromPosString(data.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "XX",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, helper.RUnknownPanic,
			"panic", []results.ResultElem{arg1}, "", []results.ResultElem{})
		data.BugWasFound = true
	case helper.ExitCodeTimeout: // timeout
		results.Result(results.CRITICAL, helper.RTimeout,
			"", []results.ResultElem{}, "", []results.ResultElem{})
	}

	if all {
		if data.ExitCode == helper.ExitCodeSendClose { // send on closed
			file, line, err := trace.PosFromPosString(data.ExitPos)
			if err != nil {
				log.Error("Could not read exit pos: ", err)
			}
			arg1 := results.TraceElementResult{ // send
				RoutineID: 0,
				ObjID:     0,
				TPre:      0,
				ObjType:   "CS",
				File:      file,
				Line:      line,
			}
			results.Result(results.CRITICAL, helper.ASendOnClosed,
				"send", []results.ResultElem{arg1}, "", []results.ResultElem{})
			data.BugWasFound = true
		}
	}
}
