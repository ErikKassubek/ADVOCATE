// Copyrigth (c) 2024 Erik Kassubek
//
// File: traceElementCond.go
// Brief: Struct and functions for operations of conditional variables in the trace
//
// Author: Erik Kassubek
// Created: 2023-12-25
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
)

type OpCond int

const (
	WaitCondOp OpCond = iota
	SignalOp
	BroadcastOp
)

/*
 * TraceElementCond is a trace element for a condition variable
 * MARK: Struct
 * Fields:
 *   routine (int): The routine id
 *   tpre (int): The timestamp at the start of the event
 *   tpost (int): The timestamp at the end of the event
 *   id (int): The id of the condition variable
 *   opC (opCond): The operation on the condition variable
 *   file (string), line(int): The position of the condition variable operation in the code
 *   tID (string): The id of the trace element, contains the position and the tpre
 */
type TraceElementCond struct {
	index   int
	routine int
	tPre    int
	tPost   int
	id      int
	opC     OpCond
	file    string
	line    int
	vc      clock.VectorClock
}

/*
 * Create a new condition variable trace element
 * MARK: New
 * Args:
 *   routine (int): The routine id
 *   tPre (string): The timestamp at the start of the event
 *   tPost (string): The timestamp at the end of the event
 *   id (string): The id of the condition variable
 *   opC (string): The operation on the condition variable
 *   pos (string): The position of the condition variable operation in the code
 */
func AddTraceElementCond(routine int, tPre string, tPost string, id string, opN string, pos string) error {
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
	var op OpCond
	switch opN {
	case "W":
		op = WaitCondOp
	case "S":
		op = SignalOp
	case "B":
		op = BroadcastOp
	default:
		return errors.New("op is not a valid operation")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementCond{
		index:   numberElemsInTrace[routine],
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		opC:     op,
		file:    file,
		line:    line,
	}

	return AddElementToTrace(&elem)
}

// MARK: Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (co *TraceElementCond) GetID() int {
	return co.id
}

/*
 * Get the routine of the element
 * Returns:
 *   (int): The routine id
 */
func (co *TraceElementCond) GetRoutine() int {
	return co.routine
}

/*
 * Get the tpre of the element.
 * Returns:
 *   int: The tpre of the element
 */
func (co *TraceElementCond) GetTPre() int {
	return co.tPre
}

/*
 * Get the tpost of the element.
 * Returns:
 *   int: The tpost of the element
 */
func (co *TraceElementCond) GetTPost() int {
	return co.tPost
}

/*
 * Get the timer, that is used for sorting the trace
 * Returns:
 *   (int): The timer of the element
 */
func (co *TraceElementCond) GetTSort() int {
	t := co.tPre
	if co.opC == WaitCondOp {
		t = co.tPost
	}
	if t == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return t
}

/*
* Get the position of the operation.
* Returns:
*   string: The position of the element
 */
func (co *TraceElementCond) GetPos() string {
	return fmt.Sprintf("%s:%d", co.file, co.line)
}

func (co *TraceElementCond) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", co.routine, co.file, co.line)
}

func (co *TraceElementCond) GetFile() string {
	return co.file
}

