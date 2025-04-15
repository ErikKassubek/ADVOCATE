// Copyright (c) 2024 Erik Kassubek
//
// File: analysis.go
// Brief: analysis of traces if performed from here
//
// Author: Erik Kassubek, Sebastian Pohsner
// Created: 2025-01-01
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/memory"
	"analyzer/results"
	"analyzer/timer"
	"analyzer/utils"
)

// RunAnalysis starts the analysis of the main trace
//
// Parameter:
//   - assume_fifo bool: True to assume fifo ordering in buffered channels
//   - ignoreCriticalSections bool: True to ignore critical sections when updating vector clocks
//   - analysisCasesMap map[string]bool: The analysis cases to run
//   - fuzzing bool: true if run with fuzzing
//   - onlyAPanicAndLeak bool: only test for actual panics and leaks
func RunAnalysis(assumeFifo bool, ignoreCriticalSections bool, analysisCasesMap map[string]bool, fuzzing bool, onlyAPanicAndLeak bool) {
	// catch panics in analysis.
	// Prevents the whole toolchain to panic if one analysis panics
	defer func() {
		if r := recover(); r != nil {
			memory.Cancel()
			utils.LogError(r)
		}
	}()

	timer.Start(timer.Analysis)
	defer timer.Stop(timer.Analysis)

	if onlyAPanicAndLeak {
		runAnalysisOnExitCodes(true)
		checkForLeakSimple()
		return
	}

	runAnalysisOnExitCodes(fuzzing)
	RunHBAnalysis(assumeFifo, ignoreCriticalSections, analysisCasesMap, fuzzing)
}

// runAnalysisOnExitCodes checks the exit codes for the recording for actual bugs
//
// Parameter:
//   - all bool: If true, check for all, else only check for the once, that are not detected by the full analysis
func runAnalysisOnExitCodes(all bool) {
	timer.Start(timer.AnaExitCode)
	defer timer.Stop(timer.AnaExitCode)

	file, line, err := posFromPosString(exitPos)
	if err != nil {
		return
	}

	switch exitCode {
	case utils.ExitCodeCloseClose: // close on closed
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "CC",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, utils.ACloseOnClosed,
			"close", []results.ResultElem{arg1}, "", []results.ResultElem{})
	case utils.ExitCodeCloseNil: // close on nil
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "CC",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, utils.ACloseOnNilChannel,
			"close", []results.ResultElem{arg1}, "", []results.ResultElem{})
	case utils.ExitCodeNegativeWG: // negative wg counter
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "WD",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, utils.ANegWG,
			"done", []results.ResultElem{arg1}, "", []results.ResultElem{})
	case utils.ExitCodeUnlockBeforeLock: // unlock of not locked mutex
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "ML",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, utils.AUnlockOfNotLockedMutex,
			"done", []results.ResultElem{arg1}, "", []results.ResultElem{})
	case utils.ExitCodePanic: // unknown panic
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "XX",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, utils.RUnknownPanic,
			"panic", []results.ResultElem{arg1}, "", []results.ResultElem{})
	case 7: // timeout
		results.Result(results.CRITICAL, utils.RTimeout,
			"", []results.ResultElem{}, "", []results.ResultElem{})
	}

	if all {
		if exitCode == utils.ExitCodeSendClose { // send on closed
			arg1 := results.TraceElementResult{ // send
				RoutineID: 0,
				ObjID:     0,
				TPre:      0,
				ObjType:   "CS",
				File:      file,
				Line:      line,
			}
			results.Result(results.CRITICAL, utils.ASendOnClosed,
				"send", []results.ResultElem{arg1}, "", []results.ResultElem{})
		}
	}
}

// checkForLeakSimple check only for leaks without using the hb info
// Do not check for potential partners
// This is done by checking for each routine, if the last element is a blocking element
// and if its tPost is 0
func checkForLeakSimple() {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	elems := getLastElemPerRout()

	for _, elem := range elems {
		if elem.GetTPost() == 0 {
			checkLeak(elem)
		}
	}
}

