// Copyright (c) 2024 Erik Kassubek
//
// File: analysisData.go
// Brief: Init data
//
// Author: Erik Kassubek
// Created: 2024-01-27
//
// License: BSD-3-Clause

package data

// InitAnalysisData initializes the analysis by setting the analysis cases and fuzzing
//
// Parameters:
//   - analysisCasesMap map[string]bool: map with information about which
//     analysis parts should be run
//   - anaFuzzingFlow bool: true if fuzzing flow, false otherwise
func InitAnalysisData(analysisCasesMap map[string]bool, anaFuzzingFlow bool) {
	AnalysisCases = analysisCasesMap
	AnalysisFuzzingFlow = anaFuzzingFlow
}
