// Copyright (c) 2025 Erik Kassubek
//
// File: gocdr_ids.go
// Brief: Get required ids and timestamps
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var gocdrCurrentRoutineID atomic.Uint64

// GetNewGocdrRoutineID returns a new id for a routine
//
// Returns:
//   - new id
func GetNewGocdrRoutineID() uint64 {
	id := gocdrCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

// GetNewGocdrRoutineID returns the next routine id that will be provided
// by GetNewGocdrRoutineID without advancing the counter
//
// Returns:
//   - next id
func GetNextGocdrRoutineID() uint64 {
	return gocdrCurrentRoutineID.Load() + 1
}

// GetGocdrObjectID returns a new id for an primitive
// Return:
//   - new id
func GetGocdrObjectID() uint64 {
	routine := currentGoRoutineInfo()

	if routine == nil {
		return 0
	}

	routine.maxObjectId++
	if routine.maxObjectId > 999999999 {
		panic("Overflow Error: Tow many objects in one routine. Max: 999999999")
	}
	id := routine.id*1000000000 + routine.maxObjectId
	return id
}
