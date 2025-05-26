// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_channel.go
// Brief: Functionality for the channel
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

var unbufferedChannelComSend = make(map[uint64]int64) // id -> tpost
var unbufferedChannelComRecv = make(map[uint64]int64) // id -> tpost
var unbufferedChannelComSendMutex mutex
var unbufferedChannelComRecvMutex mutex

// Struct to store an operation on a channel variable
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - id string: id of the channel
//   - op Operation: operation type
//   - cl bool: true if the recv was executed because the channel is closed
//   - oId uint64: operation id, communicating send and recv have the same oID
//   - qSize uint: size of the channel buffer
//   - qCount uint: number of element in the buffer after the operation finished
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
//   - isNil bool: true if the channel is nil
type AdvocateTraceChannel struct {
	tPre   int64
	tPost  int64
	id     uint64
	op     Operation
	cl     bool
	oId    uint64
	qSize  uint
	qCount uint
	file   string
	line   int
	isNil  bool
}

// AdvocateChanPre adds a channel send/receive to the trace.
//
// Parameters:
//   - id uint64: id of the channel
//   - op Operation: operation send/recv
//   - opId opID: id of the operation
//   - qSize uint: size of the channel, 0 for unbuffered
//   - isNil bool: true if the channel is nil
//
// Returns:
//   - int: index of the operation in the trace, return -1 if it is a atomic operation
func AdvocateChanPre(id uint64, op Operation, opID uint64, qSize uint, isNil bool) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(3)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := AdvocateTraceChannel{
		tPre:  timer,
		id:    id,
		op:    op,
		oId:   opID,
		qSize: qSize,
		file:  file,
		line:  line,
		isNil: isNil,
	}

	return insertIntoTrace(elem)
}

// AdvocateChanClose adds a channel close to the trace
//
// Parameter:
//   - id uint64: id of the channel
//   - qSize uint: size of the buffer
//   - qCount uint: number of messages in the buffer
//
// Returns:
//   - index of the operation in the trace
func AdvocateChanClose(id uint64, qSize uint, qCount uint) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)
	if AdvocateIgnore(file) {
		return -1
	}

	elem := AdvocateTraceChannel{
		tPre:   timer,
		tPost:  timer,
		id:     id,
		op:     OperationChannelClose,
		qSize:  qSize,
		qCount: qCount,
		file:   file,
		line:   line,
	}

	return insertIntoTrace(elem)
}

// AdvocateChanPost sets the operation as successfully finished
//
// Parameters:
//   - index: index of the operation in the trace
//   - qCount: number of elements in the queue after the operations has finished
func AdvocateChanPost(index int, qCount uint) {
	if advocateTracingDisabled {
		return
	}

	time := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceChannel)

	set := false

	if elem.qSize == 0 { // unbuffered channel
		if elem.op == OperationChannelSend {
			lock(&unbufferedChannelComRecvMutex)
			if tpost, ok := unbufferedChannelComRecv[elem.id]; ok {
				elem.tPost = tpost - 1
				delete(unbufferedChannelComRecv, elem.id)
			} else {
				elem.tPost = time
				lock(&unbufferedChannelComSendMutex)
				unbufferedChannelComSend[elem.id] = time
				unlock(&unbufferedChannelComSendMutex)
			}
			unlock(&unbufferedChannelComRecvMutex)
			set = true
		} else if elem.op == OperationChannelSend {
			lock(&unbufferedChannelComSendMutex)
			if tpost, ok := unbufferedChannelComSend[elem.id]; ok {
				elem.tPost = tpost + 1
				delete(unbufferedChannelComSend, elem.id)
			} else {
				elem.tPost = time
				unbufferedChannelComRecv[elem.id] = time
			}
			unlock(&unbufferedChannelComSendMutex)
			set = true
		}
	}

	if !set {
		elem.tPost = time
	}
	elem.qCount = qCount

	currentGoRoutineInfo().updateElement(index, elem)
}

// AdvocateChanPostCausedByClose sets the operation as successfully finished
// Args:
//   - index: index of the operation in the trace
func AdvocateChanPostCausedByClose(index int) {
	if advocateTracingDisabled {
		return
	}

	time := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceChannel)

	elem.tPost = time
	elem.cl = true

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     C,[tPre],[tPost],[id],[operation],[cl],[oId],[qSize],[qCount],[file],[line]
func (elem AdvocateTraceChannel) toString() string {
	opStr := ""
	switch elem.op {
	case OperationChannelSend:
		opStr = "S"
	case OperationChannelRecv:
		opStr = "R"
	case OperationChannelClose:
		opStr = "C"
	}

	idStr := "*"
	if !elem.isNil {
		idStr = uint64ToString(elem.id)
	}

	return buildTraceElemString("C", elem.tPre, elem.tPost, idStr, opStr, elem.cl, elem.oId, elem.qSize, elem.qCount, posToString(elem.file, elem.line))
}

// Get a string representation for the channel if it is used as a select case
//
// Returns:
//   - string: the string representation of the form
//     C,[id].[operation].[cl].[oId].[qSize].[qCount]
func (elem AdvocateTraceChannel) toStringForSelect() string {
	opStr := ""
	switch elem.op {
	case OperationChannelSend:
		opStr = "S"
	case OperationChannelRecv:
		opStr = "R"
	case OperationChannelClose:
		opStr = "C"
	}

	idStr := "*"
	if !elem.isNil {
		idStr = uint64ToString(elem.id)
	}

	return buildTraceElemStringSep(".", "C", idStr, opStr, elem.cl, elem.oId, elem.qSize, elem.qCount)
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceChannel) getOperation() Operation {
	return elem.op
}
