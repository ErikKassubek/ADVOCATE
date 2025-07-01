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

import "advocate/analysis/clock"

var (
	// current happens before vector clocks
	CurrentVC = make(map[int]*clock.VectorClock)

	// current must happens before vector clocks
	CurrentWVC = make(map[int]*clock.VectorClock)

	// vector clocks for last write times
	Lw = make(map[int]*clock.VectorClock)

	// vector clock for each buffer place in vector clock
	// the map key is the channel id. The slice is used for the buffer positions
	BufferedVCs = make(map[int]([]BufferedVC))
	// the current buffer position
	BufferedVCsCount = make(map[int]int)
	BufferedVCsSize  = make(map[int]int)

	// vector clocks for last release times
	RelW = make(map[int]*clock.VectorClock) // id -> vc
	RelR = make(map[int]*clock.VectorClock) // id -> vc

	// vector clock for each wait group
	LastChangeWG = make(map[int]*clock.VectorClock)
)

// NewLW sets the last write times for a given atomic variable to a new vector clock
// If it is already set, nothing happens
//
// Parameter:
//   - id int: the id of the atomic variable
//   - nRout int: number of routines in the trace
func NewLW(id int, nRout int) {
	if _, ok := Lw[id]; !ok {
		Lw[id] = clock.NewVectorClock(nRout)
	}
}
