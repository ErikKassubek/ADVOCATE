// GOCP-FILE_START

// File: goCR_trace_new_elem.go
// Brief: Functionality to record a make
//
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
type GoCRTraceNewElem struct {
	tPost    int64
	id       uint64
	elemType string
	num      int
	file     string
	line     int
	op       Operation
}

/*
 * GoCRChanMake adds a channel make to the trace.
 * Args:
 * 	id: id of the channel
 * 	qSize: size of the channel
 * Return:
 * 	(int): id for the channel
 */
func GoCRChanMake(qSize int) uint64 {
	if goCRTracingDisabled {
		return 0
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	id := GetGoCRObjectID()

	elem := GoCRTraceNewElem{
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
func (elem GoCRTraceNewElem) toString() string {
	operationType := "NC"
	return buildTraceElemString("N", elem.tPost, elem.id, operationType, elem.num, posToString(elem.file, elem.line))
}

// Get the string representation for the primitive type
// For now always return "NC"
//
// Returns:
//   - string representation of the primitive type
func (elem GoCRTraceNewElem) getOpStr() string {
	return "NC"
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem GoCRTraceNewElem) getOperation() Operation {
	return elem.op
}
