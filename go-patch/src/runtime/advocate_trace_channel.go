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

var advocateCounterAtomic uint64

var unbufferedChannelComSend = make(map[string]uint64) // id -> tpost
var unbufferedChannelComRecv = make(map[string]uint64) // id -> tpost
var unbufferedChannelComSendMutex mutex
var unbufferedChannelComRecvMutex mutex

// MARK: Make
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

	elem := "N," + uint64ToString(timer) + "," + uint64ToString(id) + ",NC," + intToString(qSize) + "," + file + ":" + intToString(line)

	insertIntoTrace(elem)

	return id
}

// MARK: Pre

/*
 * AdvocateChanSendPre adds a channel send to the trace.
 * Args:
 * 	id: id of the channel
 * 	opId: id of the operation
 * 	qSize: size of the channel, 0 for unbuffered
 * 	isNil: true if the channel is nil
 * Return:
 * 	index of the operation in the trace, return -1 if it is a atomic operation
 */
func AdvocateChanSendPre(id uint64, opID uint64, qSize uint, isNil bool) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(3)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := "C," + uint64ToString(timer) + ",0,"
	if isNil {
		elem += "*,S,f,0,0,0," + file + ":" + intToString(line)
	} else {
		elem += uint64ToString(id) + ",S,f," +
			uint64ToString(opID) + "," + uint32ToString(uint32(qSize)) + ",0," +
			file + ":" + intToString(line)
	}

	return insertIntoTrace(elem)
}

/*
 * AdvocateChanRecvPre adds a channel recv to the trace
 * Args:
 * 	id: id of the channel
 * 	opId: id of the operation
 * 	qSize: size of the channel
 * 	qCount: number of elems in queue after q
 * 	isNil: true if the channel is nil
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateChanRecvPre(id uint64, opID uint64, qSize uint, isNil bool) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	if !isNil && id == 0 {
		panic("A")
	}

	_, file, line, _ := Caller(3)
	if AdvocateIgnore(file) {
		return -1
	}

	elem := "C," + uint64ToString(timer) + ",0,"
	if isNil {
		elem += "*,R,f,0,0,0," + file + ":" + intToString(line)
	} else {
		elem += uint64ToString(id) + ",R,f," +
			uint64ToString(opID) + "," + uint32ToString(uint32(qSize)) + ",0," +
			file + ":" + intToString(line)
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

	timer := uint64ToString(GetNextTimeStep())

	_, file, line, _ := Caller(2)
	if AdvocateIgnore(file) {
		return -1
	}

	elem := "C," + timer + "," + timer + "," + uint64ToString(id) + ",C,f,0," +
		uint32ToString(uint32(qSize)) + "," + uint32ToString(uint32(qCount)) + "," + file + ":" + intToString(line)

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

	elem := currentGoRoutine().getElement(index)

	split := splitStringAtCommas(elem, []int{2, 3, 4, 5, 7, 8, 9})

	id := split[2]
	op := split[3]
	qSize := split[5]
	set := false

	if qSize == "0" { // unbuffered channel
		if op == "S" {
			lock(&unbufferedChannelComRecvMutex)
			if tpost, ok := unbufferedChannelComRecv[id]; ok {
				split[1] = uint64ToString(tpost - 1)
				delete(unbufferedChannelComRecv, id)
			} else {
				split[1] = uint64ToString(time)
				lock(&unbufferedChannelComSendMutex)
				unbufferedChannelComSend[id] = time
				unlock(&unbufferedChannelComSendMutex)
			}
			unlock(&unbufferedChannelComRecvMutex)
			set = true
		} else if op == "R" {
			lock(&unbufferedChannelComSendMutex)
			if tpost, ok := unbufferedChannelComSend[id]; ok {
				split[1] = uint64ToString(tpost + 1)
				delete(unbufferedChannelComSend, id)
			} else {
				split[1] = uint64ToString(time)
				unbufferedChannelComRecv[id] = time
			}
			unlock(&unbufferedChannelComSendMutex)
			set = true
		}
	}

	if !set {
		split[1] = uint64ToString(time)
	}

	split[6] = uint64ToString(uint64(qCount))

	elem = mergeString(split)

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

	elem := currentGoRoutine().getElement(index)
	split := splitStringAtCommas(elem, []int{2, 3, 5, 6})
	split[1] = uint64ToString(time)
	split[3] = "t"
	elem = mergeString(split)

	currentGoRoutine().updateElement(index, elem)
}
