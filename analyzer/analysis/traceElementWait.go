// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementWait.go
// Brief: Struct and functions for wait group operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/utils"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// enum for opW
type opW int

const (
	ChangeOp opW = iota
	WaitOp
)

/*
 * TraceElementWait is a trace element for a wait group statement
 * Fields:
 *   index (int): Index in the routine
 *   tpre (int): The timestamp at the start of the event
 *   tpost (int): The timestamp at the end of the event
 *   id (int): The id of the wait group
 *   opW (opW): The operation on the wait group
 *   delta (int): The delta of the wait group
 *   val (int): The value of the wait group
 *   file (string), line(int): The position of the wait group in the code
 *   tID (string): The id of the trace element, contains the position and the tpre
 */
type TraceElementWait struct {
	index   int
	routine int
	tPre    int
	tPost   int
	id      int
	opW     opW
	delta   int
	val     int
	file    string
	line    int
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
	rel1    []TraceElement
	rel2    []TraceElement
}

/*
 * Create a new wait group trace element
 * Args:
 * 	routine (int): The routine id
 * 	tpre (string): The timestamp at the start of the event
 * 	tpost (string): The timestamp at the end of the event
 * 	id (string): The id of the wait group
 * 	opW (string): The operation on the wait group
 * 	delta (string): The delta of the wait group
 * 	val (string): The value of the wait group
 * 	pos (string): The position of the wait group in the code
 */
