// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_new_elem.go
// Brief: Functionality to record a make
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

// Struct to store an make
// For now only channel makes are recorded
//
// Fields
//   - tPost int64: time when the operation finished
//   - id string: id of the channel
//   - elemType string: type of created primitive, for now only "NC" for channel
//   - qSizeu int: for channel the buffer size, for all other type always 0
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceAlloc struct {
	tPost int64
	id    uint64
	qSize int
	file  string
	line  int
	op    Operation
}

/*
 * AdvocateAlloc adds a channel make to the trace.
 * Args:
 *  opjType string: object type. Must be qualt to string(Operation)
 *  qSizeum int: for channel the buffer size, for all other type always 0
 * Return:
 * 	(int): id for the channel
 */
func AdvocateAlloc(objType string, qSize int) uint64 {
	if advocateTracingDisabled {
		return 0
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	id := GetAdvocateObjectID()

	if AdvocateIgnore(file) {
		return id
	}

	elem := AdvocateTraceAlloc{
		tPost: timer,
		id:    id,
		file:  file,
		line:  line,
		qSize: qSize,
		op:    Operation(OperationAllocChan),
	}

	insertIntoTrace(elem)

	return id
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (elem AdvocateTraceAlloc) toString() string {
	operationType := ""

	switch elem.op {
	case OperationAllocChan:
		operationType = "C"
	case OperationAllocMutex:
		operationType = "M"
	case OperationAllocCond:
		operationType = "D"
	case OperationAllocWg:
		operationType = "W"
	}

	return buildTraceElemString("N", elem.tPost, elem.id, operationType, elem.qSize, posToString(elem.file, elem.line))
}

// Get the string representation for the primitive type
// For now always return "NC"
//
// Returns:
//   - string representation of the primitive type
func (elem AdvocateTraceAlloc) getOpStr() string {
	return "NC"
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceAlloc) getOperation() Operation {
	return elem.op
}
