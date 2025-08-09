// GOCP-FILE_START

// File: goCR_trace_mutex.go
// Brief: Functionality for mutex
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

// Struct to store an operation on a mutex
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - id uint64: id of the mutex
//   - op Operation: operation type
//   - suc bool: false if a trymutex did not manage to lock the mutex, true otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCRTraceMutex struct {
	tPre  int64
	tPost int64
	id    uint64
	op    Operation
	suc   bool
	file  string
	line  int
}

var lastRWOp = make(map[uint64]int64) // routine -> tPost
var lastRWOpLock mutex

// GoCRMutexPre adds a mutex lock to the trace
//
// Parameter:
//   - id uint64: id of the mutex
//   - op Operation: type of operation
//
// Returns:
//   - index of the operation in the trace
func GoCRMutexPre(id uint64, op Operation) int {
	if goCRTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	if GoCRIgnore(file) {
		return -1
	}

	elem := GoCRTraceMutex{
		tPre: timer,
		id:   id,
		op:   op,
		suc:  true,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// GoCRMutexPost adds the end counter to an operation of the trace.
// For try use GoCRMutexTryPost.
//
// Parameters:
//   - index: index of the operation in the trace
//   - suc: wether the lock was successfull for try, otherwise true
func GoCRMutexPost(index int, suc bool) {
	if goCRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	// internal elements are not in the trace
	if index == -1 {
		return
	}

	// only needed to fix tests

	if currentGoRoutineInfo() == nil {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(GoCRTraceMutex)
	routine := currentGoRoutineInfo().id

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

	currentGoRoutineInfo().updateElement(index, elem)
}

// Check if the mutex is a rw mutex
//
// Returns:
//   - bool: true if it is a rwMutex, false otherwise
func (elem GoCRTraceMutex) isRw() bool {
	if elem.op == OperationMutexLock || elem.op == OperationMutexUnlock || elem.op == OperationMutexTryLock {
		return false
	}
	return true
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (elem GoCRTraceMutex) toString() string {
	opStr, rw := elem.opRwToString()

	return buildTraceElemString("M", elem.tPre, elem.tPost, elem.id, rw, opStr, elem.suc, posToString(elem.file, elem.line))
}

// Get the string representations for the operation and rw fields
//
// Returns:
//   - string: the operation string representation
//   - string: the rw string representation
func (elem GoCRTraceMutex) opRwToString() (string, string) {
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

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem GoCRTraceMutex) getOperation() Operation {
	return elem.op
}
