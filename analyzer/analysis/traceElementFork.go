// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementFork.go
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
	"fmt"
	"strconv"
)

// TraceElementFork is a trace element for a go statement
// Fields:
//
//   - index (int): Index in the routine
//   - routine (int): The routine id
//   - tpost (int): The timestamp at the end of the event
//   - id (int): The id of the new go statement
//   - file (string), line(int): The position of the trace element in the file
type TraceElementFork struct {
	index   int
	routine int
	tPost   int
	id      int
	file    string
	line    int
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
	rel1    []TraceElement
	rel2    []TraceElement
}

// Create a new go statement trace element
//
// Parameter:
//   - routine (int): The routine id
//   - tPost (string): The timestamp at the end of the event
//   - id (string): The id of the new routine
//   - pos (string): The position of the trace element in the file
func AddTraceElementFork(routine int, tPost string, id string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementFork{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPost:   tPostInt,
		id:      idInt,
		file:    file,
		line:    line,
		rel1:    make([]TraceElement, 2),
		rel2:    make([]TraceElement, 0),
		vc:      clock.NewVectorClock(MainTrace.numberOfRoutines),
		wVc:     clock.NewVectorClock(MainTrace.numberOfRoutines),
	}
	AddElementToTrace(&elem)
	return nil
}

// Get the id of the element
//
// Returns:
//   - int: The id of the element
func (fo *TraceElementFork) GetID() int {
	return fo.id
}

// Get the routine of the element
//
// Returns:
//   - int: The routine of the element
func (fo *TraceElementFork) GetRoutine() int {
	return fo.routine
}

// Get the tpre of the element. For atomic elements, tpre and tpost are the same
//
// Returns:
//   - int: The tpre of the element
func (fo *TraceElementFork) GetTPre() int {
	return fo.tPost
}

// Get the tpost of the element. For atomic elements, tpre and tpost are the same
//
// Returns:
//   - int: The tpost of the element
func (fo *TraceElementFork) GetTPost() int {
	return fo.tPost
}

// Get the timer, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (fo *TraceElementFork) GetTSort() int {
	return fo.tPost
}

// Get the position of the operation.
//
// Returns:
//   - string: The position of the element
func (fo *TraceElementFork) GetPos() string {
	return fmt.Sprintf("%s:%d", fo.file, fo.line)
}

// Get the replay id of the element
//
// Returns:
//   - The replay id
func (fo *TraceElementFork) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", fo.routine, fo.file, fo.line)
}

// Get the file of the element
//
// Returns:
//   - The file of the element
func (fo *TraceElementFork) GetFile() string {
	return fo.file
}

// Get the rline of the element
//
// Returns:
//   - The line of the element
func (fo *TraceElementFork) GetLine() int {
	return fo.line
}

// Get the tID of the element.
//
// Returns:
//   - string: The tID of the element
func (fo *TraceElementFork) GetTID() string {
	return fo.GetPos() + "@" + strconv.Itoa(fo.tPost)
}

// Get the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (fo *TraceElementFork) GetVC() *clock.VectorClock {
	return fo.vc
}

// Get the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (fo *TraceElementFork) GetwVc() *clock.VectorClock {
	return fo.wVc
}

// Get the string representation of the object type
func (fo *TraceElementFork) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeFork + "F"
	}
	return ObjectTypeFork
}

// Given a trace element, check if it is equal to this element
//
// Parameter:
//   - elem (TraceElement): The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (fo *TraceElementFork) IsEqual(elem TraceElement) bool {
	return fo.routine == elem.GetRoutine() && fo.ToString() == elem.ToString()
}

// Get the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (fo *TraceElementFork) GetTraceIndex() (int, int) {
	return fo.routine, fo.index
}

// Set the tPre and tPost of the element
//
// Parameter:
//   - time (int): The tPre and tPost of the element
func (fo *TraceElementFork) SetT(time int) {
	fo.tPost = time
}

// Set the tpre of the element.
//
// Parameter:
//   - tPre (int): The tpre of the element
func (fo *TraceElementFork) SetTPre(tPre int) {
	fo.tPost = tPre
}

// Set the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort (int): The timer of the element
func (fo *TraceElementFork) SetTSort(tpost int) {
	fo.SetTPre(tpost)
	fo.tPost = tpost
}

// Set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort (int): The timer of the element
func (fo *TraceElementFork) SetTWithoutNotExecuted(tSort int) {
	fo.SetTPre(tSort)
	if fo.tPost != 0 {
		fo.tPost = tSort
	}
}

// Get the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (fo *TraceElementFork) ToString() string {
	return "G" + "," + strconv.Itoa(fo.tPost) + "," + strconv.Itoa(fo.id) +
		"," + fo.GetPos()
}

// Update and calculate the vector clock of the element
func (fo *TraceElementFork) updateVectorClock() {
	fo.vc = currentVC[fo.routine].Copy()
	fo.wVc = currentWVC[fo.routine].Copy()

	Fork(fo)
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (fo *TraceElementFork) Copy() TraceElement {
	return &TraceElementFork{
		index:   fo.index,
		routine: fo.routine,
		tPost:   fo.tPost,
		id:      fo.id,
		file:    fo.file,
		line:    fo.line,
		vc:      fo.vc.Copy(),
		wVc:     fo.wVc.Copy(),
		rel1:    fo.rel1,
		rel2:    fo.rel2,
	}
}

// ========= For GoPie fuzzing ===========

// Add an element to the rel1 set of the element
//
// Parameter:
//   - elem (TraceElement): elem to add
//   - pos (int): before (0) or after (1)
func (fo *TraceElementFork) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}
	fo.rel1[pos] = elem
}

// Add an element to the rel2 set of the element
//
// Parameter:
//   - elem (TraceElement): elem to add
func (fo *TraceElementFork) AddRel2(elem TraceElement) {
	fo.rel2 = append(fo.rel2, elem)
}

// Return the rel1 set
//
// Returns:
//   - []*TraceElement: the rel1 set
func (fo *TraceElementFork) GetRel1() []TraceElement {
	return fo.rel1
}

// Return the rel2 set
//
// Returns:
//   - []*TraceElement: the rel2 set
func (fo *TraceElementFork) GetRel2() []TraceElement {
	return fo.rel2
}
