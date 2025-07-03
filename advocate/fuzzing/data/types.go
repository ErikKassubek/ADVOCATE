// Copyright (c) 2025 Erik Kassubek
//
// File: types.go
// Brief: types for fuzzing
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package data

// Struct to handle the selects for fuzzing
//
//   - Id string: replay id
//   - T int: tPost of the select execution, used for order
//   - ChosenCase int: id of the chosen case, -1 for default
//   - NumberCases int: number of cases not including default
//   - ContainsDefault bool: true if contains default case, otherwise false
//   - CasiWithPos[]int: list of casi with possible partner
type FuzzingSelect struct {
	ID              string
	T               int
	ChosenCase      int
	NumberCases     int
	ContainsDefault bool
	CasiWithPos     []int
}

// Encapsulating type for the different mutations
type Mutation struct {
	MutType int
	MutSel  map[string][]FuzzingSelect
	MutFlow map[string]int
	MutPie  int
}
