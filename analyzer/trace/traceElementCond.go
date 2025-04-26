// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementCond.go
// Brief: Struct and functions for operations of conditional variables in the trace
//
// Author: Erik Kassubek
// Created: 2023-12-25
//
// License: BSD-3-Clause

package trace

import (
	"analyzer/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// OpCond provides an enum for the operation of a conditional variable
type OpCond int

// Values for the OpCount enum
const (
	WaitCondOp OpCond = iota
	SignalOp
	BroadcastOp
)

// TraceElementCond is a trace element for a condition variable
// Fields:
//
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the condition variable
//   - opC opCond: The operation on the condition variable
//   - file string, The file of the condition variable operation in the code
//   - line int, The line of the condition variable operation in the code
//   - tID string: The id of the trace element, contains the position and the tPre
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

// AddTraceElementCond adds a new condition variable element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the condition variable
//   - opC string: The operation on the condition variable
//   - pos string: The position of the condition variable operation in the code
func (t *Trace) AddTraceElementCond(routine int, tPre string, tPost string, id string, opN string, pos string) error {
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

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementCond{
		index:   t.numberElemsInTrace[routine],
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		opC:     op,
		file:    file,
		line:    line,
		vc:      nil,
		wVc:     nil,
		rel1:    make([]TraceElement, 2),
		rel2:    make([]TraceElement, 0),
	}

	t.AddElement(&elem)
	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (co *TraceElementCond) GetID() int {
	return co.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine id
func (co *TraceElementCond) GetRoutine() int {
	return co.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (co *TraceElementCond) GetTPre() int {
	return co.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (co *TraceElementCond) GetTPost() int {
	return co.tPost
}

// GetTSort returns the timer, that is used for sorting the trace
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

// GetPos returns the position of the operation in the form [file]:[line].
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

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (co *TraceElementCond) GetTID() string {
	return co.GetPos() + "@" + strconv.Itoa(co.tPre)
}

// GetOpC returns the operation of the element
//
// Returns:
//   - OpCond: The operation of the element
func (co *TraceElementCond) GetOpC() OpCond {
	return co.opC
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (co *TraceElementCond) SetVc(vc *clock.VectorClock) {
	co.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (co *TraceElementCond) SetWVc(vc *clock.VectorClock) {
	co.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (co *TraceElementCond) GetVC() *clock.VectorClock {
	return co.vc
}

// GetWVc returns the vector clock of the element for the weak must happens before relation
//
// Returns:
//   - VectorClock: The vector clock of the element
func (co *TraceElementCond) GetWVc() *clock.VectorClock {
	return co.wVc
}

// GetObjType returns the string representation of the object type
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

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (co *TraceElementCond) IsEqual(elem TraceElement) bool {
	return co.routine == elem.GetRoutine() && co.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (co *TraceElementCond) GetTraceIndex() (int, int) {
	return co.routine, co.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (co *TraceElementCond) SetT(time int) {
	co.tPre = time
	co.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (co *TraceElementCond) SetTPre(tPre int) {
	co.tPre = tPre
	if co.tPost != 0 && co.tPost < tPre {
		co.tPost = tPre
	}
}

// SetTSort sets the timer that is used for sorting the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (co *TraceElementCond) SetTSort(tSort int) {
	co.SetTPre(tSort)
	if co.opC == WaitCondOp {
		co.tPost = tSort
	}
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
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

// ToString returns the string representation of the element
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

// AddRel1 adds an element to the rel1 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
//   - pos int: before (0) or after (1)
func (co *TraceElementCond) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}
	// do not add yourself
	if co.IsEqual(elem) {
		return
	}

	co.rel1[pos] = elem
}

// AddRel2 adds an element to the rel2 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
func (co *TraceElementCond) AddRel2(elem TraceElement) {
	// do not add yourself
	if co.IsEqual(elem) {
		return
	}

	co.rel2 = append(co.rel2, elem)
}

// GetRel1 returns the rel1 set
//
// Returns:
//   - []*TraceElement: the rel1 set
func (co *TraceElementCond) GetRel1() []TraceElement {
	return co.rel1
}

// GetRel2 returns the rel2 set
//
// Returns:
//   - []*TraceElement: the rel2 set
func (co *TraceElementCond) GetRel2() []TraceElement {
	return co.rel2
}
