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

type AdvocateTraceMutex struct {
	tPre  int64
	tPost int64
	id    uint64
	op    Operation
	suc   bool
	file  string
	line  int
}

// MARK: Pre

var lastRWOp = make(map[uint64]int64) // routine -> tPost
var lastRWOpLock mutex

/*
 * AdvocateMutexPre adds a mutex lock to the trace
 * Args:
 * 	id: id of the mutex
 *  rw: true if it is a rwmutex
 *  r: true if it is a rlock operation
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateMutexPre(id uint64, op Operation) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := AdvocateTraceMutex{
		tPre: timer,
		id:   id,
		op:   op,
		suc:  true,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// MARK: Post

/*
 * AdvocateMutexPost adds the end counter to an operation of the trace.
 * For try use AdvocateMutexTryPost.
 * Also used for wait group
 * Args:
 * 	index: index of the operation in the trace
 * 	suc: wether the lock was successfull for try, otherwise true
 */
func AdvocateMutexPost(index int, suc bool) {
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

	elem := currentGoRoutine().getElement(index).(AdvocateTraceMutex)
	routine := currentGoRoutine().id

	lock(&lastRWOpLock)
	if elem.isRw() && lastRWOp[routine] != 0 {
		elem.tPost = lastRWOp[routine] - 1
		lastRWOp[routine] = 0
	} else {
		elem.tPost = timer
	}

	if hasSuffix(elem.file, "sync/rwmutex.go") {
		lastRWOp[routine] = timer
	}
	unlock(&lastRWOpLock)

	elem.suc = suc

	currentGoRoutine().updateElement(index, elem)
}

func (elem AdvocateTraceMutex) isRw() bool {
	if elem.op == OperationMutexLock || elem.op == OperationMutexUnlock || elem.op == OperationMutexTryLock {
		return false
	}
	return true
}

func (elem AdvocateTraceMutex) toString() string {
	opStr, rw := elem.opRwToString()

	return buildTraceElemString("M", elem.tPre, elem.tPost, elem.id, rw, opStr, elem.suc, posToString(elem.file, elem.line))
}

func (elem AdvocateTraceMutex) opRwToString() (string, string) {
	opStr := ""
	rw := "f"
	switch elem.op {
	case OperationMutexLock:
		opStr = "L"
	case OperationMutexUnlock:
		opStr = "U"
	case OperationMutexTryLock:
		opStr = "T"
	case OperationRWMutexLock:
		opStr = "L"
		rw = "t"
	case OperationRWMutexUnlock:
		opStr = "U"
		rw = "t"
	case OperationRWMutexTryLock:
		opStr = "T"
		rw = "t"
	case OperationRWMutexRLock:
		opStr = "R"
		rw = "t"
	case OperationRWMutexRUnlock:
		opStr = "N"
		rw = "t"
	case OperationRWMutexTryRLock:
		opStr = "Y"
		rw = "t"
	}

	return opStr, rw
}

func (elem AdvocateTraceMutex) getOperation() Operation {
	return elem.op
}
