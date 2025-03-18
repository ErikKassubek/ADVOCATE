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

