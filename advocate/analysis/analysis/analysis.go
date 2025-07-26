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
	"advocate/analysis/analysis/elements"
	"advocate/analysis/analysis/scenarios"
	"advocate/analysis/data"
	"advocate/analysis/hb/cssts"
	"advocate/analysis/hb/hbCalc"
	hb "advocate/analysis/hb/hbCalc"
	"advocate/analysis/hb/pog"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/memory"
	"advocate/utils/timer"
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

	scenarios.RunAnalysisOnExitCodes(true)

	if onlyAPanicAndLeak {
		scenarios.CheckForStuckRoutine(true)

		if !fuzzing {
			return
		}
	}

	RunHBAnalysis(assumeFifo, ignoreCriticalSections, analysisCasesMap, fuzzing, !onlyAPanicAndLeak)
}

// RunHBAnalysis runs the full analysis happens before based analysis
//
// Parameter:
//   - assume_fifo bool: True to assume fifo ordering in buffered channels
//   - ignoreCriticalSections bool: True to ignore critical sections when updating vector clocks
//   - data.AnalysisCasesMap map[string]bool: The analysis cases to run
//   - fuzzing bool: true if run with fuzzing
//   - runAna bool: true to run the predictive analysis
//
// Returns:
//   - bool: true if something has been found
func RunHBAnalysis(assumeFifo bool, ignoreCriticalSections bool,
	analysisCasesMap map[string]bool, fuzzing bool, runAna bool) {
	log.Info("Start analysis")

	data.Fifo = assumeFifo
	data.ModeIsFuzzing = fuzzing

	// set which hb structures should be calculated
	// NOTE: Do not use predictive analysis if the first parameter is false
	hbCalc.SetHbSettings(true, false, false)
	if !runAna || !hbCalc.CalcVC {
		for key := range analysisCasesMap {
			analysisCasesMap[key] = false
		}
	}

	data.AnalysisCases = analysisCasesMap
	data.InitAnalysisData(data.AnalysisCases, fuzzing)

	if hb.CalcVC {
		vc.InitVC()
	}

	if hb.CalcPog {
		pog.InitPOG()
	}

	if hb.CalcCssts {
		cssts.InitCSSTs(data.GetNoRoutines(), data.GetTraceLengths())
	}

	if data.AnalysisCases["resourceDeadlock"] {
		scenarios.ResetState()
	}

	if hb.CalcVC {
		vc.CurrentVC[1].Inc(1)
		vc.CurrentWVC[1].Inc(1)
	}

	traceIter := data.MainTrace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		// add edge between element of same routine to partial order trace
		if hb.CalcPog {
			pog.AddEdgeSameRoutineAndFork(elem)
		}

		// count how many operations where executed on the underlying structure
		// do not count for operations that do not have an underlying structure
		switch e := elem.(type) {
		case *trace.ElementFork, *trace.ElementNew, *trace.ElementReplay, *trace.ElementRoutineEnd:
		default:
			data.AddOpsPerID(e.GetID())
		}

		switch e := elem.(type) {
		case *trace.ElementAtomic:
			elements.AnalyzeAtomic(e, ignoreCriticalSections)
		case *trace.ElementChannel:
			elements.UpdateChannel(e)
		case *trace.ElementMutex:
			if ignoreCriticalSections {
				elements.UpdateMutex(e, true)
			} else {
				elements.UpdateMutex(e, false)
			}
			if data.AnalysisFuzzingFlow {
				scenarios.GetConcurrentMutexForFuzzing(e)
			}
		case *trace.ElementFork:
			elements.AnalyzeFork(e)
		case *trace.ElementSelect:
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
			elements.UpdateSelect(e)
		case *trace.ElementWait:
			elements.AnalyzeWait(e)
		case *trace.ElementCond:
			elements.AnalyzeCond(e)
		case *trace.ElementOnce:
			elements.AnalyzeOnce(e)
			if data.AnalysisFuzzingFlow {
				scenarios.GetConcurrentOnceForFuzzing(e)
			}
		case *trace.ElementRoutineEnd:
			elements.AnalyzeRoutineEnd(e)
		case *trace.ElementNew:
			elements.AnalyzeNew(e)
		}

		if data.AnalysisCases["resourceDeadlock"] {
			switch e := elem.(type) {
			case *trace.ElementMutex:
				scenarios.HandleMutexEventForRessourceDeadlock(*e)
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

	data.MainTrace.SetHBWasCalc(hb.CalcVC)

	log.Info("Finished HB analysis")

	if data.ModeIsFuzzing {
		scenarios.RerunCheckForSelectCaseWithPartnerChannel()
		scenarios.CheckForSelectCaseWithPartner()
	}

	if memory.WasCanceled() {
		return
	}

	if data.AnalysisCases["leak"] {
		log.Info("Check for leak")
		scenarios.CheckForLeak()
		scenarios.CheckForStuckRoutine(false)
		log.Info("Finish check for leak")
	}

	if memory.WasCanceled() {
		return
	}

	if data.AnalysisCases["doneBeforeAdd"] {
		log.Info("Check for done before add")
		scenarios.CheckForDoneBeforeAdd()
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
		scenarios.CheckForResourceDeadlock()
		log.Info("Finish check for cyclic deadlock")
	}

	if memory.WasCanceled() {
		return
	}

	if data.AnalysisCases["unlockBeforeLock"] {
		log.Info("Check for unlock before lock")
		scenarios.CheckForUnlockBeforeLock()
		log.Info("Finish check for unlock before lock")
	}
}

// checkLeak checks for a given element if it leaked (has no tPost). If so,
// it will look for a possible way to resolve the leak
//
// Parameter:
//   - elem TraceElement: Element to check
func checkLeak(elem trace.Element) {
	switch e := elem.(type) {
	case *trace.ElementChannel:
		scenarios.CheckForLeakChannelStuck(e, vc.CurrentVC[e.GetRoutine()])
	case *trace.ElementMutex:
		scenarios.CheckForLeakMutex(e)
	case *trace.ElementWait:
		scenarios.CheckForLeakWait(e)
	case *trace.ElementSelect:
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
		scenarios.CheckForLeakSelectStuck(e, ids, buffered, vc.CurrentVC[e.GetRoutine()], opTypes)
	case *trace.ElementCond:
		scenarios.CheckForLeakCond(e)
	}
}