func AddTraceElementWait(routine int, tpre,
	tpost, id, opW, delta, val, pos string) error {
	tpre_int, err := strconv.Atoi(tpre)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	tpost_int, err := strconv.Atoi(tpost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	id_int, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	opW_op := ChangeOp
	if opW == "W" {
		opW_op = WaitOp
	} else if opW != "A" {
		return errors.New("op is not a valid operation")
	}

	delta_int, err := strconv.Atoi(delta)
	if err != nil {
		return errors.New("delta is not an integer")
	}

	val_int, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("val is not an integer")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementWait{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPre:    tpre_int,
		tPost:   tpost_int,
		id:      id_int,
		opW:     opW_op,
		delta:   delta_int,
		val:     val_int,
		file:    file,
		line:    line,
		vc:      clock.NewVectorClock(MainTrace.numberOfRoutines),
		wVc:     clock.NewVectorClock(MainTrace.numberOfRoutines),
		rel1:    make([]TraceElement, 2),
		rel2:    make([]TraceElement, 0),
	}

	AddElementToTrace(&elem)

	return nil
}

/*
 * Get the id of the element
 * Returns:
 * 	int: The id of the element
 */
func (wa *TraceElementWait) GetID() int {
	return wa.id
}

/*
 * Get the routine of the element
 * Returns:
 * 	int: The routine of the element
 */
func (wa *TraceElementWait) GetRoutine() int {
	return wa.routine
}

/*
 * Get the timestamp at the start of the event
 * Returns:
 * 	int: The timestamp at the start of the event
 */
func (wa *TraceElementWait) GetTPre() int {
	return wa.tPre
}

/*
 * Get the timestamp at the end of the event
 * Returns:
 * 	int: The timestamp at the end of the event
 */
func (wa *TraceElementWait) GetTPost() int {
	return wa.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 * 	int: The timer of the element
 */
func (wa *TraceElementWait) GetTSort() int {
	if wa.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return wa.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 * 	string: The position of the element
 */
func (wa *TraceElementWait) GetPos() string {
	return fmt.Sprintf("%s:%d", wa.file, wa.line)
}

func (wa *TraceElementWait) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", wa.routine, wa.file, wa.line)
}

func (wa *TraceElementWait) GetFile() string {
	return wa.file
}

func (wa *TraceElementWait) GetLine() int {
	return wa.line
}

/*
 * Get the tID of the element.
 * Returns:
 * 	string: The tID of the element
 */
func (wa *TraceElementWait) GetTID() string {
	return wa.GetPos() + "@" + strconv.Itoa(wa.tPre)
}

/*
 * Get if the operation is a wait op
 * Returns:
 * 	bool: True if the operation is a wait op
 */
func (wa *TraceElementWait) IsWait() bool {
	return wa.opW == WaitOp
}

func (wa *TraceElementWait) GetDelta() int {
	return wa.delta
}

/*
 * Get the vector clock of the element
 * Returns:
 * 	VectorClock: The vector clock of the element
 */
func (wa *TraceElementWait) GetVC() *clock.VectorClock {
	return wa.vc
}

func (wa *TraceElementWait) GetwVc() *clock.VectorClock {
	return wa.wVc
}

/*
 * Get the string representation of the object type
 */
func (wa *TraceElementWait) GetObjType(operation bool) string {
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

func (wa *TraceElementWait) IsEqual(elem TraceElement) bool {
	return wa.routine == elem.GetRoutine() && wa.ToString() == elem.ToString()
}

func (wa *TraceElementWait) GetTraceIndex() (int, int) {
	return wa.routine, wa.index
}

/*
 * Set the tPre and tPost of the element
 * Args:
 * 	time (int): The tPre and tPost of the element
 */
func (wa *TraceElementWait) SetT(time int) {
	wa.tPre = time
	wa.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 * 	tPre (int): The tpre of the element
 */
func (wa *TraceElementWait) SetTPre(tPre int) {
	wa.tPre = tPre
	if wa.tPost != 0 && wa.tPost < tPre {
		wa.tPost = tPre
	}
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 * 	tSort (int): The timer of the element
 */
func (wa *TraceElementWait) SetTSort(tSort int) {
	wa.SetTPre(tSort)
	wa.tPost = tSort
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 * 	tSort (int): The timer of the element
 */
func (wa *TraceElementWait) SetTWithoutNotExecuted(tSort int) {
	wa.SetTPre(tSort)
	if wa.tPost != 0 {
		wa.tPost = tSort
	}
}

/*
 * Get the simple string representation of the element
 * Returns:
 * 	string: The simple string representation of the element
 */
func (wa *TraceElementWait) ToString() string {
	res := "W,"
	res += strconv.Itoa(wa.tPre) + "," + strconv.Itoa(wa.tPost) + ","
	res += strconv.Itoa(wa.id) + ","
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

/*
 * Update and calculate the vector clock of the element
 */
func (wa *TraceElementWait) updateVectorClock() {
	wa.vc = currentVC[wa.routine].Copy()
	wa.wVc = currentWVC[wa.routine].Copy()

	switch wa.opW {
	case ChangeOp:
		Change(wa)
	case WaitOp:
		Wait(wa)
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		utils.LogError(err)
	}
}

/*
 * Copy the element
 * Returns:
 * 	TraceElement: The copy of the element
 */
func (wa *TraceElementWait) Copy() TraceElement {
	return &TraceElementWait{
		index:   wa.index,
		routine: wa.routine,
		tPre:    wa.tPre,
		tPost:   wa.tPost,
		id:      wa.id,
		opW:     wa.opW,
		delta:   wa.delta,
		val:     wa.val,
		file:    wa.file,
		line:    wa.line,
		vc:      wa.vc.Copy(),
		wVc:     wa.wVc.Copy(),
		rel1:    wa.rel1,
		rel2:    wa.rel1,
	}
}

// ========= For GoPie fuzzing ===========

/*
 * Add an element to the rel1 set of the element
 * Args:
 * 	elem (TraceElement): elem to add
 * 	pos (int): before (0) or after (1)
 */

func (wa *TraceElementWait) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}
	wa.rel1[pos] = elem
}

/*
 * Add an element to the rel2 set of the element
 * Args:
 * 	elem (TraceElement): elem to add
 */
func (wa *TraceElementWait) AddRel2(elem TraceElement) {
	wa.rel2 = append(wa.rel2, elem)
}

/*
 * Return the rel1 set
 * Returns:
 * 	[]*TraceElement: the rel1 set
 */
func (wa *TraceElementWait) GetRel1() []TraceElement {
	return wa.rel1
}

/*
 * Return the rel2 set
 * Returns:
 * 	[]*TraceElement: the rel2 set
 */
func (wa *TraceElementWait) GetRel2() []TraceElement {
	return wa.rel1
}
