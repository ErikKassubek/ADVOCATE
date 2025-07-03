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
	"advocate/analysis/clock"
	"advocate/analysis/concurrent"
	"advocate/analysis/data"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/memory"
	"advocate/utils/timer"
	"time"
)

// RunAnalysis starts the analysis of the main trace
//
// Parameter:
//   - assume_fifo bool: True to assume fifo ordering in buffered channels
//   - ignoreCriticalSections bool: True to ignore critical sections when updating vector clocks
//   - analysisCasesMap map[string]bool: The analysis cases to run
//   - fuzzing bool: true if run with fuzzing
//   - onlyAPanicAndLeak bool: only test for actual panics and leaks
func RunAnalysis(assumeFifo bool, ignoreCriticalSections bool,
	analysisCasesMap map[string]bool, fuzzing bool, onlyAPanicAndLeak bool) {
	// catch panics in analysis.
	// Prevents the whole toolchain to panic if one analysis panics
	if log.IsPanicPrevent() {
		defer func() {
			if r := recover(); r != nil {
				memory.Cancel()
				log.Error(r)
			}
		}()
	}

	timer.Start(timer.Analysis)
	defer timer.Stop(timer.Analysis)

	// if onlyAPanicAndLeak {
	// 	runAnalysisOnExitCodes(true)
	// 	checkForStuckRoutine(true)
	// 	return
	// }

	runAnalysisOnExitCodes(fuzzing)
	RunHBAnalysis(assumeFifo, ignoreCriticalSections, analysisCasesMap, fuzzing)

	start := time.Now()
	i := 0
	for _, trace := range data.MainTrace.GetTraces() {
		i++
		if len(trace) == 0 {
			continue
		}

		index := int(len(trace) / 2)
		elem := trace[index]
		res := concurrent.GetConcurrentPartialOrderGraph(elem, true)
		log.Infof("%d/%d: %d", i, data.MainTrace.GetNoRoutines(), len(res))

	}
	dur := time.Since(start)
	log.Importantf("Graph: %s", dur.String())

	start = time.Now()
	i = 0
	for _, trace := range data.MainTrace.GetTraces() {
		i++
		if len(trace) == 0 {
			continue
		}

		index := int(len(trace) / 2)
		elem := trace[index]
		res := concurrent.GetConcurrentBruteForce(elem, true)
		log.Infof("%d/%d: %d", i, data.MainTrace.GetNoRoutines(), len(res))

	}
	dur = time.Since(start)
	log.Importantf("Brute Force: %s", dur.String())
	panic("A")
}

