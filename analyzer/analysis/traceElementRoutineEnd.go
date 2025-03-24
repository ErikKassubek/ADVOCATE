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
*   index (int): Index in the routine
*   routine (int): The routine id
*   tpost (int): The timestamp at the end of the event
*   vc (clock.VectorClock): The vector clock
 */
type TraceElementRoutineEnd struct {
	index   int
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
		index:   numberElemsInTrace[routine],
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
func (re *TraceElementRoutineEnd) GetID() int {
	return 0
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (re *TraceElementRoutineEnd) GetRoutine() int {
	return re.routine
}

/*
 * Get the tpre of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpre of the element
 */
func (re *TraceElementRoutineEnd) GetTPre() int {
	return re.tPost
}

/*
 * Get the tpost of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpost of the element
 */
func (re *TraceElementRoutineEnd) GetTPost() int {
	return re.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (re *TraceElementRoutineEnd) GetTSort() int {
	return re.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The position of the element
 */
func (re *TraceElementRoutineEnd) GetPos() string {
	return ""
}

func (re *TraceElementRoutineEnd) GetReplayID() string {
	return ""
}

func (re *TraceElementRoutineEnd) GetFile() string {
	return ""
}

func (re *TraceElementRoutineEnd) GetLine() int {
	return 0
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (re *TraceElementRoutineEnd) GetTID() string {
	return ""
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (re *TraceElementRoutineEnd) GetVC() clock.VectorClock {
	return re.vc
}

/*
 * Get the string representation of the object type
 */
func (re *TraceElementRoutineEnd) GetObjType(operation bool) string {
	if !operation {
		return "R"
	}
	return "RE"
}

func (re *TraceElementRoutineEnd) IsEqual(elem TraceElement) bool {
	return re.routine == elem.GetRoutine() && re.ToString() == elem.ToString()
}

func (re *TraceElementRoutineEnd) GetTraceIndex() (int, int) {
	return re.routine, re.index
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (re *TraceElementRoutineEnd) SetT(time int) {
	re.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (re *TraceElementRoutineEnd) SetTPre(tPre int) {
	re.tPost = tPre
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (re *TraceElementRoutineEnd) SetTSort(tpost int) {
	re.SetTPre(tpost)
	re.tPost = tpost
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (re *TraceElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
	re.SetTPre(tSort)
	if re.tPost != 0 {
		re.tPost = tSort
	}
}

/*
 * Get the simple string representation of the element
 * MARK: ToString
 * Returns:
 *   string: The simple string representation of the element
 */
func (re *TraceElementRoutineEnd) ToString() string {
	return "E" + "," + strconv.Itoa(re.tPost)
}

/*
 * Update and calculate the vector clock of the element
 * MARK: VectorClock
 */
func (re *TraceElementRoutineEnd) updateVectorClock() {
	re.vc = currentVCHb[re.routine].Copy()
}

/*
 * Copy the element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (re *TraceElementRoutineEnd) Copy() TraceElement {
	return &TraceElementRoutineEnd{
		index:   re.index,
		routine: re.routine,
		tPost:   re.tPost,
		vc:      re.vc.Copy(),
	}
}

// MARK: GoPie
func (re *TraceElementRoutineEnd) AddRel1(_ TraceElement, _ int) {
	return
}

func (re *TraceElementRoutineEnd) AddRel2(_ TraceElement) {
	return
}

func (re *TraceElementRoutineEnd) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

func (re *TraceElementRoutineEnd) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
