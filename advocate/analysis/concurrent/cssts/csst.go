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

// For a given element, return concurrent events
// Parameter:
//   - elem trace.TraceElem: the element to search for
//   - all bool: if true, return all concurrent events, otherwise return one
//
// Returns:
//   - []trace.TraceElement: the concurrent element(s)
func GetConcurrentCSST(elem trace.TraceElement, all bool) []trace.TraceElement {
	// TODO: implement
	return make([]trace.TraceElement, 0)
}

func GetIndicesFromTraceElem(elem trace.TraceElement) types.Pair[int, int] {
	rout, index := elem.GetTraceIndex()
	return types.NewPair(rout, index)
}
