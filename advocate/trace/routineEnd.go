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
	"advocate/analysis/clock"
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
func (t *Trace) AddTraceElementRoutineEnd(routine int, tPost string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	elem := ElementRoutineEnd{
		index:   t.numberElemsInTrace[routine],
		routine: routine,
		tPost:   tPostInt,
		vc:      nil,
		wVc:     nil,
	}

	t.AddElement(&elem)

	return nil
}

// GetID is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (re *ElementRoutineEnd) GetID() int {
	return 0
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (re *ElementRoutineEnd) GetRoutine() int {
	return re.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (re *ElementRoutineEnd) GetTPre() int {
	return re.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (re *ElementRoutineEnd) GetTPost() int {
	return re.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (re *ElementRoutineEnd) GetTSort() int {
	return re.tPost
}

// GetPos is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (re *ElementRoutineEnd) GetPos() string {
	return ""
}

// GetReplayID is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (re *ElementRoutineEnd) GetReplayID() string {
	return ""
}

// GetFile is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (re *ElementRoutineEnd) GetFile() string {
	return ""
}

// GetLine is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (re *ElementRoutineEnd) GetLine() int {
	return 0
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (re *ElementRoutineEnd) GetTID() string {
	return ""
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (re *ElementRoutineEnd) SetVc(vc *clock.VectorClock) {
	re.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (re *ElementRoutineEnd) SetWVc(vc *clock.VectorClock) {
	re.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (re *ElementRoutineEnd) GetVC() *clock.VectorClock {
	return re.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (re *ElementRoutineEnd) GetWVc() *clock.VectorClock {
	return re.wVc
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (re *ElementRoutineEnd) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeRoutineEnd + "E"
	}
	return ObjectTypeRoutineEnd
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (re *ElementRoutineEnd) IsEqual(elem Element) bool {
	return re.routine == elem.GetRoutine() && re.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (re *ElementRoutineEnd) GetTraceIndex() (int, int) {
	return re.routine, re.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (re *ElementRoutineEnd) SetT(time int) {
	re.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (re *ElementRoutineEnd) SetTPre(tPre int) {
	re.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (re *ElementRoutineEnd) SetTSort(tPost int) {
	re.SetTPre(tPost)
	re.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (re *ElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
	re.SetTPre(tSort)
	if re.tPost != 0 {
		re.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (re *ElementRoutineEnd) ToString() string {
	return "E" + "," + strconv.Itoa(re.tPost)
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (re *ElementRoutineEnd) GetTraceID() int {
	return re.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (re *ElementRoutineEnd) setTraceID(ID int) {
	re.traceID = ID
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (re *ElementRoutineEnd) Copy() Element {
	return &ElementRoutineEnd{
		traceID: re.traceID,
		index:   re.index,
		routine: re.routine,
		tPost:   re.tPost,
		vc:      re.vc.Copy(),
		wVc:     re.wVc.Copy(),
	}
}

// AddChild adds an element as a child of this node in the partial order graph
// Dummy method to implement TraceElement interface
//
// Parameter:
//   - elem *TraceElement: the element to add
func (re *ElementRoutineEnd) AddChild(elem Element) {
}

// GetChildren returns all children of this node in the partial order graph
// Dummy method to implement TraceElement interface
//
// Returns:
//   - []*TraceElement: the children
func (re *ElementRoutineEnd) GetChildren() []Element {
	return make([]Element, 0)
}

// GetParents returns all parents of this node in the partial order graph
// Dummy method to implement TraceElement
//
// Returns:
//   - []*TraceElement: the parents
func (re *ElementRoutineEnd) GetParents() []Element {
	return make([]Element, 0)
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - number of concurrent element, or -1
func (at *ElementRoutineEnd) GetNumberConcurrent() int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
func (at *ElementRoutineEnd) SetNumberConcurrent(c int) {}
