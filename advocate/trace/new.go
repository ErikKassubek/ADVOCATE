// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementNew.go
// Brief: Trace element to store the creation (new) of relevant operations. For now this is only creates the new for channel. This may be expanded later.
//
// Author: Erik Kassubek
// Created: 2024-11-29
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/clock"
	"errors"
	"fmt"
	"strconv"
)

// newOpType is an enum for type of primitive that is created
// For now only mutex is used
type newOpType string

// Values for the newOpType enum enum
const (
	AtomicVar   newOpType = "A"
	Channel     newOpType = "C"
	Conditional newOpType = "D"
	Mutex       newOpType = "M"
	Once        newOpType = "O"
	Wait        newOpType = "W"
	None        newOpType = ""
)

// ElementNew is a trace element for the creation of an object / new
// Fields:
//   - traceID: id of the element, should never be changed
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
//   - children []TraceElement: children in partial order graph
//   - parents []TraceElement: parents in partial order graph
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//
// For now this is only creates the new for channel. This may be expanded later.
type ElementNew struct {
	traceID          int
	index            int
	routine          int
	tPost            int
	id               int
	elemType         newOpType
	num              int
	file             string
	line             int
	vc               *clock.VectorClock
	wVc              *clock.VectorClock
	children         []Element
	parents          []Element
	numberConcurrent int
}

// AddTraceElementNew adds a make trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the channel
//   - elemType string: Type of the created primitive
//   - num string: Variable field for additional information
//   - pos string: position
func (t *Trace) AddTraceElementNew(routine int, tPost string, id string, elemType string, num string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	numInt, err := strconv.Atoi(num)
	if err != nil {
		return errors.New("num is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	et := None
	switch elemType {
	case "NA":
		et = AtomicVar
	case "NC":
		et = Channel
	case "ND":
		et = Conditional
	case "NM":
		et = Mutex
	case "NO":
		et = Once
	case "NW":
		et = Wait
	}

	elem := ElementNew{
		index:            t.numberElemsInTrace[routine],
		routine:          routine,
		tPost:            tPostInt,
		id:               idInt,
		elemType:         et,
		num:              numInt,
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
func (n *ElementNew) GetID() int {
	return n.id
}

// GetTPre returns the tPre of the element
//
// Returns:
//   - int: The tPre of the element
func (n *ElementNew) GetTPre() int {
	return n.tPost
}

// GetTPost returns the tPost of the operation.
//
// Returns:
//   - string: The position of the element
func (n *ElementNew) GetTPost() int {
	return n.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - float32: The time of the element
func (n *ElementNew) GetTSort() int {
	return n.tPost
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (n *ElementNew) GetRoutine() int {
	return n.routine
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (n *ElementNew) GetPos() string {
	return fmt.Sprintf("%s:%d", n.file, n.line)
}

// GetReplayID returns the replay ID of the element
//
// Returns:
//   - int: The replayId of the element
func (n *ElementNew) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", n.routine, n.file, n.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - int: The file of the element
func (n *ElementNew) GetFile() string {
	return n.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - int: The line of the element
func (n *ElementNew) GetLine() int {
	return n.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - int: The tID of the element
func (n *ElementNew) GetTID() string {
	return n.GetPos() + "@" + strconv.Itoa(n.tPost)
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (n *ElementNew) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeNew
	}

	switch n.elemType {
	case AtomicVar:
		return ObjectTypeNew + "A"
	case Channel:
		return ObjectTypeNew + "C"
	case Conditional:
		return ObjectTypeNew + "D"
	case Mutex:
		return ObjectTypeNew + "M"
	case Once:
		return ObjectTypeNew + "O"
	case Wait:
		return ObjectTypeNew + "W"
	default:
		return ObjectTypeNew
	}
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (n *ElementNew) SetVc(vc *clock.VectorClock) {
	n.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (n *ElementNew) SetWVc(vc *clock.VectorClock) {
	n.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (n *ElementNew) GetVC() *clock.VectorClock {
	return n.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (n *ElementNew) GetWVc() *clock.VectorClock {
	return n.wVc
}

// GetNum returns the num field of the element
//
// Returns:
//   - VectorClock: The num field of the element
func (n *ElementNew) GetNum() int {
	return n.num
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (n *ElementNew) GetTraceIndex() (int, int) {
	return n.routine, n.index
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (n *ElementNew) ToString() string {
	return fmt.Sprintf("N,%d,%d,%s,%d,%s", n.tPost, n.id, string(n.elemType), n.num, n.GetPos())
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (n *ElementNew) IsEqual(elem Element) bool {
	return n.routine == elem.GetRoutine() && n.ToString() == elem.ToString()
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (n *ElementNew) SetTPre(tSort int) {
	n.tPost = tSort
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (n *ElementNew) SetT(tSort int) {
	n.tPost = tSort
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (n *ElementNew) SetTSort(tSort int) {
	n.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (n *ElementNew) SetTWithoutNotExecuted(tSort int) {
	if n.tPost == 0 {
		return
	}
	n.tPost = tSort
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (n *ElementNew) GetTraceID() int {
	return n.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (n *ElementNew) setTraceID(ID int) {
	n.traceID = ID
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (n *ElementNew) Copy() Element {
	children := make([]Element, len(n.children))
	copy(children, n.children)
	parents := make([]Element, len(n.parents))
	copy(parents, n.parents)

	return &ElementNew{
		traceID:          n.traceID,
		index:            n.index,
		routine:          n.routine,
		tPost:            n.tPost,
		id:               n.id,
		elemType:         n.elemType,
		file:             n.file,
		line:             n.line,
		vc:               n.vc.Copy(),
		wVc:              n.wVc.Copy(),
		children:         children,
		parents:          parents,
		numberConcurrent: n.numberConcurrent,
	}
}

// AddChild adds an element as a child of this node in the partial order graph
//
// Parameter:
//   - elem *TraceElement: the element to add
func (n *ElementNew) AddChild(elem Element) {
	n.children = append(n.children, elem)
}

// GetChildren returns all children of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the children
func (n *ElementNew) GetChildren() []Element {
	return n.children
}

// GetParents returns all parents of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the parents
func (n *ElementNew) GetParents() []Element {
	return n.children
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - number of concurrent element, or -1
func (n *ElementNew) GetNumberConcurrent() int {
	return n.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
func (n *ElementNew) SetNumberConcurrent(c int) {
	n.numberConcurrent = c
}
