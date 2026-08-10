// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_trace_new_elem.go
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
//   - res GoCCTTraceResource: the resource the op is applied to
//   - op Opertions: type of created primitive
//   - qSizeu int: for channel the buffer size, for all other type always 0
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCCTTraceAlloc struct {
	t     int64
	res   GoCCTTraceResource
	qSize int
	file  string
	line  int
	op    Operation
}

/*
 * GoCCTAlloc adds a channel make to the trace.
 * Args:
 *  opjType string: object type. Must be qualt to string(Operation)
 *  qSizeum int: for channel the buffer size, for all other type always 0
 * Return:
 * 	(int): id for the channel
 */
func GoCCTAlloc(objType string, qSize int) uint64 {
	if GoCCTTracingDisabled {
		return 0
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	id := GetGoCCTObjectID()

	if GoCCTIgnore(file) {
		return id
	}

	op := OperationNone
	switch objType {
	case "C":
		op = OperationAllocChan
	case "M":
		op = OperationAllocMutex
	case "D":
		op = OperationAllocCond
	case "W":
		op = OperationAllocWg
	}

	res := GoCCTTraceResource{id: id, addr: nil}

	elem := GoCCTTraceAlloc{
		t:     timer,
		res:   res,
		file:  file,
		line:  line,
		qSize: qSize,
		op:    op,
	}

	insertIntoTrace(elem)

	return id
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (self GoCCTTraceAlloc) toString() string {
	operationType := ""

	switch self.op {
	case OperationAllocChan:
		operationType = "C"
	case OperationAllocMutex:
		operationType = "M"
	case OperationAllocCond:
		operationType = "D"
	case OperationAllocWg:
		operationType = "W"
	}

	return buildTraceElemString("N", self.t, self.res.id, operationType, self.qSize, posToString(self.file, self.line))
}

// Get the string representation for the primitive type
// For now always return "NC"
//
// Returns:
//   - string representation of the primitive type
func (self GoCCTTraceAlloc) getOpStr() string {
	return "NC"
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self GoCCTTraceAlloc) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCCTTraceAlloc) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCCTTraceResource: recources
func (self GoCCTTraceAlloc) resource() []GoCCTTraceResource {
	return []GoCCTTraceResource{self.res}
}
