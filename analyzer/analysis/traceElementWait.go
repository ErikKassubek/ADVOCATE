// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementWait.go
// Brief: Struct and functions for wait group operations in the trace
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
	"math"
	"strconv"
)

// enum for opW
type opW int

// Values for the opW enum
const (
	ChangeOp opW = iota
	WaitOp
)

// TraceElementWait is a trace element for a wait group statement
//
// Fields:
//   - index int: Index in the routine
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the wait group
//   - opW opW: The operation on the wait group
//   - delta int: The delta of the wait group
//   - val int: The value of the wait group
//   - file string: The file of the wait group in the code
//   - line int: The line of the wait group in the code
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//   - the rel1 set for GoPie fuzzing
//   - the rel2 set for GoPie fuzzing
type TraceElementWait struct {
	index   int
	routine int
	tPre    int
	tPost   int
	id      int
	opW     opW
	delta   int
	val     int
	file    string
	line    int
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
	rel1    []TraceElement
	rel2    []TraceElement
}

// AddTraceElementWait adds a new wait group element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the wait group
//   - opW string: The operation on the wait group
//   - delta string: The delta of the wait group
//   - val string: The value of the wait group
//   - pos string: The position of the wait group in the code
func AddTraceElementWait(routine int, tPre,
	tPost, id, opW, delta, val, pos string) error {
	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	opWOp := ChangeOp
	if opW == "W" {
		opWOp = WaitOp
	} else if opW != "A" {
		return errors.New("op is not a valid operation")
	}

	deltaInt, err := strconv.Atoi(delta)
	if err != nil {
		return errors.New("delta is not an integer")
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("val is not an integer")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementWait{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		opW:     opWOp,
		delta:   deltaInt,
		val:     valInt,
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
func (wa *TraceElementWait) GetID() int {
	return wa.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (wa *TraceElementWait) GetRoutine() int {
	return wa.routine
}

// GetTPre returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the start of the event
func (wa *TraceElementWait) GetTPre() int {
	return wa.tPre
}

// GetTPost returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the end of the event
func (wa *TraceElementWait) GetTPost() int {
	return wa.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (wa *TraceElementWait) GetTSort() int {
	if wa.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return wa.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (wa *TraceElementWait) GetPos() string {
	return fmt.Sprintf("%s:%d", wa.file, wa.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (wa *TraceElementWait) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", wa.routine, wa.file, wa.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (wa *TraceElementWait) GetFile() string {
	return wa.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (wa *TraceElementWait) GetLine() int {
	return wa.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (wa *TraceElementWait) GetTID() string {
	return wa.GetPos() + "@" + strconv.Itoa(wa.tPre)
}

// IsWait returns if the operation is a wait op
//
// Returns:
//   - bool: True if the operation is a wait op
func (wa *TraceElementWait) IsWait() bool {
	return wa.opW == WaitOp
}

// GetDelta returns the delta of the element. The delta is the value by which the counter
// of the wait has been changed. For Add the delta is > 0, for Done it is -1,
// for Wait it is 0
//
// Returns:
//   - int: the delta of the wait element
func (wa *TraceElementWait) GetDelta() int {
	return wa.delta
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (wa *TraceElementWait) GetVC() *clock.VectorClock {
	return wa.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (wa *TraceElementWait) GetWVc() *clock.VectorClock {
	return wa.wVc
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (wa *TraceElementWait) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeWait
	}

	if wa.delta > 0 {
		return ObjectTypeWait + "A"
	} else if wa.delta < 0 {
		return ObjectTypeWait + "D"
	}
	return ObjectTypeWait + "W"
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (wa *TraceElementWait) IsEqual(elem TraceElement) bool {
	return wa.routine == elem.GetRoutine() && wa.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (wa *TraceElementWait) GetTraceIndex() (int, int) {
	return wa.routine, wa.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (wa *TraceElementWait) SetT(time int) {
	wa.tPre = time
	wa.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (wa *TraceElementWait) SetTPre(tPre int) {
	wa.tPre = tPre
	if wa.tPost != 0 && wa.tPost < tPre {
		wa.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (wa *TraceElementWait) SetTSort(tSort int) {
	wa.SetTPre(tSort)
	wa.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (wa *TraceElementWait) SetTWithoutNotExecuted(tSort int) {
	wa.SetTPre(tSort)
	if wa.tPost != 0 {
		wa.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (wa *TraceElementWait) ToString() string {
	res := "W,"
	res += strconv.Itoa(wa.tPre) + "," + strconv.Itoa(wa.tPost) + ","
	res += strconv.Itoa(wa.id) + ","
	switch wa.opW {
	case ChangeOp:
		res += "A,"
	case WaitOp:
		res += "W,"
	}

	res += strconv.Itoa(wa.delta) + "," + strconv.Itoa(wa.val)
	res += "," + wa.GetPos()
	return res
}

// Update and calculate the vector clock of the element
func (wa *TraceElementWait) updateVectorClock() {
	wa.vc = currentVC[wa.routine].Copy()
	wa.wVc = currentWVC[wa.routine].Copy()

	switch wa.opW {
	case ChangeOp:
		Change(wa)
	case WaitOp:
		Wait(wa)
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		utils.LogError(err)
	}
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (wa *TraceElementWait) Copy() TraceElement {
	return &TraceElementWait{
		index:   wa.index,
		routine: wa.routine,
		tPre:    wa.tPre,
		tPost:   wa.tPost,
		id:      wa.id,
		opW:     wa.opW,
		delta:   wa.delta,
		val:     wa.val,
		file:    wa.file,
		line:    wa.line,
		vc:      wa.vc.Copy(),
		wVc:     wa.wVc.Copy(),
		rel1:    wa.rel1,
		rel2:    wa.rel1,
	}
}

// ========= For GoPie fuzzing ===========

// AddRel1 adds an element to the rel1 set of the element
//
// Parameter:
//
//	elem TraceElement: elem to add
//	pos int: before (0) or after (1)
func (wa *TraceElementWait) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}
	wa.rel1[pos] = elem
}

// AddRel2 adds an element to the rel2 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
func (wa *TraceElementWait) AddRel2(elem TraceElement) {
	wa.rel2 = append(wa.rel2, elem)
}

// GetRel1 returns the rel1 set
//
// Returns:
//   - []*TraceElement: the rel1 set
func (wa *TraceElementWait) GetRel1() []TraceElement {
	return wa.rel1
}

// GetRel2 returns the rel2 set
//
// Returns:
//   - []*TraceElement: the rel2 set
func (wa *TraceElementWait) GetRel2() []TraceElement {
	return wa.rel1
}
