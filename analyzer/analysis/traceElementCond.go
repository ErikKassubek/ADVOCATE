// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementCond.go
// Brief: Struct and functions for operations of conditional variables in the trace
//
// Author: Erik Kassubek
// Created: 2023-12-25
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// OpCond provides an enum for the operation of a conditional variable
type OpCond int

const (
	WaitCondOp OpCond = iota
	SignalOp
	BroadcastOp
)

// TraceElementCond is a trace element for a condition variable
// Fields:
//
//   - routine int: The routine id
//   - tpre int: The timestamp at the start of the event
//   - tpost int: The timestamp at the end of the event
//   - id int: The id of the condition variable
//   - opC opCond: The operation on the condition variable
//   - file string, lineint: The position of the condition variable operation in the code
//   - tID string: The id of the trace element, contains the position and the tpre
type TraceElementCond struct {
	index   int
	routine int
	tPre    int
	tPost   int
	id      int
	opC     OpCond
	file    string
	line    int
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
	rel1    []TraceElement
	rel2    []TraceElement
}

// Create a new condition variable trace element
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the condition variable
//   - opC string: The operation on the condition variable
//   - pos string: The position of the condition variable operation in the code
func AddTraceElementCond(routine int, tPre string, tPost string, id string, opN string, pos string) error {
	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tpre is not an integer")
	}
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}
	var op OpCond
	switch opN {
	case "W":
		op = WaitCondOp
	case "S":
		op = SignalOp
	case "B":
		op = BroadcastOp
	default:
		return errors.New("op is not a valid operation")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementCond{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		opC:     op,
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

// Get the id of the element
//
// Returns:
//   - int: The id of the element
func (co *TraceElementCond) GetID() int {
	return co.id
}

// Get the routine of the element
//
// Returns:
//   - int: The routine id
func (co *TraceElementCond) GetRoutine() int {
	return co.routine
}

// Get the tpre of the element.
//
// Returns:
//   - int: The tpre of the element
func (co *TraceElementCond) GetTPre() int {
	return co.tPre
}

// Get the tpost of the element.
//
// Returns:
//   - int: The tpost of the element
func (co *TraceElementCond) GetTPost() int {
	return co.tPost
}

// Get the timer, that is used for sorting the trace
//
// Returns:
//   - int: The timer of the element
func (co *TraceElementCond) GetTSort() int {
	t := co.tPre
	if co.opC == WaitCondOp {
		t = co.tPost
	}
	if t == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return t
}

// Get the position of the operation.
//
// Returns:
//   - string: The position of the element
func (co *TraceElementCond) GetPos() string {
	return fmt.Sprintf("%s:%d", co.file, co.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (co *TraceElementCond) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", co.routine, co.file, co.line)
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (co *TraceElementCond) GetFile() string {
	return co.file
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (co *TraceElementCond) GetLine() int {
	return co.line
}

// Get the tID of the element.
//
// Returns:
//   - string: The tID of the element
func (co *TraceElementCond) GetTID() string {
	return co.GetPos() + "@" + strconv.Itoa(co.tPre)
}

// Get the operation of the element
//
// Returns:
//   - OpCond: The operation of the element
func (co *TraceElementCond) GetOpCond() OpCond {
	return co.opC
}

// Get the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (co *TraceElementCond) GetVC() *clock.VectorClock {
	return co.vc
}

// Get the vector clock of the element for the weak must happens before relation
//
// Returns:
//   - VectorClock: The vector clock of the element
func (co *TraceElementCond) GetwVc() *clock.VectorClock {
	return co.wVc
}

// Get the string representation of the object type
func (co *TraceElementCond) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeCond
	}

	switch co.opC {
	case WaitCondOp:
		return ObjectTypeCond + "W"
	case BroadcastOp:
		return ObjectTypeCond + "B"
	case SignalOp:
		return ObjectTypeCond + "S"
	}
	return ObjectTypeCond
}

// Given a trace element, check if it is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (co *TraceElementCond) IsEqual(elem TraceElement) bool {
	return co.routine == elem.GetRoutine() && co.ToString() == elem.ToString()
}

// Get the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (co *TraceElementCond) GetTraceIndex() (int, int) {
	return co.routine, co.index
}

// Set the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (co *TraceElementCond) SetT(time int) {
	co.tPre = time
	co.tPost = time
}

// Set the tpre of the element.
//
// Parameter:
//   - tPre int: The tpre of the element
func (co *TraceElementCond) SetTPre(tPre int) {
	co.tPre = tPre
	if co.tPost != 0 && co.tPost < tPre {
		co.tPost = tPre
	}
}

// Set the timer that is used for sorting the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (co *TraceElementCond) SetTSort(tSort int) {
	co.SetTPre(tSort)
	if co.opC == WaitCondOp {
		co.tPost = tSort
	}
}

// Set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (co *TraceElementCond) SetTWithoutNotExecuted(tSort int) {
	co.SetTPre(tSort)
	if co.opC == WaitCondOp {
		if co.tPost != 0 {
			co.tPost = tSort
		}
		return
	}
	if co.tPre != 0 {
		co.tPre = tSort
	}
	return
}

// Get the string representation of the element
//
// Returns:
//   - string: The string representation of the element
func (co *TraceElementCond) ToString() string {
	res := "D,"
	res += strconv.Itoa(co.tPre) + "," + strconv.Itoa(co.tPost) + ","
	res += strconv.Itoa(co.id) + ","
	switch co.opC {
	case WaitCondOp:
		res += "W"
	case SignalOp:
		res += "S"
	case BroadcastOp:
		res += "B"
	}
	res += "," + co.GetPos()
	return res
}

// Update the vector clock of the trace and element
func (co *TraceElementCond) updateVectorClock() {
	co.vc = currentVC[co.routine].Copy()
	co.wVc = currentWVC[co.routine].Copy()

	switch co.opC {
	case WaitCondOp:
		CondWait(co)
	case SignalOp:
		CondSignal(co)
	case BroadcastOp:
		CondBroadcast(co)
	}

}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (co *TraceElementCond) Copy() TraceElement {
	return &TraceElementCond{
		index:   co.index,
		routine: co.routine,
		tPre:    co.tPre,
		tPost:   co.tPost,
		id:      co.id,
		opC:     co.opC,
		file:    co.file,
		line:    co.line,
		vc:      co.vc.Copy(),
		wVc:     co.wVc.Copy(),
		rel1:    co.rel1,
		rel2:    co.rel1,
	}
}

// ========= For GoPie fuzzing ===========

// Add an element to the rel1 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
//   - pos int: before (0) or after (1)
func (co *TraceElementCond) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}
	co.rel1[pos] = elem
}

// Add an element to the rel2 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
func (co *TraceElementCond) AddRel2(elem TraceElement) {
	co.rel2 = append(co.rel2, elem)
}

// Return the rel1 set
//
// Returns:
//   - []*TraceElement: the rel1 set
func (co *TraceElementCond) GetRel1() []TraceElement {
	return co.rel1
}

// Return the rel2 set
//
// Returns:
//   - []*TraceElement: the rel2 set
func (co *TraceElementCond) GetRel2() []TraceElement {
	return co.rel2
}
