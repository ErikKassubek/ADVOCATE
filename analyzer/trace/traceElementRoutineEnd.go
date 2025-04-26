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
	"analyzer/clock"
	"errors"
	"strconv"
)

// TraceElementRoutineEnd is a trace element for the termination of a routine end
// Fields:
//
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp at the end of the event
//   - vc clock.VectorClock: The vector clock
type TraceElementRoutineEnd struct {
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

	elem := TraceElementRoutineEnd{
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
func (re *TraceElementRoutineEnd) GetID() int {
	return 0
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (re *TraceElementRoutineEnd) GetRoutine() int {
	return re.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (re *TraceElementRoutineEnd) GetTPre() int {
	return re.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (re *TraceElementRoutineEnd) GetTPost() int {
	return re.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (re *TraceElementRoutineEnd) GetTSort() int {
	return re.tPost
}

// GetPos is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (re *TraceElementRoutineEnd) GetPos() string {
	return ""
}

// GetReplayID is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (re *TraceElementRoutineEnd) GetReplayID() string {
	return ""
}

// GetFile is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (re *TraceElementRoutineEnd) GetFile() string {
	return ""
}

// GetLine is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (re *TraceElementRoutineEnd) GetLine() int {
	return 0
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (re *TraceElementRoutineEnd) GetTID() string {
	return ""
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (re *TraceElementRoutineEnd) SetVc(vc *clock.VectorClock) {
	re.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (re *TraceElementRoutineEnd) SetWVc(vc *clock.VectorClock) {
	re.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (re *TraceElementRoutineEnd) GetVC() *clock.VectorClock {
	return re.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (re *TraceElementRoutineEnd) GetWVc() *clock.VectorClock {
	return re.wVc
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (re *TraceElementRoutineEnd) GetObjType(operation bool) string {
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
func (re *TraceElementRoutineEnd) IsEqual(elem TraceElement) bool {
	return re.routine == elem.GetRoutine() && re.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (re *TraceElementRoutineEnd) GetTraceIndex() (int, int) {
	return re.routine, re.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (re *TraceElementRoutineEnd) SetT(time int) {
	re.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (re *TraceElementRoutineEnd) SetTPre(tPre int) {
	re.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (re *TraceElementRoutineEnd) SetTSort(tPost int) {
	re.SetTPre(tPost)
	re.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (re *TraceElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
	re.SetTPre(tSort)
	if re.tPost != 0 {
		re.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (re *TraceElementRoutineEnd) ToString() string {
	return "E" + "," + strconv.Itoa(re.tPost)
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (re *TraceElementRoutineEnd) Copy() TraceElement {
	return &TraceElementRoutineEnd{
		index:   re.index,
		routine: re.routine,
		tPost:   re.tPost,
		vc:      re.vc.Copy(),
		wVc:     re.wVc.Copy(),
	}
}

// AddRel1 is a dummy function to implement the traceElement interface
func (re *TraceElementRoutineEnd) AddRel1(_ TraceElement, _ int) {
	return
}

// AddRel2 is a dummy function to implement the traceElement interface
func (re *TraceElementRoutineEnd) AddRel2(_ TraceElement) {
	return
}

// GetRel1 is a dummy function to implement the traceElement interface
func (re *TraceElementRoutineEnd) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

// GetRel2 is a dummy function to implement the traceElement interface
func (re *TraceElementRoutineEnd) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
