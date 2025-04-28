// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementReplayStop.go
// Brief: Struct and functions for replay control elements in the trace
//
// Author: Erik Kassubek
// Created: 2024-04-03
//
// License: BSD-3-Clause

package trace

import (
	"analyzer/clock"
	"strconv"
)

// TraceElementReplay is a struct to save an end of replay marker in the trace
// Fields:
//   - traceID: id of the element, should never be changed
//   - tPost int: The timestamp of the event
//   - exitCode int: expected exit code
type TraceElementReplay struct {
	traceID  int
	tPost    int
	exitCode int
}

// AddTraceElementReplay adds an replay end element to a trace
//
// Parameter:
//   - ts string: The timestamp of the event
//   - exitCode int: The exit code of the event
//
// Returns:
//   - error
func (t *Trace) AddTraceElementReplay(ts int, exitCode int) error {
	elem := TraceElementReplay{
		tPost:    ts,
		exitCode: exitCode,
	}

	t.AddElement(&elem)

	return nil
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

// SetVc is a dummy function to implement the TraceElement interface
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (er *TraceElementReplay) SetVc(_ *clock.VectorClock) {
}

// SetWVc is a dummy function to implement the TraceElement interface
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (er *TraceElementReplay) SetWVc(_ *clock.VectorClock) {
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
	res := "X," + strconv.Itoa(er.tPost) + "," + strconv.Itoa(er.exitCode)
	return res
}

// UpdateVectorClock update and stores the vector clock of the element
func (er *TraceElementReplay) UpdateVectorClock() {
	// nothing to do
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (er *TraceElementReplay) GetTraceID() int {
	return er.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (er *TraceElementReplay) setTraceID(ID int) {
	er.traceID = ID
}

// Copy creates a copy of the element
//
// Returns:
//   - TraceElement: The copy of the element
func (er *TraceElementReplay) Copy() TraceElement {
	return &TraceElementReplay{
		traceID:  er.traceID,
		tPost:    er.tPost,
		exitCode: er.exitCode,
	}
}
