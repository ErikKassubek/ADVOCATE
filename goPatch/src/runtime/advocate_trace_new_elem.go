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
//   - num int: additional numerical field
//     for channel: qSize
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceNewElem struct {
	tPost    int64
	id       uint64
	elemType string
	num      int
	file     string
	line     int
	op       Operation
}

/*
 * AdvocateChanMake adds a channel make to the trace.
 * Args:
 * 	id: id of the channel
 * 	qSize: size of the channel
 * Return:
 * 	(int): id for the channel
 */
func AdvocateChanMake(qSize int) uint64 {
	if advocateTracingDisabled {
		return 0
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	id := GetAdvocateObjectID()

	if AdvocateIgnore(file) {
		return id
	}

	elem := AdvocateTraceNewElem{
		tPost:    timer,
		id:       id,
		elemType: "NC",
		num:      qSize,
		file:     file,
		line:     line,
		op:       OperationNewChan,
	}

	insertIntoTrace(elem)

	return id
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (elem AdvocateTraceNewElem) toString() string {
	operationType := "NC"
	return buildTraceElemString("N", elem.tPost, elem.id, operationType, elem.num, posToString(elem.file, elem.line))
}

// Get the string representation for the primitive type
// For now always return "NC"
//
// Returns:
//   - string representation of the primitive type
func (elem AdvocateTraceNewElem) getOpStr() string {
	return "NC"
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceNewElem) getOperation() Operation {
	return elem.op
}
