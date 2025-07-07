// Copyright (c) 2025 Erik Kassubek
//
// File: csst.go
// Brief: functions to use the csst
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package cssts

import (
	"advocate/trace"
	"advocate/utils/types"
)

var (
	Csst         IncrementalCSST
	CsstInverted IncrementalCSST
)

func InitCSSTs(numberRoutines int, lengths []int) {
	Csst = NewIncrementalCSST(lengths)
	CsstInverted = NewIncrementalCSST(lengths)
}

// For a trace element, return the routine id and elem rout index used as identifier
// in the CSST
//
// Parameters:
//   - elem trace.Element: the element to find the index for
//
// Returns:
//   - types.Pair[int, int]: routine id of elem, routine local index of elem
func getIndicesFromTraceElem(elem trace.Element) types.Pair[int, int] {
	rout, index := elem.GetTraceIndex()
	return types.NewPair(rout, index)
}

// Function to filter out element which do not correspond to valid operations, e.g.
// end of a routine
//
// Parameter:
//   - elem trace.Element: the element to test
//
// Returns:
//   - bool: true if the element is valid, false otherwise
func valid(elem trace.Element) bool {
	t := elem.GetObjType(false)
	return !(t == trace.ObjectTypeReplay || t == trace.ObjectTypeNew || t == trace.ObjectTypeRoutineEnd)
}
