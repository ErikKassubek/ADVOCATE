// Copyrigth (c) 2024 Erik Kassubek
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
	"analyzer/results"
	timemeasurement "analyzer/timeMeasurement"
	"log"
	"strconv"
	"strings"
)

/*
* Run the analysis
* MARK: run analysis
* Args:
*   assume_fifo (bool): True to assume fifo ordering in buffered channels
*   ignoreCriticalSections (bool): True to ignore critical sections when updating
*   	vector clocks
*   analysisCasesMap (map[string]bool): The analysis cases to run
*   fuzzing (bool): true if run with fuzzing
*   onlyAPanicAndLeak (bool): only test for actual panics and leaks
 */
func RunAnalysis(assumeFifo bool, ignoreCriticalSections bool, analysisCasesMap map[string]bool, fuzzing bool, onlyAPanicAndLeak bool) {
	if onlyAPanicAndLeak {
		runAnalysisOnExitCodes(true)
		checkForLeakSimple()
		return
	}

	RunFullAnalysis(assumeFifo, ignoreCriticalSections, analysisCasesMap, fuzzing)

	runAnalysisOnExitCodes(false)
}

/*
 * Check the exit codes for the recording for actual bugs
 * If all true, check for all, else only check for the once, that are not detected by the full analysis
 */
func runAnalysisOnExitCodes(all bool) {
	exit := strings.Split(exitPos, ":")
	if len(exit) != 2 {
		return
	}

	file := exit[0]
	line, err := strconv.Atoi(exit[1])
	if err != nil {
		return
	}

	if exitCode == 3 { // close on nil
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "CC",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, results.ACloseOnNilChannel,
			"close", []results.ResultElem{arg1}, "", []results.ResultElem{})
	} else if exitCode == 4 { // negative wg counter
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "WD",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, results.ANegWG,
			"done", []results.ResultElem{arg1}, "", []results.ResultElem{})
	} else if exitCode == 5 { // unlock of not locked mutex
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "ML",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, results.AUnlockOfNotLockedMutex,
			"done", []results.ResultElem{arg1}, "", []results.ResultElem{})
	} else if exitCode == 6 { // unknown panic
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "XX",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, results.AUnknownPanic,
			"panic", []results.ResultElem{arg1}, "", []results.ResultElem{})
	}

	if all {
		if exitCode == 1 { // send on closed
			arg1 := results.TraceElementResult{ // send
				RoutineID: 0,
				ObjID:     0,
				TPre:      0,
				ObjType:   "CS",
				File:      file,
				Line:      line,
			}
			results.Result(results.CRITICAL, results.ASendOnClosed,
				"send", []results.ResultElem{arg1}, "", []results.ResultElem{})
		}
	} else if exitCode == 2 { // close on closed
		arg1 := results.TraceElementResult{
			RoutineID: 0,
			ObjID:     0,
			TPre:      0,
			ObjType:   "CC",
			File:      file,
			Line:      line,
		}
		results.Result(results.CRITICAL, results.ACloseOnClosed,
			"close", []results.ResultElem{arg1}, "", []results.ResultElem{})
	}
}

/*
 * Check only for leaks
 * Do not check for potential partners
 * This is done by checking for each routine, if the last element is a blocking element
 * and if its tPost is 0
 */
func checkForLeakSimple() {
	elems := getLastElemPerRout()

	for _, elem := range elems {
		if elem.GetTPost() == 0 {
			leak(elem)
		}
	}
}

/*
* Run the full analysis
* Args:
*   assume_fifo (bool): True to assume fifo ordering in buffered channels
*   ignoreCriticalSections (bool): True to ignore critical sections when updating
*   	vector clocks
*   analysisCasesMap (map[string]bool): The analysis cases to run
*   fuzzing (bool): true if run with fuzzing
 */
func RunFullAnalysis(assumeFifo bool, ignoreCriticalSections bool, analysisCasesMap map[string]bool, fuzzing bool) {
	log.Print("Run full analysis on trace")

	fifo = assumeFifo
	runFuzzing = fuzzing

	analysisCases = analysisCasesMap
	InitAnalysis(analysisCases)

	if analysisCases["resourceDeadlock"] {
		ResetState()
	}

	for i := 1; i <= numberOfRoutines; i++ {
		currentVCHb[i] = clock.NewVectorClock(numberOfRoutines)
		currentVCWmhb[i] = clock.NewVectorClock(numberOfRoutines)
	}

	currentVCHb[1] = currentVCHb[1].Inc(1)
	currentVCWmhb[1] = currentVCWmhb[1].Inc(1)

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
		case *TraceElementNew:
			// do noting
		}

		if analysisCases["resourceDeadlock"] {
			switch e := elem.(type) {
			case *TraceElementMutex:
				HandleMutexEventForRessourceDeadlock(*e, currentVCWmhb[e.routine])
			}
		}

		// check for leak
		if analysisCases["leak"] && elem.GetTPost() == 0 {
			leak(elem)
		}

	}

	if analysisCases["selectWithoutPartner"] || runFuzzing {
		timemeasurement.Start("other")
		rerunCheckForSelectCaseWithoutPartnerChannel()
		CheckForSelectCaseWithoutPartner()
		timemeasurement.End("other")
	}

	if analysisCases["leak"] {
		timemeasurement.Start("leak")
		checkForLeak()
		checkForStuckRoutine()
		timemeasurement.End("leak")
	}

	if analysisCases["doneBeforeAdd"] {
		timemeasurement.Start("panic")
		checkForDoneBeforeAdd()
		timemeasurement.End("panic")
	}

	if analysisCases["cyclicDeadlock"] {
		timemeasurement.Start("other")
		checkForCyclicDeadlock()
		timemeasurement.End("other")
	}

	if analysisCases["resourceDeadlock"] {
		timemeasurement.Start("other")
		CheckForResourceDeadlock()
		timemeasurement.End("other")
	}

	if analysisCases["unlockBeforeLock"] {
		timemeasurement.Start("panic")
		checkForUnlockBeforeLock()
		timemeasurement.End("panic")
	}

	log.Print("Finished analyzing trace")
}

func leak(elem TraceElement) {
	timemeasurement.Start("leak")

	switch e := elem.(type) {
	case *TraceElementChannel:
		CheckForLeakChannelStuck(e, currentVCHb[e.routine])
	case *TraceElementMutex:
		CheckForLeakMutex(e)
	case *TraceElementWait:
		CheckForLeakWait(e)
	case *TraceElementSelect:
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
		CheckForLeakSelectStuck(e, ids, buffered, currentVCHb[e.routine], opTypes)
	case *TraceElementCond:
		CheckForLeakCond(e)
	}

	timemeasurement.End("leak")
}

/*
 * Rerun the CheckForSelectCaseWithoutPartnerChannel for all channel. This
 * is needed to find potential communication partners for not executed
 * select cases, if the select was executed after the channel
 */
func rerunCheckForSelectCaseWithoutPartnerChannel() {
	for _, trace := range traces {
		for _, elem := range trace {
			if e, ok := elem.(*TraceElementChannel); ok {
				CheckForSelectCaseWithoutPartnerChannel(e, e.GetVC(),
					e.Operation() == SendOp, e.IsBuffered())
			}
		}
	}
}
