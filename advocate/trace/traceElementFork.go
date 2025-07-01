// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementFork.go
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
	"fmt"
	"strconv"
)

// TraceElementFork is a trace element for a go statement
// Fields:
//   - traceID: id of the element, should never be changed
//   - index int: ID in the new routine
//   - routine int: The routine id of
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the new go statement
//   - file (string), line int: The position of the trace element in the file
//   - vc *clock.VectorClock: the vector clock of the element
//   - wVc *clock.VectorClock: the weak vector clock of the element
//   - children []TraceElement: children in partial order graph
//   - parent []TraceElement: parents in partial order graph
type TraceElementFork struct {
	traceID  int
	index    int
	routine  int
	tPost    int
	id       int
	file     string
	line     int
	vc       *clock.VectorClock
	wVc      *clock.VectorClock
	children []TraceElement
	parents  []TraceElement
}

// AddTraceElementFork adds a new go statement element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the new routine
//   - pos string: The position of the trace element in the file
func (t *Trace) AddTraceElementFork(routine int, tPost string, id string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementFork{
		index:    t.numberElemsInTrace[routine],
		routine:  routine,
		tPost:    tPostInt,
		id:       idInt,
		file:     file,
		line:     line,
		vc:       nil,
		wVc:      nil,
		children: make([]TraceElement, 0),
		parents:  make([]TraceElement, 0),
	}

	t.AddElement(&elem)
	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (fo *TraceElementFork) GetID() int {
	return fo.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (fo *TraceElementFork) GetRoutine() int {
	return fo.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (fo *TraceElementFork) GetTPre() int {
	return fo.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (fo *TraceElementFork) GetTPost() int {
	return fo.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (fo *TraceElementFork) GetTSort() int {
	return fo.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (fo *TraceElementFork) GetPos() string {
	return fmt.Sprintf("%s:%d", fo.file, fo.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (fo *TraceElementFork) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", fo.routine, fo.file, fo.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (fo *TraceElementFork) GetFile() string {
	return fo.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (fo *TraceElementFork) GetLine() int {
	return fo.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (fo *TraceElementFork) GetTID() string {
	return fo.GetPos() + "@" + strconv.Itoa(fo.tPost)
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (fo *TraceElementFork) SetVc(vc *clock.VectorClock) {
	fo.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (fo *TraceElementFork) SetWVc(vc *clock.VectorClock) {
	fo.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (fo *TraceElementFork) GetVC() *clock.VectorClock {
	return fo.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (fo *TraceElementFork) GetWVc() *clock.VectorClock {
	return fo.wVc
}

// GetObjType returns the string representation of the object type
func (fo *TraceElementFork) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeFork + "F"
	}
	return ObjectTypeFork
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (fo *TraceElementFork) IsEqual(elem TraceElement) bool {
	return fo.routine == elem.GetRoutine() && fo.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (fo *TraceElementFork) GetTraceIndex() (int, int) {
	return fo.routine, fo.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (fo *TraceElementFork) SetT(time int) {
	fo.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (fo *TraceElementFork) SetTPre(tPre int) {
	fo.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (fo *TraceElementFork) SetTSort(tPost int) {
	fo.SetTPre(tPost)
	fo.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (fo *TraceElementFork) SetTWithoutNotExecuted(tSort int) {
	fo.SetTPre(tSort)
	if fo.tPost != 0 {
		fo.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (fo *TraceElementFork) ToString() string {
	return "G" + "," + strconv.Itoa(fo.tPost) + "," + strconv.Itoa(fo.id) +
		"," + fo.GetPos()
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (fo *TraceElementFork) GetTraceID() int {
	return fo.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (fo *TraceElementFork) setTraceID(ID int) {
	fo.traceID = ID
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (fo *TraceElementFork) Copy() TraceElement {
	children := make([]TraceElement, len(fo.children))
	copy(children, fo.children)
	parents := make([]TraceElement, len(fo.parents))
	copy(parents, fo.parents)

	return &TraceElementFork{
		traceID:  fo.traceID,
		index:    fo.index,
		routine:  fo.routine,
		tPost:    fo.tPost,
		id:       fo.id,
		file:     fo.file,
		line:     fo.line,
		vc:       fo.vc.Copy(),
		wVc:      fo.wVc.Copy(),
		children: children,
		parents:  parents,
	}
}

// AddChild adds an element as a child of this node in the partial order graph
//
// Parameter:
//   - elem *TraceElement: the element to add
func (fo *TraceElementFork) AddChild(elem TraceElement) {
	fo.children = append(fo.children, elem)
}

// GetChildren returns all children of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the children
func (fo *TraceElementFork) GetChildren() []TraceElement {
	return fo.children
}

// GetParents returns all parents of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the parents
func (fo *TraceElementFork) GetParents() []TraceElement {
	return fo.children
}
