//
// File: analysisConcurrentCommunication.go
// Brief: Data collected during analysis for use in fuzzing
//
// Created: 2025-07-01
//
// License: BSD-3-Clause

package data

// for fuzzing flow
var (
	FuzzingCounter = make(map[int]map[string]int) // id -> pos -> counter
)
