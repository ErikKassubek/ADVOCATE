// Copyrigth (c) 2024 Erik Kassubek
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

/*
* TraceElementFork is a trace element for a go statement
* MARK: Struct
* Fields:
*   routine (int): The routine id
*   tpost (int): The timestamp at the end of the event
*   id (int): The id of the new go statement
*   file (string), line(int): The position of the trace element in the file
 */
type TraceElementFork struct {
	routine int
	tPost   int
	id      int
	file    string
	line    int
	vc      clock.VectorClock
	vcWmHB  clock.VectorClock
}

/*
 * Create a new go statement trace element
 * MARK: New
 * Args:
 *   routine (int): The routine id
 *   tPost (string): The timestamp at the end of the event
 *   id (string): The id of the new routine
 *   pos (string): The position of the trace element in the file
 */
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
		routine: routine,
		tPost:   tPostInt,
		id:      idInt,
		file:    file,
		line:    line,
	}
	return AddElementToTrace(&elem)
}

// MARK Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (fo *TraceElementFork) GetID() int {
	return fo.id
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (fo *TraceElementFork) GetRoutine() int {
	return fo.routine
}

/*
 * Get the tpre of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpre of the element
 */
func (fo *TraceElementFork) GetTPre() int {
	return fo.tPost
}

/*
 * Get the tpost of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpost of the element
 */
func (fo *TraceElementFork) GetTPost() int {
	return fo.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (fo *TraceElementFork) GetTSort() int {
	return fo.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The position of the element
 */
func (fo *TraceElementFork) GetPos() string {
	return fmt.Sprintf("%s:%d", fo.file, fo.line)
}

func (fo *TraceElementFork) GetFile() string {
	return fo.file
}

func (fo *TraceElementFork) GetLine() int {
	return fo.line
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (fo *TraceElementFork) GetTID() string {
	return fo.GetPos() + "@" + strconv.Itoa(fo.tPost)
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (fo *TraceElementFork) GetVC() clock.VectorClock {
	return fo.vc
}

func (fo *TraceElementFork) GetVCWmHB() clock.VectorClock {
	return fo.vcWmHB
}

/*
 * Get the string representation of the object type
 */
func (fo *TraceElementFork) GetObjType() string {
	return "GF"
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (fo *TraceElementFork) SetT(time int) {
	fo.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (fo *TraceElementFork) SetTPre(tPre int) {
	fo.tPost = tPre
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (fo *TraceElementFork) SetTSort(tpost int) {
	fo.SetTPre(tpost)
	fo.tPost = tpost
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (fo *TraceElementFork) SetTWithoutNotExecuted(tSort int) {
	fo.SetTPre(tSort)
	if fo.tPost != 0 {
		fo.tPost = tSort
	}
}

/*
 * Get the simple string representation of the element
 * MARK: ToString
 * Returns:
 *   string: The simple string representation of the element
 */
func (fo *TraceElementFork) ToString() string {
	return "G" + "," + strconv.Itoa(fo.tPost) + "," + strconv.Itoa(fo.id) +
		"," + fo.GetPos()
}

/*
 * Update and calculate the vector clock of the element
 * MARK: VectorClock
 */
func (fo *TraceElementFork) updateVectorClock() {
	fo.vc = currentVCHb[fo.routine].Copy()
	fo.vcWmHB = currentVCWmhb[fo.routine].Copy()

	Fork(fo, currentVCHb, currentVCWmhb)
}

/*
 * Copy the element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (fo *TraceElementFork) Copy() TraceElement {
	return &TraceElementFork{
		routine: fo.routine,
		tPost:   fo.tPost,
		id:      fo.id,
		file:    fo.file,
		line:    fo.line,
		vc:      fo.vc.Copy(),
		vcWmHB:  fo.vcWmHB.Copy(),
	}
}
