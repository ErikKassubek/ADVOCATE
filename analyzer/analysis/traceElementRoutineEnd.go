// Copyrigth (c) 2024 Erik Kassubek
//
// File: TraceElementRoutineEnd.go
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
	"strconv"
)

/*
* TraceElementRoutineEnd is a trace element for the termination of a routine end
* MARK: Struct
* Fields:
*   routine (int): The routine id
*   tpost (int): The timestamp at the end of the event
 */
type TraceElementRoutineEnd struct {
	routine int
	tPost   int
	vc      clock.VectorClock
}

/*
 * End a routine
 * MARK: New
 * Args:
 *   routine (int): The routine id
 *   tPost (string): The timestamp at the end of the event
 *   id (string): The id of the new routine
 *   pos (string): The position of the trace element in the file
 */
func AddTraceElementRoutineEnd(routine int, tPost string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	elem := TraceElementRoutineEnd{
		routine: routine,
		tPost:   tPostInt,
	}
	return AddElementToTrace(&elem)
}

// MARK Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (fo *TraceElementRoutineEnd) GetID() int {
	return 0
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (fo *TraceElementRoutineEnd) GetRoutine() int {
	return fo.routine
}

/*
 * Get the tpre of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpre of the element
 */
func (fo *TraceElementRoutineEnd) GetTPre() int {
	return fo.tPost
}

/*
 * Get the tpost of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpost of the element
 */
func (fo *TraceElementRoutineEnd) GetTPost() int {
	return fo.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (fo *TraceElementRoutineEnd) GetTSort() int {
	return fo.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The position of the element
 */
func (fo *TraceElementRoutineEnd) GetPos() string {
	return ""
}

func (fo *TraceElementRoutineEnd) GetFile() string {
	return ""
}

func (fo *TraceElementRoutineEnd) GetLine() int {
	return 0
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (fo *TraceElementRoutineEnd) GetTID() string {
	return ""
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (fo *TraceElementRoutineEnd) GetVC() clock.VectorClock {
	return fo.vc
}

/*
 * Get the string representation of the object type
 */
func (fo *TraceElementRoutineEnd) GetObjType() string {
	return "GE"
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (fo *TraceElementRoutineEnd) SetT(time int) {
	fo.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (fo *TraceElementRoutineEnd) SetTPre(tPre int) {
	fo.tPost = tPre
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (fo *TraceElementRoutineEnd) SetTSort(tpost int) {
	fo.SetTPre(tpost)
	fo.tPost = tpost
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (fo *TraceElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
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
func (fo *TraceElementRoutineEnd) ToString() string {
	return "E" + "," + strconv.Itoa(fo.tPost)
}

/*
 * Update and calculate the vector clock of the element
 * MARK: VectorClock
 */
func (fo *TraceElementRoutineEnd) updateVectorClock() {
	fo.vc = currentVCHb[fo.routine].Copy()
}

/*
 * Copy the element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (fo *TraceElementRoutineEnd) Copy() TraceElement {
	return &TraceElementRoutineEnd{
		routine: fo.routine,
		tPost:   fo.tPost,
		vc:      fo.vc.Copy(),
	}
}
