// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_cond.go
// Brief: Functionality for the conditional variables
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause


package runtime

/*
 * AdvocateCondPre adds a cond wait to the trace
 * MARK: Pre
 * Args:
 * 	id: id of the cond
 * 	op: 0 for wait, 1 for signal, 2 for broadcast
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateCondPre(id uint64, op int) int {
	timer := GetNextTimeStep()
	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	var opC string
	switch op {
	case 0:
		opC = "W"
	case 1:
		opC = "S"
	case 2:
		opC = "B"
	default:
		panic("Unknown cond operation")
	}

	elem := "D," + uint64ToString(timer) + ",0," + uint64ToString(id) +
		"," + opC + "," + file + ":" + uint64ToString(uint64(line))
	return insertIntoTrace(elem)
}

/*
 * AdvocateCondPost adds the end counter to an operation of the trace
 * MARK: Post
 * Args:
 * 	index: index of the operation in the trace
 */
func AdvocateCondPost(index int) {
	timer := GetNextTimeStep()
	if index == -1 {
		return
	}
	elem := currentGoRoutine().getElement(index)

	split := splitStringAtCommas(elem, []int{2, 3})
	split[1] = uint64ToString(timer)

	elem = mergeString(split)

	currentGoRoutine().updateElement(index, elem)
}
