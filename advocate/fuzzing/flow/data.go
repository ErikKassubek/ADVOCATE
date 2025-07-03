// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data for flow
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package flow

const maxFlowMut = 10

var (
	alreadyDelayedElems = make(map[string][]int)
)

func ClearData() {
	alreadyDelayedElems = make(map[string][]int)
}
