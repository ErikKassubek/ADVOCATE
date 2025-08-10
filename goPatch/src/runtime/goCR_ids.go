// File: goCR_ids.go
// Brief: Get required ids and timestamps
//
// Created: 2025-03-21
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var goCRCurrentRoutineID atomic.Uint64

// GetNewGoCRRoutineID returns a new id for a routine
//
// Returns:
//   - new id
func GetNewGoCRRoutineID() uint64 {
	id := goCRCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

// GetNextGoCRRoutineID returns the next routine id that will be provided
// by GetNextGoCRRoutineID without advancing the counter
//
// Returns:
//   - next id
func GetNextGoCRRoutineID() uint64 {
	return goCRCurrentRoutineID.Load() + 1
}

// GetGoCRObjectID returns a new id for an primitive
// Return:
//   - new id
func GetGoCRObjectID() uint64 {
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
