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

// MARK: Pre

/*
 * AdvocateChanPre adds a channel send/receive to the trace.
 * Args:
 * 	id: id of the channel
 * 	op: operation send/recv
 * 	opId: id of the operation
 * 	qSize: size of the channel, 0 for unbuffered
 * 	isNil: true if the channel is nil
 * Return:
 * 	index of the operation in the trace, return -1 if it is a atomic operation
 */
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

// MARK: Close

/*
 * AdvocateChanClose adds a channel close to the trace
 * Args:
 * 	id: id of the channel
 * Return:
 * 	index of the operation in the trace
 */
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

// MARK: Post

/*
 * AdvocateChanPost sets the operation as successfully finished
 * Args:
 * 	index: index of the operation in the trace
 * 	qCount: number of elements in the queue after the operations has finished
 */
func AdvocateChanPost(index int, qCount uint) {
	if advocateTracingDisabled {
		return
	}

	time := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutine().getElement(index).(AdvocateTraceChannel)

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

	currentGoRoutine().updateElement(index, elem)
}

/*
 * AdvocateChanPostCausedByClose sets the operation as successfully finished
 * Args:
 * 	index: index of the operation in the trace
 */
func AdvocateChanPostCausedByClose(index int) {
	if advocateTracingDisabled {
		return
	}

	time := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutine().getElement(index).(AdvocateTraceChannel)

	elem.tPost = time
	elem.cl = true

	currentGoRoutine().updateElement(index, elem)
}

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

	return buildTraceElemStringSep(".", "C", elem.tPre, elem.tPost, idStr, opStr, elem.cl, elem.oId, elem.qSize, elem.qCount)
}

func (elem AdvocateTraceChannel) getOperation() Operation {
	return elem.op
}
