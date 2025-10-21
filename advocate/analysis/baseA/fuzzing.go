// Copyright (c) 2024 Erik Kassubek
//
// File: analysisConcurrentCommunication.go
// Brief: Data collected during analysis for use in fuzzing
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package baseA

// for fuzzing flow
var (
	FuzzingFlowOnce  = make([]ConcurrentEntry, 0)
	FuzzingFlowMutex = make([]ConcurrentEntry, 0)
	FuzzingFlowSend  = make([]ConcurrentEntry, 0)
	FuzzingFlowRecv  = make([]ConcurrentEntry, 0)

	FuzzingCounter = make(map[int]map[string]int) // id -> pos -> counter

	T1 = false
)