// RunHBAnalysis runs the full analysis happens before based analysis
//
// Parameter:
//   - assume_fifo bool: True to assume fifo ordering in buffered channels
//   - ignoreCriticalSections bool: True to ignore critical sections when updating vector clocks
//   - analysisCasesMap map[string]bool: The analysis cases to run
//   - fuzzing bool: true if run with fuzzing
func RunHBAnalysis(assumeFifo bool, ignoreCriticalSections bool, analysisCasesMap map[string]bool, fuzzing bool) {
	fifo = assumeFifo
	modeIsFuzzing = fuzzing

	analysisCases = analysisCasesMap
	InitAnalysis(analysisCases, fuzzing)

	if analysisCases["resourceDeadlock"] {
		ResetState()
	}

	for i := 1; i <= MainTrace.numberOfRoutines; i++ {
		currentVC[i] = clock.NewVectorClock(MainTrace.numberOfRoutines)
		currentWVC[i] = clock.NewVectorClock(MainTrace.numberOfRoutines)
	}

	currentVC[1].Inc(1)
	currentWVC[1].Inc(1)

	utils.LogInfo("Start HB analysis")

	for elem := getNextElement(); elem != nil; elem = getNextElement() {
		switch e := elem.(type) {
		case *TraceElementAtomic:
			if ignoreCriticalSections {
				e.updateVectorClockAlt()
			} else {
				e.updateVectorClock()
			}
		case *TraceElementChannel:
			e.updateVectorClock()
		case *TraceElementMutex:
			if ignoreCriticalSections {
				e.updateVectorClockAlt()
			} else {
				e.updateVectorClock()
			}
			if analysisFuzzing {
				getConcurrentMutexForFuzzing(e)
			}
		case *TraceElementFork:
			e.updateVectorClock()
		case *TraceElementSelect:
			cases := e.GetCases()
			ids := make([]int, 0)
			opTypes := make([]int, 0)
			for _, c := range cases {
				switch c.opC {
				case SendOp:
					ids = append(ids, c.GetID())
					opTypes = append(opTypes, 0)
				case RecvOp:
					ids = append(ids, c.GetID())
					opTypes = append(opTypes, 1)
				}
			}
			e.updateVectorClock()
		case *TraceElementWait:
			e.updateVectorClock()
		case *TraceElementCond:
			e.updateVectorClock()
		case *TraceElementOnce:
			e.updateVectorClock()
			if analysisFuzzing {
				getConcurrentOnceForFuzzing(e)
			}
		case *TraceElementNew:
			// do noting
		}

		if analysisCases["resourceDeadlock"] {
			switch e := elem.(type) {
			case *TraceElementMutex:
				HandleMutexEventForRessourceDeadlock(*e)
			}
		}

		// check for leak
		if analysisCases["leak"] && elem.GetTPost() == 0 {
			checkLeak(elem)
		}

		if memory.WasCanceled() {
			return
		}
	}

	MainTrace.hbWasCalc = true

	utils.LogInfo("Finished HB analysis")

	if analysisCases["selectWithoutPartner"] || modeIsFuzzing {
		rerunCheckForSelectCaseWithoutPartnerChannel()
		CheckForSelectCaseWithoutPartner()
	}

	if memory.WasCanceled() {
		return
	}

	if analysisCases["leak"] {
		utils.LogInfo("Check for leak")
		checkForLeak()
		checkForStuckRoutine()
		utils.LogInfo("Finish check for leak")
	}

	if memory.WasCanceled() {
		return
	}

	if analysisCases["doneBeforeAdd"] {
		utils.LogInfo("Check for done before add")
		checkForDoneBeforeAdd()
		utils.LogInfo("Finish check for done before add")
	}

	if memory.WasCanceled() {
		return
	}

	// if memory.WasCanceled() {
	// 	return
	// }

	if analysisCases["resourceDeadlock"] {
		utils.LogInfo("Check for cyclic deadlock")
		CheckForResourceDeadlock()
		utils.LogInfo("Finish check for cyclic deadlock")
	}

	if memory.WasCanceled() {
		return
	}

	if analysisCases["unlockBeforeLock"] {
		utils.LogInfo("Check for unlock before lock")
		checkForUnlockBeforeLock()
		utils.LogInfo("Finish check for unlock before lock")
	}
}

// checkLeak checks for a given element if it leaked (has no tPost). If so,
// it will look for a possible way to resolve the leak
//
// Parameter:
//   - elem TraceElement: Element to check
func checkLeak(elem TraceElement) {
	switch e := elem.(type) {
	case *TraceElementChannel:
		CheckForLeakChannelStuck(e, currentVC[e.routine])
	case *TraceElementMutex:
		CheckForLeakMutex(e)
	case *TraceElementWait:
		CheckForLeakWait(e)
	case *TraceElementSelect:
		timer.Start(timer.AnaLeak)
		cases := e.GetCases()
		ids := make([]int, 0)
		buffered := make([]bool, 0)
		opTypes := make([]int, 0)
		for _, c := range cases {
			switch c.opC {
			case SendOp:
				ids = append(ids, c.GetID())
				opTypes = append(opTypes, 0)
				buffered = append(buffered, c.IsBuffered())
			case RecvOp:
				ids = append(ids, c.GetID())
				opTypes = append(opTypes, 1)
				buffered = append(buffered, c.IsBuffered())
			}
		}
		timer.Stop(timer.AnaLeak)
		CheckForLeakSelectStuck(e, ids, buffered, currentVC[e.routine], opTypes)
	case *TraceElementCond:
		CheckForLeakCond(e)
	}
}
