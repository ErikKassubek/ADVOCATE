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
var advocateGlobalCounter atomic.Int64

/*
 * GetAdvocateRoutineID returns a new id for a routine
 * Return:
 * 	new id
 */
func GetAdvocateRoutineID() uint64 {
	id := advocateCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

/*
 * GetAdvocateObjectID returns a new id for a mutex, channel or waitgroup
 * Return:
 * 	new id
 */
func GetAdvocateObjectID() uint64 {
	routine := currentGoRoutine()

	if routine == nil {
		getg().advocateRoutineInfo = newAdvocateRoutine(getg())
		routine = currentGoRoutine()
	}

	routine.maxObjectId++
	if routine.maxObjectId > 999999999 {
		panic("Overflow Error: Tow many objects in one routine. Max: 999999999")
	}
	id := routine.id*1000000000 + routine.maxObjectId
	return id
}

/*
 * GetAdvocateCounter will update the timer and return the new value
 * Return:
 * 	new time value
 */
func GetNextTimeStep() int64 {
	return nanotime()
}
