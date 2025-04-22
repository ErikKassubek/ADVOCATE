// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementReplayStop.go
// Brief: Struct and functions for replay control elements in the trace
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

// TraceElementReplay is a struct to save an end of replay marker in the trace
// Fields:
//
//   - tPost int: The timestamp of the event
//   - controlCode string: control code or expected exit code
type TraceElementReplay struct {
	tPost       int
	controlCode string
}

// AddTraceElementReplayExitCode adds an end of replay event to the main trace
//
// Parameter:
//   - t string: The timestamp of the event
//   - exitCode int: The exit code
func AddTraceElementReplayExitCode(t int, exitCode int) error {
	return MainTrace.AddTraceElementReplayExitCode(t, exitCode)
}

// AddTraceElementReplayControlCode adds an end of replay event to the main trace
//
// Parameter:
//   - t string: The timestamp of the event
//   - controlCode string: The control code
func AddTraceElementReplayControlCode(t int, controlCode string) error {
	return MainTrace.AddTraceElementReplayControlCode(t, controlCode)
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (er *TraceElementReplay) GetID() int {
	return 0
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (er *TraceElementReplay) GetRoutine() int {
	return 1
}

// GetTPre returns the tPre of the element.
//
//   - int: The tPost of the element
func (er *TraceElementReplay) GetTPre() int {
	return er.tPost
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (er *TraceElementReplay) GetTPost() int {
	return er.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (er *TraceElementReplay) GetTSort() int {
	return er.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (er *TraceElementReplay) GetPos() string {
	return ""
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (er *TraceElementReplay) GetReplayID() string {
	return ""
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (er *TraceElementReplay) GetFile() string {
	return ""
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (er *TraceElementReplay) GetLine() int {
	return 0
}

// GetTID returns the tID of the element.
// The tID is normally a string of form [file]:[line]@[tPre]
// Since the replay element is not used for any analysis, it returns an empty string
//
// Returns:
//   - string: The tID of the element
func (er *TraceElementReplay) GetTID() string {
	return ""
}

// GetVC is a dummy function to implement the TraceElement interface
//
// Returns:
//   - VectorClock: The vector clock of the element
func (er *TraceElementReplay) GetVC() *clock.VectorClock {
	return &clock.VectorClock{}
}

// GetWVc is a dummy function to implement the TraceElement interface
func (er *TraceElementReplay) GetWVc() *clock.VectorClock {
	return &clock.VectorClock{}
}

// GetObjType returns the string representation of the object type
func (er *TraceElementReplay) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeReplay + "R"
	}
	return ObjectTypeReplay
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (er *TraceElementReplay) IsEqual(elem TraceElement) bool {
	return er.ToString() == elem.ToString()
}

// GetTraceIndex returns the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (er *TraceElementReplay) GetTraceIndex() (int, int) {
	return -1, -1
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (er *TraceElementReplay) SetT(time int) {
	er.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (er *TraceElementReplay) SetTPre(tPre int) {
	tPre = max(1, tPre)
	er.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (er *TraceElementReplay) SetTSort(tSort int) {
	tSort = max(1, tSort)
	er.SetTPre(tSort)
	er.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (er *TraceElementReplay) SetTWithoutNotExecuted(tSort int) {
	tSort = max(1, tSort)
	er.SetTPre(tSort)
	er.tPost = tSort
}

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (er *TraceElementReplay) ToString() string {
	res := "X," + strconv.Itoa(er.tPost) + "," + er.controlCode
	return res
}

// Update and calculate the vector clock of the element
func (er *TraceElementReplay) updateVectorClock() {
	// nothing to do
}

// Copy creates a copy of the element
//
// Returns:
//   - TraceElement: The copy of the element
func (er *TraceElementReplay) Copy() TraceElement {
	return &TraceElementReplay{
		tPost:       er.tPost,
		controlCode: er.controlCode,
	}
}

// AddRel1 is a dummy function to implement the traceElement interface
func (er *TraceElementReplay) AddRel1(_ TraceElement, _ int) {
	return
}

// AddRel2 is a dummy function to implement the traceElement interface
func (er *TraceElementReplay) AddRel2(_ TraceElement) {
	return
}

// GetRel1 is a dummy function to implement the traceElement interface
func (er *TraceElementReplay) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

// GetRel2 is a dummy function to implement the traceElement interface
func (er *TraceElementReplay) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