func (co *TraceElementCond) GetLine() int {
	return co.line
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (co *TraceElementCond) GetTID() string {
	return co.GetPos() + "@" + strconv.Itoa(co.tPre)
}

/*
 * Get the operation of the element
 * Returns:
 *   (OpCond): The operation of the element
 */
func (co *TraceElementCond) GetOpCond() OpCond {
	return co.opC
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (co *TraceElementCond) GetVC() clock.VectorClock {
	return co.vc
}

/*
 * Get all to element concurrent wait, broadcast and signal operations on the same condition variable
 * Args:
 *   element (traceElement): The element
 *   filter ([]string): The types of the elements to return
 * Returns:
 *   []*traceElement: The concurrent elements
 */
func GetConcurrentWaitgroups(element TraceElement) map[string][]TraceElement {
	res := make(map[string][]TraceElement)
	res["broadcast"] = make([]TraceElement, 0)
	res["signal"] = make([]TraceElement, 0)
	res["wait"] = make([]TraceElement, 0)
	for _, trace := range traces {
		for _, elem := range trace {
			switch elem.(type) {
			case *TraceElementCond:
			default:
				continue
			}

			if elem.GetTID() == element.GetTID() {
				continue
			}

			e := elem.(*TraceElementCond)

			if e.opC == WaitCondOp {
				continue
			}

			if clock.GetHappensBefore(element.GetVC(), e.GetVC()) == clock.Concurrent {
				e := elem.(*TraceElementCond)
				if e.opC == SignalOp {
					res["signal"] = append(res["signal"], elem)
				} else if e.opC == BroadcastOp {
					res["broadcast"] = append(res["broadcast"], elem)
				} else if e.opC == WaitCondOp {
					res["wait"] = append(res["wait"], elem)
				}
			}
		}
	}
	return res
}

/*
 * Get the string representation of the object type
 */
func (co *TraceElementCond) GetObjType(operation bool) string {
	if !operation {
		return "D"
	}

	switch co.opC {
	case WaitCondOp:
		return "DW"
	case BroadcastOp:
		return "DB"
	case SignalOp:
		return "DS"
	}
	return "D"
}

func (co *TraceElementCond) IsEqual(elem TraceElement) bool {
	return co.routine == elem.GetRoutine() && co.ToString() == elem.ToString()
}

func (co *TraceElementCond) GetTraceIndex() (int, int) {
	return co.routine, co.index
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (co *TraceElementCond) SetT(time int) {
	co.tPre = time
	co.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (co *TraceElementCond) SetTPre(tPre int) {
	co.tPre = tPre
	if co.tPost != 0 && co.tPost < tPre {
		co.tPost = tPre
	}
}

/*
 * Set the timer that is used for sorting the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (co *TraceElementCond) SetTSort(tSort int) {
	co.SetTPre(tSort)
	if co.opC == WaitCondOp {
		co.tPost = tSort
	}
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (co *TraceElementCond) SetTWithoutNotExecuted(tSort int) {
	co.SetTPre(tSort)
	if co.opC == WaitCondOp {
		if co.tPost != 0 {
			co.tPost = tSort
		}
		return
	}
	if co.tPre != 0 {
		co.tPre = tSort
	}
	return
}

/*
 * Get the string representation of the element
 * MARK: ToString
 * Returns:
 *   (string): The string representation of the element
 */
func (co *TraceElementCond) ToString() string {
	res := "D,"
	res += strconv.Itoa(co.tPre) + "," + strconv.Itoa(co.tPost) + ","
	res += strconv.Itoa(co.id) + ","
	switch co.opC {
	case WaitCondOp:
		res += "W"
	case SignalOp:
		res += "S"
	case BroadcastOp:
		res += "B"
	}
	res += "," + co.GetPos()
	return res
}

/*
 * Update the vector clock of the trace and element
 * MARK: VectorClock
 */
func (co *TraceElementCond) updateVectorClock() {
	co.vc = currentVCHb[co.routine].Copy()

	switch co.opC {
	case WaitCondOp:
		CondWait(co, currentVCHb)
	case SignalOp:
		CondSignal(co, currentVCHb)
	case BroadcastOp:
		CondBroadcast(co, currentVCHb)
	}

}

// MARK: Copy

/*
 * Copy the element
 * Returns:
 *   (TraceElement): The copy of the element
 */
func (co *TraceElementCond) Copy() TraceElement {
	return &TraceElementCond{
		index:   co.index,
		routine: co.routine,
		tPre:    co.tPre,
		tPost:   co.tPost,
		id:      co.id,
		opC:     co.opC,
		file:    co.file,
		line:    co.line,
		vc:      co.vc.Copy(),
	}
}

// MARK: GoPie
func (co *TraceElementCond) AddRel1(_ TraceElement, _ int) {
	return
}

func (co *TraceElementCond) AddRel2(_ TraceElement) {
	return
}

func (co *TraceElementCond) GetRel1() []TraceElement {
	return make([]TraceElement, 0)
}

func (co *TraceElementCond) GetRel2() []TraceElement {
	return make([]TraceElement, 0)
}
