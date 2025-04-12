// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementReplayStop.go
// Brief: Struct and functions for replay controll elements in the trace
//
// Author: Erik Kassubek
// Created: 2024-04-03
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"strconv"
)

// Struct to save an atomic event in the trace
// Fields:
//
//   - tpost (int): The timestamp of the event
//   - exitCode (int): expected exit code
type TraceElementReplay struct {
	tPost    int
	exitCode int
}

// Create an end of replay event
//
// Parameter:
//   - t (string): The timestamp of the event
//   - exitCode (int): The exit code of the event
func AddTraceElementReplay(t int, exitCode int) error {
	return MainTrace.AddTraceElementReplay(t, exitCode)
}

// Get the id of the element
//
// Returns:
//   - int: The id of the element
func (er *TraceElementReplay) GetID() int {
	return 0
}

// Get the routine of the element
//
// Returns:
//   - int: The routine of the element
func (er *TraceElementReplay) GetRoutine() int {
	return 1
}

// Get the tpost of the element.
//
//   - int: The tpost of the element
func (er *TraceElementReplay) GetTPre() int {
	return er.tPost
}

// Get the tpost of the element.
//
// Returns:
//   - int: The tpost of the element
func (er *TraceElementReplay) GetTPost() int {
	return er.tPost
}

// Get the timer, ther is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (er *TraceElementReplay) GetTSort() int {
	return er.tPost
}

// Get the position of the operation.
//
// Returns:
//   - string: The file of the element
func (er *TraceElementReplay) GetPos() string {
	return ""
}

func (er *TraceElementReplay) GetReplayID() string {
	return ""
}

func (er *TraceElementReplay) GetFile() string {
	return ""
}

func (er *TraceElementReplay) GetLine() int {
	return 0
}

// Get the tID of the element.
//
// Returns:
//   - string: The tID of the element
func (er *TraceElementReplay) GetTID() string {
	return ""
}

// Dummy function to implement the interface
//
// Returns:
//   - VectorClock: The vector clock of the element
func (er *TraceElementReplay) GetVC() *clock.VectorClock {
	return &clock.VectorClock{}
}

// GetwVc implements TraceElement.
func (at *TraceElementReplay) GetwVc() *clock.VectorClock {
	return &clock.VectorClock{}
}

// Get the string representation of the object type
func (er *TraceElementReplay) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeReplay + "R"
	}
	return ObjectTypeReplay
}

func (er *TraceElementReplay) IsEqual(elem TraceElement) bool {
	return er.ToString() == elem.ToString()
}

func (er *TraceElementReplay) GetTraceIndex() (int, int) {
	return -1, -1
}

// Set the tPre and tPost of the element
//
// Parameter:
//   - time (int): The tPre and tPost of the element
func (mu *TraceElementReplay) SetT(time int) {
	mu.tPost = time
}

// Set the tpre of the element.
//
// Parameter:
//   - tPre (int): The tpre of the element
func (mu *TraceElementReplay) SetTPre(tPre int) {
	tPre = max(1, tPre)
	mu.tPost = tPre
}

// Set the timer, ther is used for the sorting of the trace
//
// Parameter:
//   - tSort (int): The timer of the element
func (er *TraceElementReplay) SetTSort(tSort int) {
	tSort = max(1, tSort)
	er.SetTPre(tSort)
	er.tPost = tSort
}

// Set the timer, ther is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort (int): The timer of the element
func (er *TraceElementReplay) SetTWithoutNotExecuted(tSort int) {
	tSort = max(1, tSort)
	er.SetTPre(tSort)
	er.tPost = tSort
}

// Get the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (er *TraceElementReplay) ToString() string {
	res := "X," + strconv.Itoa(er.tPost) + "," + strconv.Itoa(er.exitCode)
	return res
}

// Update and calculate the vector clock of the element
func (er *TraceElementReplay) updateVectorClock() {
	// nothing to do
}

// Create a copy of the element
//
// Returns:
//   - TraceElement: The copy of the element
func (er *TraceElementReplay) Copy() TraceElement {
	return &TraceElementReplay{
		tPost:    er.tPost,
		exitCode: er.exitCode,
	}
}

// Dummy function for traceElement
func (er *TraceElementReplay) AddRel1(_ TraceElement, _ int) {
	return
}

// Dummy function for traceElement
func (er *TraceElementReplay) AddRel2(_ TraceElement) {
	return
}

// Dummy function for traceElement
func (er *TraceElementReplay) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

// Dummy function for traceElement
func (er *TraceElementReplay) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
