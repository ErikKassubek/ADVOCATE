// gocdr/analysis/analysis/analysis.go

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
	"gocdr/analysis/analysis/scenarios"
	"gocdr/analysis/baseA"
	"gocdr/utils/control"
	"gocdr/utils/log"
	"gocdr/utils/timer"
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
	scenarios.RunAnalysisLeak()

}
