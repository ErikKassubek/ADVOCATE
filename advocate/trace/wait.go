// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementWait.go
// Brief: Struct and functions for wait group operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// enum for OpWait
type OpWait int

// Values for the opW enum
const (
	ChangeOp OpWait = iota
	WaitOp
)

// ElementWait is a trace element for a wait group statement
//
// Fields:
//   - traceID: id of the element, should never be changed
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
//   - children []TraceElement: children in partial order graph
//   - parent []TraceElement: parents in partial order graph
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
type ElementWait struct {
	traceID          int
	index            int
	routine          int
	tPre             int
	tPost            int
	ID               int
	opW              OpWait
	delta            int
	val              int
	file             string
	line             int
	vc               *clock.VectorClock
	wVc              *clock.VectorClock
	children         []Element
	parents          []Element
	numberConcurrent int
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
func (t *Trace) AddTraceElementWait(routine int, tPre,
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

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementWait{
		index:            t.numberElemsInTrace[routine],
		routine:          routine,
		tPre:             tPreInt,
		tPost:            tPostInt,
		ID:               idInt,
		opW:              opWOp,
		delta:            deltaInt,
		val:              valInt,
		file:             file,
		line:             line,
		vc:               nil,
		wVc:              nil,
		children:         make([]Element, 0),
		parents:          make([]Element, 0),
		numberConcurrent: -1,
	}

	t.AddElement(&elem)

	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (wa *ElementWait) GetID() int {
	return wa.ID
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (wa *ElementWait) GetRoutine() int {
	return wa.routine
}

// GetTPre returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the start of the event
func (wa *ElementWait) GetTPre() int {
	return wa.tPre
}

// GetTPost returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the end of the event
func (wa *ElementWait) GetTPost() int {
	return wa.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (wa *ElementWait) GetTSort() int {
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
func (wa *ElementWait) GetPos() string {
	return fmt.Sprintf("%s:%d", wa.file, wa.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (wa *ElementWait) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", wa.routine, wa.file, wa.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (wa *ElementWait) GetFile() string {
	return wa.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (wa *ElementWait) GetLine() int {
	return wa.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (wa *ElementWait) GetTID() string {
	return wa.GetPos() + "@" + strconv.Itoa(wa.tPre)
}

// IsWait returns if the operation is a wait op
//
// Returns:
//   - bool: True if the operation is a wait op
func (wa *ElementWait) IsWait() bool {
	return wa.opW == WaitOp
}

// GetOpW returns the operation type
//
// Returns:
//   - opWait: the wait operations
func (wa *ElementWait) GetOpW() OpWait {
	return wa.opW
}

// GetDelta returns the delta of the element. The delta is the value by which the counter
// of the wait has been changed. For Add the delta is > 0, for Done it is -1,
// for Wait it is 0
//
// Returns:
//   - int: the delta of the wait element
func (wa *ElementWait) GetDelta() int {
	return wa.delta
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (wa *ElementWait) SetVc(vc *clock.VectorClock) {
	wa.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (wa *ElementWait) SetWVc(vc *clock.VectorClock) {
	wa.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (wa *ElementWait) GetVC() *clock.VectorClock {
	return wa.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (wa *ElementWait) GetWVc() *clock.VectorClock {
	return wa.wVc
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (wa *ElementWait) GetObjType(operation bool) string {
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
func (wa *ElementWait) IsEqual(elem Element) bool {
	return wa.routine == elem.GetRoutine() && wa.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (wa *ElementWait) GetTraceIndex() (int, int) {
	return wa.routine, wa.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (wa *ElementWait) SetT(time int) {
	wa.tPre = time
	wa.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (wa *ElementWait) SetTPre(tPre int) {
	wa.tPre = tPre
	if wa.tPost != 0 && wa.tPost < tPre {
		wa.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (wa *ElementWait) SetTSort(tSort int) {
	wa.SetTPre(tSort)
	wa.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (wa *ElementWait) SetTWithoutNotExecuted(tSort int) {
	wa.SetTPre(tSort)
	if wa.tPost != 0 {
		wa.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (wa *ElementWait) ToString() string {
	res := "W,"
	res += strconv.Itoa(wa.tPre) + "," + strconv.Itoa(wa.tPost) + ","
	res += strconv.Itoa(wa.ID) + ","
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

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (wa *ElementWait) GetTraceID() int {
	return wa.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (wa *ElementWait) setTraceID(ID int) {
	wa.traceID = ID
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (wa *ElementWait) Copy() Element {
	children := make([]Element, len(wa.children))
	copy(children, wa.children)
	parents := make([]Element, len(wa.parents))
	copy(parents, wa.parents)

	return &ElementWait{
		traceID:          wa.traceID,
		index:            wa.index,
		routine:          wa.routine,
		tPre:             wa.tPre,
		tPost:            wa.tPost,
		ID:               wa.ID,
		opW:              wa.opW,
		delta:            wa.delta,
		val:              wa.val,
		file:             wa.file,
		line:             wa.line,
		vc:               wa.vc.Copy(),
		wVc:              wa.wVc.Copy(),
		children:         children,
		parents:          parents,
		numberConcurrent: wa.numberConcurrent,
	}
}

// AddChild adds an element as a child of this node in the partial order graph
//
// Parameter:
//   - elem *TraceElement: the element to add
func (wa *ElementWait) AddChild(elem Element) {
	wa.children = append(wa.children, elem)
}

// GetChildren returns all children of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the children
func (wa *ElementWait) GetChildren() []Element {
	return wa.children
}

// GetParents returns all parents of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the parents
func (wa *ElementWait) GetParents() []Element {
	return wa.children
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - number of concurrent element, or -1
func (wa *ElementWait) GetNumberConcurrent() int {
	return wa.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
func (wa *ElementWait) SetNumberConcurrent(c int) {
	wa.numberConcurrent = c
}
