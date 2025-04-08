// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementOnce.go
// Brief: Struct and functions for once operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-09-25
//
// License: BSD-3-Clause

package analysis

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"analyzer/clock"
)

/*
 * traceElementMutex is a trace element for a once
 * MARK: Struct
 * Fields:
 *   routine (int): The routine id
 *   tpre (int): The timestamp at the start of the event
 *   tpost (int): The timestamp at the end of the event
 *   id (int): The id of the mutex
 *   suc (bool): Whether the operation was successful
 *   file (string), line (int): The position of the mutex operation in the code
 */
type TraceElementOnce struct {
	index   int
	routine int
	tPre    int
	tPost   int
	id      int
	suc     bool
	file    string
	line    int
	vc      clock.VectorClock
}

/*
 * Create a new mutex trace element
 * MARK: New
 * Args:
 *   routine (int): The routine id
 *   tPre (string): The timestamp at the start of the event
 *   tPost (string): The timestamp at the end of the event
 *   id (string): The id of the mutex
 *   suc (string): Whether the operation was successful (only for trylock else always true)
 *   pos (string): The position of the mutex operation in the code
 */
func AddTraceElementOnce(routine int, tPre string,
	tPost string, id string, suc string, pos string) error {
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

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementOnce{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		suc:     sucBool,
		file:    file,
		line:    line,
	}

	AddElementToTrace(&elem)

	return nil
}

// MARK: Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (on *TraceElementOnce) GetID() int {
	return on.id
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (on *TraceElementOnce) GetRoutine() int {
	return on.routine
}

/*
 * Get the tpre of the element.
 * Returns:
 *   int: The tpre of the element
 */
func (on *TraceElementOnce) GetTPre() int {
	return on.tPre
}

/*
 * Get the tpost of the element.
 * Returns:
 *   int: The tpost of the element
 */
func (on *TraceElementOnce) GetTPost() int {
	return on.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (on *TraceElementOnce) GetTSort() int {
	if on.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return on.tPre
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The position of the element
 */
func (on *TraceElementOnce) GetPos() string {
	return fmt.Sprintf("%s:%d", on.file, on.line)
}

func (on *TraceElementOnce) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", on.routine, on.file, on.line)
}

func (on *TraceElementOnce) GetFile() string {
	return on.file
}

func (on *TraceElementOnce) GetLine() int {
	return on.line
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (on *TraceElementOnce) GetTID() string {
	return on.GetPos() + "@" + strconv.Itoa(on.tPre)
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (on *TraceElementOnce) GetVC() clock.VectorClock {
	return on.vc
}

/*
 * Get the string representation of the object type
 */
func (on *TraceElementOnce) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeOnce
	}

	if on.suc {
		return ObjectTypeOnce + "E"
	}
	return ObjectTypeOnce + "N"
}

/*
 * Get whether the once do was executed (successful)
 */
func (on *TraceElementOnce) GetSuc() bool {
	return on.suc
}

func (on *TraceElementOnce) IsEqual(elem TraceElement) bool {
	return on.routine == elem.GetRoutine() && on.ToString() == elem.ToString()
}

func (on *TraceElementOnce) GetTraceIndex() (int, int) {
	return on.routine, on.index
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (on *TraceElementOnce) SetT(time int) {
	on.tPre = time
	on.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (on *TraceElementOnce) SetTPre(tPre int) {
	on.tPre = tPre
	if on.tPost != 0 && on.tPost < tPre {
		on.tPost = tPre
	}
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (on *TraceElementOnce) SetTSort(tSort int) {
	on.SetTPre(tSort)
	on.tPost = tSort
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (on *TraceElementOnce) SetTWithoutNotExecuted(tSort int) {
	on.SetTPre(tSort)
	if on.tPost != 0 {
		on.tPost = tSort
	}
}

/*
 * Get the simple string representation of the element
 * MARK: ToString
 * Returns:
 *   string: The simple string representation of the element
 */
func (on *TraceElementOnce) ToString() string {
	res := "O,"
	res += strconv.Itoa(on.tPre) + ","
	res += strconv.Itoa(on.tPost) + ","
	res += strconv.Itoa(on.id) + ","
	if on.suc {
		res += "t"
	} else {
		res += "f"
	}
	res += "," + on.GetPos()
	return res
}

/*
 * Update the vector clock of the trace and element
 * MARK: VectorClock
 */
func (on *TraceElementOnce) updateVectorClock() {
	on.vc = currentVCHb[on.routine].Copy()

	if on.suc {
		DoSuc(on, currentVCHb)
	} else {
		DoFail(on, currentVCHb)
	}

}

/*
 * Copy the element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (on *TraceElementOnce) Copy() TraceElement {
	return &TraceElementOnce{
		index:   on.index,
		routine: on.routine,
		tPre:    on.tPre,
		tPost:   on.tPost,
		id:      on.id,
		suc:     on.suc,
		file:    on.file,
		line:    on.line,
		vc:      on.vc.Copy(),
	}
}

// MARK: GoPie
func (on *TraceElementOnce) AddRel1(_ TraceElement, _ int) {
	return
}

func (on *TraceElementOnce) AddRel2(_ TraceElement) {
	return
}

func (on *TraceElementOnce) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

func (on *TraceElementOnce) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
