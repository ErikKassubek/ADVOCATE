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
	"advocate/analysis/baseA"
	"advocate/analysis/hb/cssts"
	"advocate/analysis/hb/hbcalc"
	hb "advocate/analysis/hb/hbcalc"
	"advocate/analysis/hb/pog"
	"advocate/analysis/hb/vc"
	"advocate/fuzzing/baseF"
	"advocate/trace"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// RunAnalysis starts the analysis of the main trace
//
// Parameter:
//   - fuzzing bool: true if run with fuzzing
func RunAnalysis(fuzzing bool) {
	// catch panics in analysis.
	// Prevents the whole toolchain to panic if one analysis panics
	if log.IsPanicPrevent() {
		defer func() {
			if r := recover(); r != nil {
				control.Cancel()
				log.Error(r)
			}
		}()
	}

	baseA.AnalysisFuzzingFlow = fuzzing

	timer.Start(timer.Analysis)
	defer timer.Stop(timer.Analysis)

	scenarios.RunAnalysisOnExitCodes(true)

	if baseA.AnalysisCasesMap[flags.Leak] || flags.OnlyAPanicAndLeak {
		err := scenarios.Blocked()
		if err != nil {
			log.Error("Failed to read partial deadlock info: ", err.Error())
		}
	}

	if flags.OnlyAPanicAndLeak {
		scenarios.CheckForStuckRoutine(true)

		if !fuzzing {
			return
		}
	}

	if !fuzzing || baseF.UseHBInfoFuzzing {
		RunHBAnalysis(fuzzing)
	}
}

// RunHBAnalysis runs the full analysis happens before based analysis
//
// Parameter:
//   - fuzzing bool: true if run with fuzzing
//   - runAna bool: true to run the predictive analysis
//
// Returns:
//   - bool: true if something has been found
func RunHBAnalysis(fuzzing bool) {
	log.Info("Start Analysis")

	baseA.ModeIsFuzzing = fuzzing

	// set which hb structures should be calculated
	// NOTE: Do not use predictive analysis if the first parameter is false
	hbcalc.SetHbSettings(true, false, false)
	if flags.OnlyAPanicAndLeak || !hbcalc.CalcVC {
		for key := range baseA.AnalysisCasesMap {
			baseA.AnalysisCasesMap[key] = false
		}
	}

	if hb.CalcVC {
		vc.InitVC()
	}

	if hb.CalcPog {
		pog.InitPOG()
	}

	if hb.CalcCssts {
		cssts.InitCSSTs(baseA.GetTraceLengths())
	}

	if baseA.AnalysisCasesMap[flags.ResourceDeadlock] {
		scenarios.ResetState()
	}

	if hb.CalcVC {
		vc.CurrentVC[1].Inc(1)
		vc.CurrentWVC[1].Inc(1)
	}

	traceIter := baseA.MainTrace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {

		// not enough memory
		if control.WasCanceledRAM.Load() {
			return
		}

		// add edge between element of same routine to partial order trace
		if hb.CalcPog {
			pog.AddEdgeSameRoutineAndFork(elem)
		}

		// count how many operations where executed on the underlying structure
		// do not count for operations that do not have an underlying structure
		switch e := elem.(type) {
		case *trace.ElementFork, *trace.ElementNew, *trace.ElementReplay, *trace.ElementRoutineEnd:
		default:
			baseA.AddOpsPerID(e.GetID())
		}

		switch e := elem.(type) {
		case *trace.ElementAtomic:
			elements.AnalyzeAtomic(e)
		case *trace.ElementChannel:
			elements.UpdateChannel(e)
		case *trace.ElementMutex:
			if flags.IgnoreCriticalSection {
				elements.UpdateMutex(e, true)
			} else {
				elements.UpdateMutex(e, false)
			}
			if baseA.AnalysisFuzzingFlow {
				scenarios.GetConcurrentMutexForFuzzing(e)
			}
		case *trace.ElementFork:
			elements.AnalyzeFork(e)
		case *trace.ElementSelect:
			cases := e.GetCases()
			ids := make([]int, 0)
			opTypes := make([]int, 0)
			for _, c := range cases {
				switch c.GetType(true) {
				case trace.ChannelSend:
					ids = append(ids, c.GetID())
					opTypes = append(opTypes, 0)
				case trace.ChannelRecv:
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
			if baseA.AnalysisFuzzingFlow {
				scenarios.GetConcurrentOnceForFuzzing(e)
			}
		case *trace.ElementRoutineEnd:
			elements.AnalyzeRoutineEnd(e)
		case *trace.ElementNew:
			elements.AnalyzeNew(e)
		}

		if baseA.AnalysisCasesMap[flags.ResourceDeadlock] {
			switch e := elem.(type) {
			case *trace.ElementMutex:
				scenarios.HandleMutexEventForRessourceDeadlock(*e)
			}
		}

		// check for leak
		if baseA.AnalysisCasesMap[flags.Leak] && elem.GetTPost() == 0 {
			checkLeak(elem)
		}

		if control.CheckCanceled() {
			return
		}
	}

	baseA.MainTrace.SetHBWasCalc(hb.CalcVC)

	log.Info("Finished HB analysis")

	if baseF.FuzzingModeGFuzz || baseA.AnalysisCasesMap[flags.Leak] {
		scenarios.RerunCheckForSelectCaseWithPartnerChannel()
		scenarios.CheckForSelectCaseWithPartner()
	}

	if control.CheckCanceled() {
		return
	}

	if baseA.AnalysisCasesMap[flags.Leak] {
		log.Info("Check for leak")
		scenarios.CheckForLeak()
		scenarios.CheckForStuckRoutine(false)
		log.Info("Finish check for leak")
	}

	if control.CheckCanceled() {
		return
	}

	if baseA.AnalysisCasesMap[flags.DoneBeforeAdd] {
		log.Info("Check for done before add")
		scenarios.CheckForDoneBeforeAdd()
		log.Info("Finish check for done before add")
	}

	if control.CheckCanceled() {
		return
	}

	if baseA.AnalysisCasesMap[flags.ResourceDeadlock] {
		log.Info("Check for cyclic deadlock")
		scenarios.CheckForResourceDeadlock()
		log.Info("Finish check for cyclic deadlock")
	}

	if control.CheckCanceled() {
		return
	}

	if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
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
		opTypes := make([]trace.ObjectType, 0)
		for _, c := range cases {
			ids = append(ids, c.GetID())
			opTypes = append(opTypes, c.GetType(true))
			buffered = append(buffered, c.IsBuffered())

		}
		timer.Stop(timer.AnaLeak)
		scenarios.CheckForLeakSelectStuck(e, ids, buffered, vc.CurrentVC[e.GetRoutine()], opTypes)
	case *trace.ElementCond:
		scenarios.CheckForLeakCond(e)
	}
}
