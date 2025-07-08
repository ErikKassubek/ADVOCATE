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
	"advocate/analysis/concurrent/clock"
	"strconv"
)

// ElementReplay is a struct to save an end of replay marker in the trace
// Fields:
//   - traceID: id of the element, should never be changed
//   - tPost int: The timestamp of the event
//   - exitCode int: expected exit code
type ElementReplay struct {
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
	elem := ElementReplay{
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
func (er *ElementReplay) GetID() int {
	return 0
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (er *ElementReplay) GetRoutine() int {
	return 1
}

// GetTPre returns the tPre of the element.
//
//   - int: The tPost of the element
func (er *ElementReplay) GetTPre() int {
	return er.tPost
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (er *ElementReplay) GetTPost() int {
	return er.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (er *ElementReplay) GetTSort() int {
	return er.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (er *ElementReplay) GetPos() string {
	return ""
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (er *ElementReplay) GetReplayID() string {
	return ""
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (er *ElementReplay) GetFile() string {
	return ""
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (er *ElementReplay) GetLine() int {
	return 0
}

// GetTID returns the tID of the element.
// The tID is normally a string of form [file]:[line]@[tPre]
// Since the replay element is not used for any analysis, it returns an empty string
//
// Returns:
//   - string: The tID of the element
func (er *ElementReplay) GetTID() string {
	return ""
}

// SetVc is a dummy function to implement the TraceElement interface
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (er *ElementReplay) SetVc(_ *clock.VectorClock) {
}

// SetWVc is a dummy function to implement the TraceElement interface
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (er *ElementReplay) SetWVc(_ *clock.VectorClock) {
}

// GetVC is a dummy function to implement the TraceElement interface
//
// Returns:
//   - VectorClock: The vector clock of the element
func (er *ElementReplay) GetVC() *clock.VectorClock {
	return &clock.VectorClock{}
}

// GetWVc is a dummy function to implement the TraceElement interface
func (er *ElementReplay) GetWVc() *clock.VectorClock {
	return &clock.VectorClock{}
}

// GetObjType returns the string representation of the object type
func (er *ElementReplay) GetObjType(operation bool) string {
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
func (er *ElementReplay) IsEqual(elem Element) bool {
	return er.ToString() == elem.ToString()
}

// GetTraceIndex returns the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (er *ElementReplay) GetTraceIndex() (int, int) {
	return -1, -1
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (er *ElementReplay) SetT(time int) {
	er.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (er *ElementReplay) SetTPre(tPre int) {
	tPre = max(1, tPre)
	er.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (er *ElementReplay) SetTSort(tSort int) {
	tSort = max(1, tSort)
	er.SetTPre(tSort)
	er.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (er *ElementReplay) SetTWithoutNotExecuted(tSort int) {
	tSort = max(1, tSort)
	er.SetTPre(tSort)
	er.tPost = tSort
}

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (er *ElementReplay) ToString() string {
	res := "X," + strconv.Itoa(er.tPost) + "," + strconv.Itoa(er.exitCode)
	return res
}

// UpdateVectorClock update and stores the vector clock of the element
func (er *ElementReplay) UpdateVectorClock() {
	// nothing to do
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (er *ElementReplay) GetTraceID() int {
	return er.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (er *ElementReplay) setTraceID(ID int) {
	er.traceID = ID
}

// Copy creates a copy of the element
//
// Returns:
//   - TraceElement: The copy of the element
func (er *ElementReplay) Copy() Element {
	return &ElementReplay{
		traceID:  er.traceID,
		tPost:    er.tPost,
		exitCode: er.exitCode,
	}
}

// AddChild adds an element as a child of this node in the partial order graph
// Dummy method to implement traceElement
//
// Parameter:
//   - elem *TraceElement: the element to add
func (er *ElementReplay) AddChild(elem Element) {
}

// AddParent adds an element as a parent of this node in the partial order graph
// Dummy method to implement traceElement
//
// Parameter:
//   - elem *TraceElement: the element to add
func (er *ElementReplay) AddParent(elem Element) {
}

// GetChildren returns all children of this node in the partial order graph
// Dummy method to implement traceElement
//
// Returns:
//   - []*TraceElement: the children
func (er *ElementReplay) GetChildren() []Element {
	return make([]Element, 0)
}

// GetParents returns all parents of this node in the partial order graph
// Dummy method to implement traceElement
//
// Returns:
//   - []*TraceElement: the parents
func (er *ElementReplay) GetParents() []Element {
	return make([]Element, 0)
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - number of concurrent element, or -1
func (er *ElementReplay) GetNumberConcurrent() int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
func (er *ElementReplay) SetNumberConcurrent(c int) {}
