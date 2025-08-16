//
// File: analysis.go
// Brief: analysis of traces if performed from here
//
// Created: 2025-01-01
//
// License: BSD-3-Clause

package analysis

import (
	"goCR/analysis/analysis/elements"
	"goCR/analysis/analysis/scenarios"
	"goCR/analysis/data"
	"goCR/analysis/hb/vc"
	fuzzdata "goCR/fuzzing/data"
	"goCR/trace"
	"goCR/utils/control"
	"goCR/utils/log"
	"goCR/utils/timer"
)

// RunAnalysis starts the analysis of the main trace
//
// Parameter:
//   - fuzzing bool: true if run with fuzzing
//   - onlyAPanicAndLeak bool: only test for actual panics and leaks
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

	data.AnalysisFuzzingFlow = fuzzing

	timer.Start(timer.Analysis)
	defer timer.Stop(timer.Analysis)

	scenarios.RunAnalysisOnExitCodes(true)

	scenarios.CheckForStuckRoutine(true)

	if !fuzzing {
		return
	}

	if !fuzzing || fuzzdata.UseHBInfoFuzzing {
		RunHBAnalysis()
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
func RunHBAnalysis() {
	log.Info("Start Analysis")

	vc.InitVC()

	vc.CurrentVC[1].Inc(1)
	vc.CurrentWVC[1].Inc(1)

	traceIter := data.MainTrace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {

		// not enough memory
		if control.WasCanceledRAM.Load() {
			return
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
			elements.AnalyzeAtomic(e)
		case *trace.ElementChannel:
			elements.UpdateChannel(e)
		case *trace.ElementMutex:
			elements.UpdateMutex(e, false)
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
		case *trace.ElementRoutineEnd:
			elements.AnalyzeRoutineEnd(e)
		case *trace.ElementNew:
			elements.AnalyzeNew(e)
		}

		if elem.GetTPost() == 0 {
			checkLeak(elem)
		}

		if control.CheckCanceled() {
			return
		}
	}

	data.MainTrace.SetHBWasCalc(true)

	log.Info("Finished HB analysis")

	if fuzzdata.FuzzingModeGFuzz {
		scenarios.RerunCheckForSelectCaseWithPartnerChannel()
		scenarios.CheckForSelectCaseWithPartner()
	}

	if control.CheckCanceled() {
		return
	}

	log.Info("Check for leak")
	scenarios.CheckForLeak()
	scenarios.CheckForStuckRoutine(false)
	log.Info("Finish check for leak")
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
