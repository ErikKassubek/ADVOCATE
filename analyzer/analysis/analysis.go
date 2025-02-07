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
	timemeasurement "analyzer/timeMeasurement"
	"log"
)

/*
* Calculate vector clocks
* MARK: run analysis
* Args:
*   assume_fifo (bool): True to assume fifo ordering in buffered channels
*   ignoreCriticalSections (bool): True to ignore critical sections when updating
*   	vector clocks
*   analysisCasesMap (map[string]bool): The analysis cases to run
*   fuzzing (bool): true if run with fuzzing
 */
func RunAnalysis(assumeFifo bool, ignoreCriticalSections bool, analysisCasesMap map[string]bool, fuzzing bool) string {

	log.Print("Analyze the trace")

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

	return result
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