// runAnalysisOnExitCodes checks the exit codes for the recording for actual bugs
//
// Parameter:
//   - all bool: If true, check for all, else only check for the once, that are not detected by the full analysis
func runAnalysisOnExitCodes(all bool) {
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

// RunHBAnalysis runs the full analysis happens before based analysis
//
// Parameter:
//   - assume_fifo bool: True to assume fifo ordering in buffered channels
//   - ignoreCriticalSections bool: True to ignore critical sections when updating vector clocks
//   - data.AnalysisCasesMap map[string]bool: The analysis cases to run
//   - fuzzing bool: true if run with fuzzing
//
// Returns:
//   - bool: true if something has been found
func RunHBAnalysis(assumeFifo bool, ignoreCriticalSections bool,
	analysisCasesMap map[string]bool, fuzzing bool) {
	data.Fifo = assumeFifo
	data.ModeIsFuzzing = fuzzing

	data.AnalysisCases = analysisCasesMap
	data.InitAnalysisData(data.AnalysisCases, fuzzing)

	if data.AnalysisCases["resourceDeadlock"] {
		ResetState()
	}

	noRoutine := data.MainTrace.GetNoRoutines()
	for i := 1; i <= noRoutine; i++ {
		data.CurrentVC[i] = clock.NewVectorClock(noRoutine)
		data.CurrentWVC[i] = clock.NewVectorClock(noRoutine)
	}

	data.CurrentVC[1].Inc(1)
	data.CurrentWVC[1].Inc(1)

	log.Info("Start HB analysis")

	traceIter := data.MainTrace.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		// add edge between element of same routine to partial order trace
		concurrent.AddEdgePartialOrderGraph(elem)

		switch e := elem.(type) {
		case *trace.TraceElementAtomic:
			if ignoreCriticalSections {
				UpdateVCAtomicAlt(e)
			} else {
				UpdateVCAtomic(e)
			}
		case *trace.TraceElementChannel:
			UpdateVCChannel(e)
		case *trace.TraceElementMutex:
			if ignoreCriticalSections {
				UpdateVCMutexAlt(e)
			} else {
				UpdateVCMutex(e)
			}
			if data.AnalysisFuzzing {
				getConcurrentMutexForFuzzing(e)
			}
		case *trace.TraceElementFork:
			UpdateVCFork(e)
		case *trace.TraceElementSelect:
			cases := e.GetCases()
			ids := make([]int, 0)
			opTypes := make([]int, 0)
			for _, c := range cases {
				switch c.GetOpC() {
				case trace.SendOp:
					ids = append(ids, c.GetID())
					opTypes = append(opTypes, 0)
				case trace.RecvOp:
					ids = append(ids, c.GetID())
					opTypes = append(opTypes, 1)
				}
			}
			UpdateVCSelect(e)
		case *trace.TraceElementWait:
			UpdateVCWait(e)
		case *trace.TraceElementCond:
			UpdateVCCond(e)
		case *trace.TraceElementOnce:
			UpdateVCOnce(e)
			if data.AnalysisFuzzing {
				getConcurrentOnceForFuzzing(e)
			}
		case *trace.TraceElementRoutineEnd:
			UpdateVCRoutineEnd(e)
		case *trace.TraceElementNew:
			UpdateVCNew(e)
		}

		if data.AnalysisCases["resourceDeadlock"] {
			switch e := elem.(type) {
			case *trace.TraceElementMutex:
				HandleMutexEventForRessourceDeadlock(*e)
			}
		}

		// check for leak
		if data.AnalysisCases["leak"] && elem.GetTPost() == 0 {
			checkLeak(elem)
		}

		if memory.WasCanceled() {
			return
		}
	}

	data.MainTrace.SetHBWasCalc(true)

	log.Info("Finished HB analysis")

	if data.ModeIsFuzzing {
		rerunCheckForSelectCaseWithPartnerChannel()
		CheckForSelectCaseWithPartner()
	}

	if memory.WasCanceled() {
		return
	}

	if data.AnalysisCases["leak"] {
		log.Info("Check for leak")
		checkForLeak()
		checkForStuckRoutine(false)
		log.Info("Finish check for leak")
	}

	if memory.WasCanceled() {
		return
	}

	if data.AnalysisCases["doneBeforeAdd"] {
		log.Info("Check for done before add")
		checkForDoneBeforeAdd()
		log.Info("Finish check for done before add")
	}

	if memory.WasCanceled() {
		return
	}

	// if memory.WasCanceled() {
	// 	return
	// }

	if data.AnalysisCases["resourceDeadlock"] {
		log.Info("Check for cyclic deadlock")
		CheckForResourceDeadlock()
		log.Info("Finish check for cyclic deadlock")
	}

	if memory.WasCanceled() {
		return
	}

	if data.AnalysisCases["unlockBeforeLock"] {
		log.Info("Check for unlock before lock")
		checkForUnlockBeforeLock()
		log.Info("Finish check for unlock before lock")
	}
}

// checkLeak checks for a given element if it leaked (has no tPost). If so,
// it will look for a possible way to resolve the leak
//
// Parameter:
//   - elem TraceElement: Element to check
func checkLeak(elem trace.TraceElement) {
	switch e := elem.(type) {
	case *trace.TraceElementChannel:
		CheckForLeakChannelStuck(e, data.CurrentVC[e.GetRoutine()])
	case *trace.TraceElementMutex:
		CheckForLeakMutex(e)
	case *trace.TraceElementWait:
		CheckForLeakWait(e)
	case *trace.TraceElementSelect:
		timer.Start(timer.AnaLeak)
		cases := e.GetCases()
		ids := make([]int, 0)
		buffered := make([]bool, 0)
		opTypes := make([]int, 0)
		for _, c := range cases {
			switch c.GetOpC() {
			case trace.SendOp:
				ids = append(ids, c.GetID())
				opTypes = append(opTypes, 0)
				buffered = append(buffered, c.IsBuffered())
			case trace.RecvOp:
				ids = append(ids, c.GetID())
				opTypes = append(opTypes, 1)
				buffered = append(buffered, c.IsBuffered())
			}
		}
		timer.Stop(timer.AnaLeak)
		CheckForLeakSelectStuck(e, ids, buffered, data.CurrentVC[e.GetRoutine()], opTypes)
	case *trace.TraceElementCond:
		CheckForLeakCond(e)
	}
}
