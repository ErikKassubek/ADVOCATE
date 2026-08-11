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

import (
	"internal/runtime/atomic"
)

var gocdrCurrentRoutineID atomic.Uint64

// GetNewGoCDRRoutineID returns a new id for a routine
//
// Returns:
//   - new id
func GetNewGoCDRRoutineID() uint64 {
	id := gocdrCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

// GetNewGoCDRRoutineID returns the next routine id that will be provided
// by GetNewGoCDRRoutineID without advancing the counter
//
// Returns:
//   - next id
func GetNextGoCDRRoutineID() uint64 {
	return gocdrCurrentRoutineID.Load() + 1
}

func NewIdIfReq(currentId uint64, memOld, memCurr uintptr) (uint64, uintptr) {
	if currentId == 0 {
		return GetGoCDRObjectID(), memCurr
	}

	if memOld == memCurr {
		return currentId, memCurr
	}

	return GetGoCDRObjectID(), memCurr
}

// GetGoCDRObjectID returns a new id for an primitive
// Return:
//   - new id
func GetGoCDRObjectID() uint64 {
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
