// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_mutex.go
// Brief: Functionality for mutex
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause


package runtime

// MARK: Pre

var lastRWOp = make(map[uint64]uint64) // routine -> tPost
var lastRWOpLock mutex

/*
 * AdvocateMutexLockPre adds a mutex lock to the trace
 * Args:
 * 	id: id of the mutex
 *  rw: true if it is a rwmutex
 *  r: true if it is a rlock operation
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateMutexLockPre(id uint64, rw bool, r bool) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	var op string
	var rwStr string
	if !rw { // Mutex
		rwStr = "-"
		if !r { // Lock
			op = "L"
		} else { // rLock, invalid case
			panic("Tried to RLock a non-RW Mutex")
		}
	} else { // RWMutex
		rwStr = "R"
		if !r { // Lock
			op = "L"
		} else { // RLock
			op = "R"
		}
	}

	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := "M," + uint64ToString(timer) + ",0," + uint64ToString(id) + "," +
		rwStr + "," + op + ",t," + file + ":" + uint64ToString(uint64(line))

	return insertIntoTrace(elem)
}

/*
 * AdvocateMutexLockTry adds a mutex trylock to the trace
 * Args:
 * 	id: id of the mutex
 * 	rw: true if it is a rwmutex
 * 	r: true if it is a rlock operation
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateMutexLockTry(id uint64, rw bool, r bool) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	var op string
	var rwStr string
	if !rw { // Mutex
		rwStr = "-"
		if !r { // Lock
			op = "T"
		} else { // rLock, invalid case
			panic("Tried to TryRLock a non-RW Mutex")
		}
	} else { // RWMutex
		rwStr = "R"
		if !r { // Lock
			op = "T"
		} else { // RLock
			op = "Y"
		}
	}

	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := "M," + uint64ToString(timer) + ",0," + uint64ToString(id) + "," +
		rwStr + "," + op + ",f," + file + ":" + uint64ToString(uint64(line))

	return insertIntoTrace(elem)
}

/*
 * AdvocateUnlockPre adds a mutex unlock to the trace
 * Args:
 * 	id: id of the mutex
 * 	rw: true if it is a runlock
 * 	r: true if it is a rlock operation
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateUnlockPre(id uint64, rw bool, r bool) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	var op string
	var rwStr string
	if !rw { // Mutex
		rwStr = "-"
		if !r { // Lock
			op = "U"
		} else { // rLock, invalid case
			panic("Tried to RUnlock a non-RW Mutex")
		}
	} else { // RWMutex
		rwStr = "R"
		if !r { // Lock
			op = "U"
		} else { // RLock
			op = "N"
		}
	}

	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := "M," + uint64ToString(timer) + ",0," + uint64ToString(id) + "," +
		rwStr + "," + op + ",t," + file + ":" + uint64ToString(uint64(line))

	return insertIntoTrace(elem)
}

// MARK: Post

/*
 * AdvocateMutexPost adds the end counter to an operation of the trace.
 * For try use AdvocatePostTry.
 * Also used for wait group
 * Args:
 * 	index: index of the operation in the trace
 * 	c: number of the send
 */
func AdvocateMutexPost(index int) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	// internal elements are not in the trace
	if index == -1 {
		return
	}

	// only needed to fix tests

	if currentGoRoutine() == nil {
		return
	}

	elem := currentGoRoutine().getElement(index)
	split := splitStringAtCommas(elem, []int{2, 3, 4, 5, 6, 7})
	routine := currentGoRoutine().id

	lock(&lastRWOpLock)
	if split[3] == "R" && lastRWOp[routine] != 0 {
		split[1] = uint64ToString(lastRWOp[routine] - 1)
		lastRWOp[routine] = 0
	} else {
		split[1] = uint64ToString(timer)
	}

	path := splitStringAtSeparator(split[6], ':', nil)
	if hasSuffix(path[0], "sync/rwmutex.go") {
		lastRWOp[routine] = timer
	}
	unlock(&lastRWOpLock)

	elem = mergeString(split)

	currentGoRoutine().updateElement(index, elem)
}

/*
 * AdvocatePostTry adds the end counter to an try operation of the trace
 * Args:
 * 	index: index of the operation in the trace
 * 	suc: true if the try was successful, false otherwise
 */
func AdvocatePostTry(index int, suc bool) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	// internal elements are not in the trace
	if index == -1 {
		return
	}

	elem := currentGoRoutine().getElement(index)
	split := splitStringAtCommas(elem, []int{2, 3, 4, 5, 6, 7})
	routine := currentGoRoutine().id

	lock(&lastRWOpLock)
	if split[3] == "R" && lastRWOp[routine] != 0 {
		split[1] = uint64ToString(lastRWOp[routine] - 1)
		lastRWOp[routine] = 0
	} else {
		split[1] = uint64ToString(timer)
	}

	path := splitStringAtSeparator(split[6], ':', nil)
	if hasSuffix(path[0], "sync/rwmutex.go") {
		lastRWOp[routine] = timer
	}
	unlock(&lastRWOpLock)

	split[3] = boolToString(suc)

	elem = mergeString(split)

	currentGoRoutine().updateElement(index, elem)
}
