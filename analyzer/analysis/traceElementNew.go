// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementNew.go
// Brief: Trace element to store the creation (new) of relevant operations. For now this is only creates the new for channel. This may be expanded later.
//
// Author: Erik Kassubek
// Created: 2024-11-29
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"errors"
	"fmt"
	"strconv"
)

type newOpType string

const (
	atomicVar   newOpType = "A"
	channel     newOpType = "C"
	conditional newOpType = "D"
	mutex       newOpType = "M"
	once        newOpType = "O"
	wait        newOpType = "W"
)

// TraceElementNew is a trace element for the creation of an object / new
// Fields:
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp of the new
//   - id int: The id of the underlying operation
//   - elemType newOpType: The type of the created object
//   - num int: Variable field for additional information
//   - file string: The file of the new
//   - line int: The line of the new
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//
// For now this is only creates the new for channel. This may be expanded later.
type TraceElementNew struct {
	index    int
	routine  int
	tPost    int
	id       int
	elemType newOpType
	num      int
	file     string
	line     int
	vc       *clock.VectorClock
	wVc      *clock.VectorClock
}

// Create a new trace element
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the channel
//   - elemType string: Type of the created primitive
//   - num string: Variable field for additional information
//   - pos string: position
func AddTraceElementNew(routine int, tPost string, id string, elemType string, num string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	numInt, err := strconv.Atoi(num)
	if err != nil {
		return errors.New("num is not an integer")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementNew{
		index:    numberElemsInTrace(routine),
		routine:  routine,
		tPost:    tPostInt,
		id:       idInt,
		elemType: newOpType(elemType),
		num:      numInt,
		file:     file,
		line:     line,
		vc:       clock.NewVectorClock(MainTrace.numberOfRoutines),
		wVc:      clock.NewVectorClock(MainTrace.numberOfRoutines),
	}

	AddElementToTrace(&elem)
	return nil
}

// Get the id of the element
//
// Returns:
//   - int: The id of the element
func (n *TraceElementNew) GetID() int {
	return n.id
}

// Get the tpre of the element
//
// Returns:
//   - int: The tpre of the element
func (n *TraceElementNew) GetTPre() int {
	return n.tPost
}

// Get the position of the operation.
//
// Returns:
//   - string: The position of the element
func (n *TraceElementNew) GetTPost() int {
	return n.tPost
}

// Get the timer, that is used for the sorting of the trace
//
// Returns:
//   - float32: The time of the element
func (n *TraceElementNew) GetTSort() int {
	return n.tPost
}

// Get the routine of the element
//
// Returns:
//   - int: The routine of the element
func (n *TraceElementNew) GetRoutine() int {
	return n.routine
}

// Get the position of the operation.
//
// Returns:
//   - string: The position of the element
func (n *TraceElementNew) GetPos() string {
	return fmt.Sprintf("%s:%d", n.file, n.line)
}

// Get the replayId of the element
//
// Returns:
//   - int: The replayId of the element
func (n *TraceElementNew) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", n.routine, n.file, n.line)
}

// Get the file of the element
//
// Returns:
//   - int: The file of the element
func (n *TraceElementNew) GetFile() string {
	return n.file
}

// Get the line of the element
//
// Returns:
//   - int: The line of the element
func (n *TraceElementNew) GetLine() int {
	return n.line
}

// Get the tID of the element
//
// Returns:
//   - int: The tID of the element
func (n *TraceElementNew) GetTID() string {
	return n.GetPos() + "@" + strconv.Itoa(n.tPost)
}

// Get the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (n *TraceElementNew) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeNew
	}

	switch n.elemType {
	case atomicVar:
		return ObjectTypeNew + "A"
	case channel:
		return ObjectTypeNew + "C"
	case conditional:
		return ObjectTypeNew + "D"
	case mutex:
		return ObjectTypeNew + "M"
	case once:
		return ObjectTypeNew + "O"
	case wait:
		return ObjectTypeNew + "W"
	default:
		return ObjectTypeNew
	}
}

// Get the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (n *TraceElementNew) GetVC() *clock.VectorClock {
	return n.vc
}

// Get the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (n *TraceElementNew) GetwVc() *clock.VectorClock {
	return n.wVc
}

// Get the num field of the element
//
// Returns:
//   - VectorClock: The num field of the element
func (n *TraceElementNew) GetNum() int {
	return n.num
}

// Get the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (n *TraceElementNew) GetTraceIndex() (int, int) {
	return n.routine, n.index
}

// Get the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (n *TraceElementNew) ToString() string {
	return fmt.Sprintf("N,%d,%d,%s,%d,%s", n.tPost, n.id, string(n.elemType), n.num, n.GetPos())
}

// Given a trace element, check if it is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (n *TraceElementNew) IsEqual(elem TraceElement) bool {
	return n.routine == elem.GetRoutine() && n.ToString() == elem.ToString()
}

// Set the tpre of the element.
//
// Parameter:
//   - tPre int: The tpre of the element
func (n *TraceElementNew) SetTPre(tSort int) {
	n.tPost = tSort
}

// Set the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (n *TraceElementNew) SetT(tSort int) {
	n.tPost = tSort
}

// Set the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (n *TraceElementNew) SetTSort(tSort int) {
	n.tPost = tSort
}

// Set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (n *TraceElementNew) SetTWithoutNotExecuted(tSort int) {
	if n.tPost == 0 {
		return
	}
	n.tPost = tSort
}

// Store and update the vector clock of the element
func (n *TraceElementNew) updateVectorClock() {
	n.vc = currentVC[n.routine].Copy()
	n.wVc = currentWVC[n.routine].Copy()

	currentVC[n.routine].Inc(n.routine)
	currentWVC[n.routine].Inc(n.routine)
}

func (n *TraceElementNew) Copy() TraceElement {
	return &TraceElementNew{
		index:    n.index,
		routine:  n.routine,
		tPost:    n.tPost,
		id:       n.id,
		elemType: n.elemType,
		file:     n.file,
		line:     n.line,
		vc:       n.vc.Copy(),
		wVc:      n.wVc.Copy(),
	}
}

// Dummy function for traceElement
func (n *TraceElementNew) AddRel1(_ TraceElement, _ int) {
	return
}

// Dummy function for traceElement
func (n *TraceElementNew) AddRel2(_ TraceElement) {
	return
}

// Dummy function for traceElement
func (n *TraceElementNew) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

// Dummy function for traceElement
func (n *TraceElementNew) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
