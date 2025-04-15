// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementAtomic.go
// Brief: Struct and functions for atomic operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/utils"
	"errors"
	"fmt"
	"strconv"
)

// enum for operation
type opAtomic int

// Values for the opAtomic enum
const (
	LoadOp opAtomic = iota
	StoreOp
	AddOp
	SwapOp
	CompSwapOp
	AndOp
	OrOp
)

// TraceElementAtomic is a struct to save an atomic event in the trace
// Fields:
//
//   - index int: index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp of the event
//   - id int: The id of the atomic variable
//   - opA opAtomic: The operation on the atomic variable
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//   - file string: the file of the operation
//   - line int: the line of the operation
//   - the rel1 set for GoPie fuzzing
//   - the rel2 set for GoPie fuzzing
type TraceElementAtomic struct {
	index   int
	routine int
	tPost   int
	id      int
	opA     opAtomic
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
	file    string
	line    int
	rel1    []TraceElement
	rel2    []TraceElement
}

// AddTraceElementAtomic adds a new atomic trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp of the event
//   - id string: The id of the atomic variable
//   - operation string: The operation on the atomic variable
//   - pos string: The position of the atomic
func AddTraceElementAtomic(routine int, tPost string,
	id string, operation string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	var opAInt opAtomic
	switch operation {
	case "L":
		opAInt = LoadOp
	case "S":
		opAInt = StoreOp
	case "A":
		opAInt = AddOp
	case "W":
		opAInt = SwapOp
	case "C":
		opAInt = CompSwapOp
	case "N":
		opAInt = AndOp
	case "O":
		opAInt = OrOp
	default:
		return fmt.Errorf("Atomic operation '%s' is not a valid operation", operation)
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementAtomic{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPost:   tPostInt,
		id:      idInt,
		opA:     opAInt,
		file:    file,
		line:    line,
		vc:      clock.NewVectorClock(MainTrace.numberOfRoutines),
		wVc:     clock.NewVectorClock(MainTrace.numberOfRoutines),
		rel1:    make([]TraceElement, 2),
		rel2:    make([]TraceElement, 0),
	}

	AddElementToTrace(&elem)
	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (at *TraceElementAtomic) GetID() int {
	return at.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (at *TraceElementAtomic) GetRoutine() int {
	return at.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (at *TraceElementAtomic) GetTPre() int {
	return at.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (at *TraceElementAtomic) GetTPost() int {
	return at.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (at *TraceElementAtomic) GetTSort() int {
	return at.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (at *TraceElementAtomic) GetPos() string {
	return fmt.Sprintf("%s:%d", at.file, at.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (at *TraceElementAtomic) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", at.routine, at.file, at.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (at *TraceElementAtomic) GetFile() string {
	return at.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (at *TraceElementAtomic) GetLine() int {
	return at.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (at *TraceElementAtomic) GetTID() string {
	return at.GetPos() + "@" + strconv.Itoa(at.tPost)
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (at *TraceElementAtomic) GetVC() *clock.VectorClock {
	return at.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The weak vector clock of the element
func (at *TraceElementAtomic) GetWVc() *clock.VectorClock {
	return at.wVc
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (at *TraceElementAtomic) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeAtomic
	}

	switch at.opA {
	case LoadOp:
		return ObjectTypeAtomic + "L"
	case StoreOp:
		return ObjectTypeAtomic + "S"
	case AddOp:
		return ObjectTypeAtomic + "A"
	case SwapOp:
		return ObjectTypeAtomic + "W"
	case CompSwapOp:
		return ObjectTypeAtomic + "C"
	}

	return ObjectTypeAtomic
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (at *TraceElementAtomic) IsEqual(elem TraceElement) bool {
	return at.routine == elem.GetRoutine() && at.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (at *TraceElementAtomic) GetTraceIndex() (int, int) {
	return at.routine, at.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (at *TraceElementAtomic) SetT(time int) {
	at.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPost of the element
func (at *TraceElementAtomic) SetTPre(tPre int) {
	at.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (at *TraceElementAtomic) SetTSort(tSort int) {
	at.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (at *TraceElementAtomic) SetTWithoutNotExecuted(tSort int) {
	if at.tPost != 0 {
		at.tPost = tSort
	}
}

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (at *TraceElementAtomic) ToString() string {
	opString := ""

	switch at.opA {
	case LoadOp:
		opString = "L"
	case StoreOp:
		opString = "S"
	case AddOp:
		opString = "A"
	case SwapOp:
		opString = "W"
	case CompSwapOp:
		opString = "C"
	default:
		opString = "U"
	}

	return fmt.Sprintf("A,%d,%d,%s,%s", at.tPost, at.id, opString, at.GetPos())
}

// Store and update the vector clock of the element
func (at *TraceElementAtomic) updateVectorClock() {
	at.vc = currentVC[at.routine].Copy()
	at.wVc = currentWVC[at.routine].Copy()

	switch at.opA {
	case LoadOp:
		Read(at, true)
	case StoreOp, AddOp, AndOp, OrOp:
		Write(at)
	case SwapOp, CompSwapOp:
		Swap(at, true)
	default:
		err := "Unknown operation: " + at.ToString()
		utils.LogError(err)
	}
}

// Store and update the vector clock of the element if the IgnoreCriticalSections
// tag has been set
func (at *TraceElementAtomic) updateVectorClockAlt() {
	at.vc = currentVC[at.routine].Copy()

	switch at.opA {
	case LoadOp:
		Read(at, false)
	case StoreOp, AddOp, AndOp, OrOp:
		Write(at)
	case SwapOp, CompSwapOp:
		Swap(at, false)
	default:
		err := "Unknown operation: " + at.ToString()
		utils.LogError(err)
	}
}

// Copy the atomic element
//
// Returns:
//   - TraceElement: The copy of the element
func (at *TraceElementAtomic) Copy() TraceElement {
	return &TraceElementAtomic{
		index:   at.index,
		routine: at.routine,
		tPost:   at.tPost,
		id:      at.id,
		opA:     at.opA,
		vc:      at.vc.Copy(),
		wVc:     at.wVc.Copy(),
		rel1:    at.rel1,
		rel2:    at.rel1,
	}
}

// ========= For GoPie fuzzing ===========

// AddRel1 adds an element to the rel1 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
//   - pos int: before (0) or after (1)
func (at *TraceElementAtomic) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}

	// do not add yourself
	if at.IsEqual(elem) {
		return
	}

	at.rel1[pos] = elem
}

// AddRel2 adds an element to the rel2 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
func (at *TraceElementAtomic) AddRel2(elem TraceElement) {
	// do not add yourself
	if at.IsEqual(elem) {
		return
	}

	at.rel2 = append(at.rel2, elem)
}

// GetRel1 returns the rel1 set
//
// Returns:
//   - []TraceElement: the rel1 set
func (at *TraceElementAtomic) GetRel1() []TraceElement {
	return at.rel1
}

// GetRel2 returns the rel2 set
//
// Returns:
//   - []TraceElement: the rel2 set
func (at *TraceElementAtomic) GetRel2() []TraceElement {
	return at.rel2
}
