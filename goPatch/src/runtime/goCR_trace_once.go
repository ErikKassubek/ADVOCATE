// GOCP-FILE_START

// File: goCR_trace_once.go
// Brief: Functionality for once
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

// Struct to store an operation on a once
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - id string: id of the once
//   - suc bool: true if the func in the Do was executed, false otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCRTraceOnce struct {
	tPre  int64
	tPost int64
	id    uint64
	suc   bool
	file  string
	line  int
}

// GoCROncePre adds a once to the trace
//
// Parameter:
//   - id uint64: id of the once
//
// Returns:
//   - int: index of the operation in the trace
func GoCROncePre(id uint64) int {
	if goCRTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	if GoCRIgnore(file) {
		return -1
	}

	elem := GoCRTraceOnce{
		tPre: timer,
		id:   id,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// Add the end counter to an operation of the trace
//
// Parameter:
//   - index int: index of the operation in the trace
//   - suc bool: true if the do on the once was called for the first time, false otherwise
func GoCROncePost(index int, suc bool) {
	if goCRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}
	elem := currentGoRoutineInfo().getElement(index).(GoCRTraceOnce)

	elem.tPost = timer
	elem.suc = suc

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     O,[tPre],[tPost],[id],[suc],[file],[line]
func (elem GoCRTraceOnce) toString() string {
	return buildTraceElemString("O", elem.tPre, elem.tPost, elem.id, elem.suc, posToString(elem.file, elem.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elen GoCRTraceOnce) getOperation() Operation {
	return OperationOnceDo
}
