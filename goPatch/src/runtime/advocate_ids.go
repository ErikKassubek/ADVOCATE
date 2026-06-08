// Copyright (c) 2025 Erik Kassubek
//
// File: oosc_ids.go
// Brief: Get required ids and timestamps
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var ooscCurrentRoutineID atomic.Uint64

// GetNewOoscRoutineID returns a new id for a routine
//
// Returns:
//   - new id
func GetNewOoscRoutineID() uint64 {
	id := ooscCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

// GetNewOoscRoutineID returns the next routine id that will be provided
// by GetNewOoscRoutineID without advancing the counter
//
// Returns:
//   - next id
func GetNextOoscRoutineID() uint64 {
	return ooscCurrentRoutineID.Load() + 1
}

// GetOoscObjectID returns a new id for an primitive
// Return:
//   - new id
func GetOoscObjectID() uint64 {
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
