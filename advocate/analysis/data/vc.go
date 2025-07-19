// Copyright (c) 2025 Erik Kassubek
//
// File: vc.go
// Brief: Data required for calculating the vector clocks
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package data

import (
	"advocate/analysis/concurrent/clock"
	"advocate/trace"
)

var (
	// current happens before vector clocks
	CurrentVC = make(map[int]*clock.VectorClock)

	// current must happens before vector clocks
	CurrentWVC = make(map[int]*clock.VectorClock)

	// vector clock for each buffer place in vector clock
	// the map key is the channel id. The slice is used for the buffer positions
	BufferedVCs = make(map[int]([]BufferedVC))
	// the current buffer position
	BufferedVCsCount = make(map[int]int)
	BufferedVCsSize  = make(map[int]int)

	// vector clocks for last release times
	RelW = make(map[int]*ElemWithVc) // id -> release
	RelR = make(map[int]*ElemWithVc) // id -> release

	// vector clock for each wait group
	LastChangeWG = make(map[int]*trace.ElementWait)
)

func InitVC() {
	noRoutine := MainTrace.GetNoRoutines()
	for i := 1; i <= noRoutine; i++ {
		CurrentVC[i] = clock.NewVectorClock(noRoutine)
		CurrentWVC[i] = clock.NewVectorClock(noRoutine)
	}
}
