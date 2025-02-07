// Copyrigth (c) 2024 Erik Kassubek
//
// File: traceElementReplayStop.go
// Brief: Struct and functions for replay controll elements in the trace
//
// Author: Erik Kassubek
// Created: 2024-04-03
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"strconv"
)

/*
* Struct to save an atomic event in the trace
* MARK: Struct
* Fields:
*   tpost (int): The timestamp of the event
*   exitCode (int): expected exit code
*   lastElemTPre (int): tpre of the last elem, e.g. the tpre of the stuck elem in leak
 */
type TraceElementReplay struct {
	tPost        int
	exitCode     int
	lastElemTPre int
}

/*
 * Create an end of replay event
 * MARK: New
 * Args:
 *   t (string): The timestamp of the event
 *   exitCode (int): The exit code of the event
 *   lastElemT (int): TPre of the
 */
func AddTraceElementReplay(t int, exitCode int, lastElemTPre int) error {
	elem := TraceElementReplay{
		tPost:        t,
		exitCode:     exitCode,
		lastElemTPre: lastElemTPre,
	}

	return AddElementToTrace(&elem)
}

// MARK: Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (at *TraceElementReplay) GetID() int {
	return 0
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (at *TraceElementReplay) GetRoutine() int {
	return 1
}

/*
 * Get the tpost of the element.
 *   int: The tpost of the element
 */
func (at *TraceElementReplay) GetTPre() int {
	return at.tPost
}

/*
 * Get the tpost of the element.
 * Returns:
 *   int: The tpost of the element
 */
func (at *TraceElementReplay) GetTPost() int {
	return at.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (at *TraceElementReplay) GetTSort() int {
	return at.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The file of the element
 */
func (at *TraceElementReplay) GetPos() string {
	return ""
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (at *TraceElementReplay) GetTID() string {
	return ""
}

/*
 * Dummy function to implement the interface
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (at *TraceElementReplay) GetVC() clock.VectorClock {
	return clock.VectorClock{}
}

/*
 * Get the string representation of the object type
 */
func (at *TraceElementReplay) GetObjType() string {
	return "RR"
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (mu *TraceElementReplay) SetT(time int) {
	mu.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (mu *TraceElementReplay) SetTPre(tPre int) {
	tPre = max(1, tPre)
	mu.tPost = tPre
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (at *TraceElementReplay) SetTSort(tSort int) {
	tSort = max(1, tSort)
	at.SetTPre(tSort)
	at.tPost = tSort
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (at *TraceElementReplay) SetTWithoutNotExecuted(tSort int) {
	tSort = max(1, tSort)
	at.SetTPre(tSort)
	at.tPost = tSort
}

/*
 * Get the simple string representation of the element.
 * MARK: ToString
 * Returns:
 *   string: The simple string representation of the element
 */
func (at *TraceElementReplay) ToString() string {
	res := "X," + strconv.Itoa(at.tPost) + "," + strconv.Itoa(at.exitCode) + "," + strconv.Itoa(at.lastElemTPre)
	return res
}

/*
 * Update and calculate the vector clock of the element
 * MARK: VectorClock
 */
func (at *TraceElementReplay) updateVectorClock() {
	// nothing to do
}

// MARK: Copy

/*
 * Create a copy of the element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (at *TraceElementReplay) Copy() TraceElement {
	return &TraceElementReplay{
		tPost:        at.tPost,
		exitCode:     at.exitCode,
		lastElemTPre: at.lastElemTPre,
	}
}
