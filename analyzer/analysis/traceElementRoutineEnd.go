// Copyright (c) 2024 Erik Kassubek
//
// File: TraceElementRoutineEnd.go
// Brief: Struct and functions for fork operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"errors"
	"strconv"
)

// TraceElementRoutineEnd is a trace element for the termination of a routine end
// Fields:
//
//   - index (int): Index in the routine
//   - routine (int): The routine id
//   - tpost (int): The timestamp at the end of the event
//   - vc (clock.VectorClock): The vector clock
type TraceElementRoutineEnd struct {
	index   int
	routine int
	tPost   int
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
}

// End a routine
//
// Parameter:
//   - routine (int): The routine id
//   - tPost (string): The timestamp at the end of the event
//   - id (string): The id of the new routine
//   - pos (string): The position of the trace element in the file
func AddTraceElementRoutineEnd(routine int, tPost string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	elem := TraceElementRoutineEnd{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPost:   tPostInt,
		vc:      clock.NewVectorClock(MainTrace.numberOfRoutines),
		wVc:     clock.NewVectorClock(MainTrace.numberOfRoutines),
	}

	AddElementToTrace(&elem)

	return nil
}

// Dummy function for traceElement
//
// Returns:
//   - int: 0
func (re *TraceElementRoutineEnd) GetID() int {
	return 0
}

// Get the routine of the element
//
// Returns:
//   - int: The routine of the element
func (re *TraceElementRoutineEnd) GetRoutine() int {
	return re.routine
}

// Get the tpre of the element. For atomic elements, tpre and tpost are the same
//
// Returns:
//   - int: The tpre of the element
func (re *TraceElementRoutineEnd) GetTPre() int {
	return re.tPost
}

// Get the tpost of the element. For atomic elements, tpre and tpost are the same
//
// Returns:
//   - int: The tpost of the element
func (re *TraceElementRoutineEnd) GetTPost() int {
	return re.tPost
}

// Get the timer, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (re *TraceElementRoutineEnd) GetTSort() int {
	return re.tPost
}

// Dummy function for traceElement
//
// Returns:
//   - string: empty string
func (re *TraceElementRoutineEnd) GetPos() string {
	return ""
}

// Dummy function for traceElement
//
// Returns:
//   - string: empty string
func (re *TraceElementRoutineEnd) GetReplayID() string {
	return ""
}

// Dummy function for traceElement
//
// Returns:
//   - string: empty string
func (re *TraceElementRoutineEnd) GetFile() string {
	return ""
}

// Dummy function for traceElement
//
// Returns:
//   - int: 0
func (re *TraceElementRoutineEnd) GetLine() int {
	return 0
}

// Get the tID of the element.
//
// Returns:
//   - string: The tID of the element
func (re *TraceElementRoutineEnd) GetTID() string {
	return ""
}

// Get the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (re *TraceElementRoutineEnd) GetVC() *clock.VectorClock {
	return re.vc
}

// Get the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (fo *TraceElementRoutineEnd) GetwVc() *clock.VectorClock {
	return fo.wVc
}

// Get the string representation of the object type
//
// Parameter:
//   - operation (bool): if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (re *TraceElementRoutineEnd) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeRoutineEnd + "E"
	}
	return ObjectTypeRoutineEnd
}

// Given a trace element, check if it is equal to this element
//
// Parameter:
//   - elem (TraceElement): The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (re *TraceElementRoutineEnd) IsEqual(elem TraceElement) bool {
	return re.routine == elem.GetRoutine() && re.ToString() == elem.ToString()
}

// Get the trace local index of the element in the trace
//
// Returns:
//   - VectorClock: The trace local index of the element in the trace
func (re *TraceElementRoutineEnd) GetTraceIndex() (int, int) {
	return re.routine, re.index
}

// Set the tPre and tPost of the element
//
// Parameter:
//   - time (int): The tPre and tPost of the element
func (re *TraceElementRoutineEnd) SetT(time int) {
	re.tPost = time
}

// Set the tpre of the element.
//
// Parameter:
//   - tPre (int): The tpre of the element
func (re *TraceElementRoutineEnd) SetTPre(tPre int) {
	re.tPost = tPre
}

// Set the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort (int): The timer of the element
func (re *TraceElementRoutineEnd) SetTSort(tpost int) {
	re.SetTPre(tpost)
	re.tPost = tpost
}

// Set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort (int): The timer of the element
func (re *TraceElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
	re.SetTPre(tSort)
	if re.tPost != 0 {
		re.tPost = tSort
	}
}

// Get the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (re *TraceElementRoutineEnd) ToString() string {
	return "E" + "," + strconv.Itoa(re.tPost)
}

// Update and calculate the vector clock of the element
func (re *TraceElementRoutineEnd) updateVectorClock() {
	re.vc = currentVC[re.routine].Copy()
	re.wVc = currentVC[re.routine].Copy()
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

// Dummy function for traceElement
func (re *TraceElementRoutineEnd) AddRel1(_ TraceElement, _ int) {
	return
}

// Dummy function for traceElement
func (re *TraceElementRoutineEnd) AddRel2(_ TraceElement) {
	return
}

// Dummy function for traceElement
func (re *TraceElementRoutineEnd) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

// Dummy function for traceElement
func (re *TraceElementRoutineEnd) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
