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
	"advocate/analysis/hb/clock"
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
func (this *Trace) AddTraceElementReplay(ts int, exitCode int) error {
	elem := ElementReplay{
		tPost:    ts,
		exitCode: exitCode,
	}

	this.AddElement(&elem)

	return nil
}

// Get the ElemMin representation of the operation
//
// Returns:
//   - ElemMin: the ElemMin representations of the operation
//   - bool: true if it should be part of a min trace, false otherwise
func (this *ElementReplay) GetElemMin() (ElemMin, bool) {
	return ElemMin{
		ID:      -1,
		Op:      None,
		Pos:     "",
		Routine: -1,
		Vc:      *clock.NewVectorClock(0),
	}, false
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementReplay) GetID() int {
	return 0
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementReplay) GetRoutine() int {
	return 1
}

// GetTPre returns the tPre of the element.
//
//   - int: The tPost of the element
func (this *ElementReplay) GetTPre() int {
	return this.tPost
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (this *ElementReplay) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementReplay) GetTSort() int {
	return this.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (this *ElementReplay) GetPos() string {
	return ""
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementReplay) GetReplayID() string {
	return ""
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementReplay) GetFile() string {
	return ""
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementReplay) GetLine() int {
	return 0
}

// GetTID returns the tID of the element.
// The tID is normally a string of form [file]:[line]@[tPre]
// Since the replay element is not used for any analysis, it returns an empty string
//
// Returns:
//   - string: The tID of the element
func (this *ElementReplay) GetTID() string {
	return ""
}

// SetVc is a dummy function to implement the TraceElement interface
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementReplay) SetVc(_ *clock.VectorClock) {
}

// SetWVc is a dummy function to implement the TraceElement interface
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementReplay) SetWVc(_ *clock.VectorClock) {
}

// GetVC is a dummy function to implement the TraceElement interface
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementReplay) GetVC() *clock.VectorClock {
	return &clock.VectorClock{}
}

// GetWVC is a dummy function to implement the TraceElement interface
func (this *ElementReplay) GetWVC() *clock.VectorClock {
	return &clock.VectorClock{}
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementReplay) GetType(operation bool) OperationType {
	if !operation {
		return Replay
	}
	return ReplayOP
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementReplay) IsSameElement(elem Element) bool {
	return false
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementReplay) IsEqual(elem Element) bool {
	return this.ToString() == elem.ToString()
}

// GetTraceIndex returns the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementReplay) GetTraceIndex() (int, int) {
	return -1, -1
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementReplay) SetT(time int) {
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementReplay) SetTPre(tPre int) {
	tPre = max(1, tPre)
	this.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementReplay) SetTSort(tSort int) {
	tSort = max(1, tSort)
	this.SetTPre(tSort)
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementReplay) SetTWithoutNotExecuted(tSort int) {
	tSort = max(1, tSort)
	this.SetTPre(tSort)
	this.tPost = tSort
}

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementReplay) ToString() string {
	res := "X," + strconv.Itoa(this.tPost) + "," + strconv.Itoa(this.exitCode)
	return res
}

// UpdateVectorClock update and stores the vector clock of the element
func (this *ElementReplay) UpdateVectorClock() {
	// nothing to do
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementReplay) GetTraceID() int {
	return this.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementReplay) setTraceID(ID int) {
	this.traceID = ID
}

// Copy creates a copy of the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since conds do not contain reference to other elements and no other
//     elements contain referents to conds, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementReplay) Copy(_ map[string]Element) Element {
	return &ElementReplay{
		traceID:  this.traceID,
		tPost:    this.tPost,
		exitCode: this.exitCode,
	}
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - int: -1
func (this *ElementReplay) GetNumberConcurrent(_, _ bool) int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
func (this *ElementReplay) SetNumberConcurrent(_ int, _, _ bool) {}

// GetConcurrent returns the elements that are concurrent to the element
//
// Parameter:
//   - weak bool: get number of weak concurrent
//
// Returns:
//   - []Element: empty
func (this *ElementReplay) GetConcurrent(_, _ bool) []Element {
	return []Element{}
}

// SetConcurrent sets the concurrent elements
func (this *ElementReplay) SetConcurrent(_ []Element, _, _ bool) {
}
