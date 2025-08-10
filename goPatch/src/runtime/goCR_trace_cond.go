// GOCP-FILE_START

// File: goCR_trace_cond.go
// Brief: Functionality for the conditional variables
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

// Struct to store an operation on a conditional variable
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - id string: id of the channel
//   - op Operation: operation type
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCRTraceCond struct {
	tPre  int64
	tPost int64
	id    uint64
	op    Operation
	file  string
	line  int
}

/*
 * GoCRCondPre adds a cond wait to the trace
 * Args:
 * 	id: id of the cond
 * 	op: Operation
 * Return:
 * 	index of the operation in the trace
 */
func GoCRCondPre(id uint64, op Operation) int {
	if goCRTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()
	_, file, line, _ := Caller(2)

	if GoCRIgnore(file) {
		return -1
	}

	elem := GoCRTraceCond{
		tPre: timer,
		id:   id,
		op:   op,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

/*
 * GoCRCondPost adds the end counter to an operation of the trace
 * Args:
 * 	index: index of the operation in the trace
 */
func GoCRCondPost(index int) {
	if goCRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	if index == -1 {
		return
	}
	elem := currentGoRoutineInfo().getElement(index).(GoCRTraceCond)

	elem.tPost = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (elem GoCRTraceCond) toString() string {
	var opC string
	switch elem.op {
	case OperationCondWait:
		opC = "W"
	case OperationCondSignal:
		opC = "S"
	case OperationCondBroadcast:
		opC = "B"
	}

	return buildTraceElemString("D", elem.tPre, elem.tPost, elem.id, opC, posToString(elem.file, elem.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem GoCRTraceCond) getOperation() Operation {
	return elem.op
}
