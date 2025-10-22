// Copyright (c) 2024 Erik Kassubek
//
// File: TraceElementRoutineEnd.go
// Brief: Struct and functions for fork operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"errors"
	"strconv"
)

// ElementRoutineEnd is a trace element for the termination of a routine end
// Fields:
//   - traceID: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp at the end of the event
//   - vc clock.VectorClock: The vector clock
type ElementRoutineEnd struct {
	traceID int
	index   int
	routine int
	tPost   int
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
}

// AddTraceElementRoutineEnd add a routine and element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the new routine
//   - pos string: The position of the trace element in the file
func (this *Trace) AddTraceElementRoutineEnd(routine int, tPost string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	elem := ElementRoutineEnd{
		index:   this.numberElemsInTrace[routine],
		routine: routine,
		tPost:   tPostInt,
		vc:      nil,
		wVc:     nil,
	}

	this.AddElement(&elem)

	return nil
}

// Get the ElemMin representation of the operation
//
// Returns:
//   - ElemMin: the ElemMin representations of the operation
//   - bool: true if it should be part of a min trace, false otherwise
func (this *ElementRoutineEnd) GetElemMin() (ElemMin, bool) {
	return ElemMin{
		ID:      -1,
		Op:      EndRoutine,
		Pos:     "",
		Routine: this.routine,
	}, false
}

// GetID is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (this *ElementRoutineEnd) GetID() int {
	return 0
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementRoutineEnd) GetRoutine() int {
	return this.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (this *ElementRoutineEnd) GetTPre() int {
	return this.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (this *ElementRoutineEnd) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementRoutineEnd) GetTSort() int {
	return this.tPost
}

// GetPos is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) GetPos() string {
	return ""
}

// GetReplayID is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) GetReplayID() string {
	return ""
}

// GetFile is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) GetFile() string {
	return ""
}

// GetLine is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (this *ElementRoutineEnd) GetLine() int {
	return 0
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementRoutineEnd) GetTID() string {
	return ""
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementRoutineEnd) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementRoutineEnd) SetWVc(vc *clock.VectorClock) {
	this.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementRoutineEnd) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementRoutineEnd) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (this *ElementRoutineEnd) GetType(operation bool) ObjectType {
	if !operation {
		return End
	}
	return EndRoutine
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementRoutineEnd) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementRoutineEnd) IsSameElement(elem Element) bool {
	return false
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementRoutineEnd) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementRoutineEnd) SetT(time int) {
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementRoutineEnd) SetTPre(tPre int) {
	this.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementRoutineEnd) SetTSort(tPost int) {
	this.SetTPre(tPost)
	this.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementRoutineEnd) ToString() string {
	return "E" + "," + strconv.Itoa(this.tPost)
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementRoutineEnd) GetTraceID() int {
	return this.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementRoutineEnd) setTraceID(ID int) {
	this.traceID = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since conds do not contain reference to other elements and no other
//     elements contain referents to conds, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementRoutineEnd) Copy(_ map[string]Element) Element {
	return &ElementRoutineEnd{
		traceID: this.traceID,
		index:   this.index,
		routine: this.routine,
		tPost:   this.tPost,
		vc:      this.vc.Copy(),
		wVc:     this.wVc.Copy(),
	}
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
func (this *ElementRoutineEnd) GetNumberConcurrent(_, _ bool) int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
func (this *ElementRoutineEnd) SetNumberConcurrent(_ int, _, _ bool) {}

// GetConcurrent returns the elements that are concurrent to the element
func (this *ElementRoutineEnd) GetConcurrent(_, _ bool) []Element {
	return []Element{}
}

// SetConcurrent sets the concurrent elements
func (this *ElementRoutineEnd) SetConcurrent(_ []Element, _, _ bool) {
}
