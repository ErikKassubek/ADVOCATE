// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_ids.go
// Brief: Get required ids and timestamps
//
// Author: Erik Kassubek
// Created: 2025-03-21
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var advocateCurrentRoutineID atomic.Uint64

// GetNewAdvocateRoutineID returns a new id for a routine
//
// Returns:
//   - new id
func GetNewAdvocateRoutineID() uint64 {
	id := advocateCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

// GetNewAdvocateRoutineID returns the next routine id that will be provided
// by GetNewAdvocateRoutineID without advancing the counter
//
// Returns:
//   - next id
func GetNextAdvocateRoutineID() uint64 {
	return advocateCurrentRoutineID.Load() + 1
}

// GetAdvocateObjectID returns a new id for an primitive
// Return:
//   - new id
func GetAdvocateObjectID() uint64 {
	routine := currentGoRoutine()

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
