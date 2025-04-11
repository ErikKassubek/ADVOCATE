// Copyright (c) 2024 Erik Kassubek
//
// File: trace.go
// Brief: Functions and structs for the trace
//
// Author: Erik Kassubek
// Created: 2024-08-08
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/utils"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Trace struct {
	// the trace
	traces map[int][]TraceElement

	// hb clocks have been calculated for this trace
	hbWasCalc bool

	numberOfRoutines   int
	numberElemsInTrace map[int]int // routine -> number

	currentIndex map[int]int
}

/*
 * Create a new trace structure
 * Returns:
 * 	Trace: the new trace
 */
func NewTrace() Trace {
	return Trace{
		traces:             make(map[int][]TraceElement),
		hbWasCalc:          false,
		numberOfRoutines:   0,
		numberElemsInTrace: make(map[int]int),
		currentIndex:       make(map[int]int),
	}
}

/*
 * Add an element to the trace
 * Args:
 * 	elem (TraceElement): Element to add
 */
func (t *Trace) AddElement(elem TraceElement) {
	routine := elem.GetRoutine()
	t.traces[routine] = append(t.traces[routine], elem)
	t.numberElemsInTrace[routine]++
}

/*
 * Add an replay end element to a trace
 * Args:
 * 	ts (string): The timestamp of the event
 * 	exitCode (int): The exit code of the event
 * Returns:
 * 	error
 */
func (t *Trace) AddTraceElementReplay(ts int, exitCode int) error {
	elem := TraceElementReplay{
		tPost:    ts,
		exitCode: exitCode,
	}

	t.AddElement(&elem)

	return nil
}

/*
 * Helper functions to sort the trace by tSort
 */
type sortByTSort []TraceElement

// len function required for sorting
func (a sortByTSort) Len() int { return len(a) }

// swap function required for sorting
func (a sortByTSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// order function required for sorting
func (a sortByTSort) Less(i, j int) bool {
	return a[i].GetTSort() < a[j].GetTSort()
}

/*
 * Sort each routine of the trace by tpost
 */
func (t *Trace) Sort() {
	for routine, trace := range t.traces {
		sort.Sort(sortByTSort(trace))
		t.traces[routine] = trace
	}
}

/*
 * Sort each routine of the trace by tpost
 * Args:
 * 	routines ([]int): List of routines to sort. For routines that are not in the trace, do nothing
 */
func (t *Trace) SortRoutines(routines []int) {
	for _, routine := range routines {
		if trace, ok := t.traces[routine]; ok {
			sort.Sort(sortByTSort(trace))
			t.traces[routine] = trace
		}
	}
}

/*
 * Get the traces
 * Returns:
 * 	map[int][]traceElement: The traces
 */
func (t *Trace) GetTraces() map[int][]TraceElement {
	return t.traces
}

/*
 * Return the number of TraceElement with cap and len
 */
func (t *Trace) GetTraceSize() (int, int) {
	capTot := 0
	lenTot := 0
	for _, elems := range t.traces {
		capTot += cap(elems)
		lenTot += len(elems)
	}
	return capTot, lenTot
}

/*
 * Get the trace of the given routine
 * Args:
 * 	id (int): The id of the routine
 * Returns:
 * 	[]traceElement: The trace of the routine
 */
func (t *Trace) GetRoutineTrace(id int) []TraceElement {
	return t.traces[id]
}

/*
 * Given the file and line info, return the routine and index of the element
 * in trace.
 * Args:
 * 	tID (string): The tID of the element
 * Returns:
 * 	*TraceElement: The element
 * 	error: An error if the element does not exist
 */
func (t *Trace) GetTraceElementFromTID(tID string) (TraceElement, error) {
	if tID == "" {
		return nil, errors.New("tID is empty")
	}

	for routine, trace := range t.traces {
		for index, elem := range trace {
			if elem.GetTID() == tID {
				return t.traces[routine][index], nil
			}
		}
	}
	return nil, errors.New("Element " + tID + " does not exist")
}

/*
 * Given the bug info from the machine readable result file, return the element
 * in the trace.
 * Args:
 * 	bugArg (string): The bug info from the machine readable result file
 * Returns:
 * 	*TraceElement: The element
 * 	error: An error if the element does not exist
 */
func (t *Trace) GetTraceElementFromBugArg(bugArg string) (TraceElement, error) {
	splitArg := strings.Split(bugArg, ":")

	if splitArg[0] != "T" {
		return nil, errors.New("Bug argument is not a trace element (does not start with T): " + bugArg)
	}

	if len(splitArg) != 7 {
		return nil, errors.New("Bug argument is not a trace element (incorrect number of arguments): " + bugArg)
	}

	routine, err := strconv.Atoi(splitArg[1])
	if err != nil {
		return nil, errors.New("Could not parse routine from bug argument: " + bugArg)
	}

	tPre, err := strconv.Atoi(splitArg[3])
	if err != nil {
		return nil, errors.New("Could not parse tPre from bug argument: " + bugArg)
	}

	for index, elem := range t.traces[routine] {
		if elem.GetTPre() == tPre {
			return t.traces[routine][index], nil
		}
	}

	for routine, trace := range t.traces {
		for index, elem := range trace {
			if elem.GetTPre() == tPre {
				return t.traces[routine][index], nil
			}
		}
	}

	return nil, fmt.Errorf("Element %s not in trace", bugArg)
}

/*
 * Shorten the trace by removing all elements after the given time
 * Args:
 * 	time (int): The time to shorten the trace to
 * 	incl (bool): True if an element with the same time should stay included in the trace
 */
func (t *Trace) ShortenTrace(time int, incl bool) {
	for routine, trace := range t.traces {
		for index, elem := range trace {
			if incl && elem.GetTSort() > time {
				t.traces[routine] = t.traces[routine][:index]
				break
			}
			if !incl && elem.GetTSort() >= time {
				t.traces[routine] = t.traces[routine][:index]
				break
			}
		}
	}
}

/*
 * Remove the element with the given tID from the trace
 * Args:
 * 	tID (string): The tID of the element to remove
 */
func (t *Trace) RemoveElementFromTrace(tID string) {
	for routine, trace := range MainTrace.traces {
		for index, elem := range trace {
			if elem.GetTID() == tID {
				MainTrace.traces[routine] = append(MainTrace.traces[routine][:index], MainTrace.traces[routine][index+1:]...)
				break
			}
		}
	}
}

/*
 * Shorten the trace of the given routine by removing all elements after and equal the given time
 * Args:
 * 	routine (int): The routine to shorten
 * 	time (int): The time to shorten the trace to
 */
func (t *Trace) ShortenRoutine(routine int, time int) {
	for index, elem := range t.traces[routine] {
		if elem.GetTSort() >= time {
			t.traces[routine] = t.traces[routine][:index]
			break
		}
	}
}

/*
 * Shorten a given a routine to index
 * Args:
 * 	routine (int): the routine to shorten
 * 	index (int): the index to which it should be shortened
 * 	incl (bool): if true, the value a index will remain in the routine, otherwise it will be removed
 */
func (t *Trace) ShortenRoutineIndex(routine, index int, incl bool) {
	if incl {
		t.traces[routine] = t.traces[routine][:index+1]
	} else {
		t.traces[routine] = t.traces[routine][:index]
	}
}

/*
 * Set the number of routines
 * Args:
 * 	n (int): The number of routines
 */
func (t *Trace) SetNoRoutines(n int) {
	t.numberOfRoutines = n
}

/*
 * Get the number of routines
 * Return:
 * 	(int): The number of routines
 */
func (t *Trace) GetNoRoutines() int {
	return t.numberOfRoutines
}

/*
 * Get the next element from a trace
 * Update the current index of the trace
 * Returns:
 * 	(TraceElement) The element in the trace with the smallest TSort that
 * 		has not been returned yet
 */
func (t *Trace) getNextElement() TraceElement {
	// find the local trace, where the element on which currentIndex points to
	// has the smallest tpost
	minTSort := -1
	minRoutine := -1
	for routine, trace := range t.traces {
		// no more elements in the routine trace
		if t.currentIndex[routine] == -1 {
			continue
		}
		// ignore non executed operations
		tSort := trace[t.currentIndex[routine]].GetTSort()
		if tSort == 0 || tSort == math.MaxInt {
			continue
		}
		if minTSort == -1 || trace[t.currentIndex[routine]].GetTSort() < minTSort {
			minTSort = trace[t.currentIndex[routine]].GetTSort()
			minRoutine = routine
		}
	}

	// all elements have been processed
	if minRoutine == -1 {
		return nil
	}

	// return the element and increase the index
	element := t.traces[minRoutine][t.currentIndex[minRoutine]]
	t.increaseIndex(minRoutine)

	return element
}

/*
 * Get the last elements in each routine
 * Returns
 * 	[]TraceElements: List of elements that are the last element in a routine
 */
func (t *Trace) getLastElemPerRout() []TraceElement {
	res := make([]TraceElement, 0)
	for _, trace := range t.traces {
		if len(trace) == 0 {
			continue
		}

		res = append(res, trace[len(trace)-1])
	}

	return res
}

/*
 * Update the currentIndex value of a trace for a routine
 * Args:
 * 	routine (int): the routine to update
 */
func (t *Trace) increaseIndex(routine int) {
	if t.currentIndex[routine] == -1 {
		utils.LogError("Tried to increase index of -1 at routine ", routine)
	}
	t.currentIndex[routine]++
	if t.currentIndex[routine] >= len(t.traces[routine]) {
		t.currentIndex[routine] = -1
	}
}

/*
 * For a given waitgroup id, get the number of add and done operations that were
 * executed before a given time.
 * Args:
 * 	wgID (int): The id of the waitgroup
 * 	waitTime (int): The time to check
 * Returns:
 * 	int: The number of add operations
 * 	int: The number of done operations
 */
func (t *Trace) GetNrAddDoneBeforeTime(wgID int, waitTime int) (int, int) {
	nrAdd := 0
	nrDone := 0

	for _, trace := range t.traces {
		for _, elem := range trace {
			switch e := elem.(type) {
			case *TraceElementWait:
				if e.GetID() == wgID {
					if e.GetTPre() < waitTime {
						delta := e.GetDelta()
						if delta > 0 {
							nrAdd++
						} else if delta < 0 {
							nrDone++
						}
					}
				}
			}
		}
	}

	return nrAdd, nrDone
}

/*
 * Shift all elements with time greater or equal to startTSort by shift
 * Only shift forward
 * Args:
 * 	startTPre (int): The time to start shifting
 * 	shift (int): The shift
 */
func (t *Trace) ShiftTrace(startTPre int, shift int) bool {
	if shift <= 0 {
		return false
	}

	for routine, trace := range t.traces {
		for index, elem := range trace {
			if elem.GetTPre() >= startTPre {
				t.traces[routine][index].SetTWithoutNotExecuted(elem.GetTSort() + shift)
			}
		}
	}

	return true
}

/*
 * Shift all elements that are concurrent or HB-later than the element such
 * that they are after the element without changing the order of these elements
 * Args:
 * 	element (traceElement): The element
 */
func (t *Trace) ShiftConcurrentOrAfterToAfter(element TraceElement) {
	elemsToShift := make([]TraceElement, 0)
	minTime := -1

	for _, trace := range t.traces {
		for _, elem := range trace {
			if elem.GetTID() == element.GetTID() {
				continue
			}

			if !(clock.GetHappensBefore(elem.GetVC(), element.GetVC()) == clock.Before) {
				elemsToShift = append(elemsToShift, elem)
				if minTime == -1 || elem.GetTPre() < minTime {
					minTime = elem.GetTPre()
				}
			}
		}
	}

	distance := element.GetTPre() - minTime + 1

	for _, elem := range elemsToShift {
		tSort := elem.GetTPre()
		elem.SetT(tSort + distance)
	}
}

/*
 * Shift all elements that are concurrent or HB-later than the element such
 * that they are after the element without changeing the order of these elements
 * Only shift elements that are after start
 * Args:
 * 	element (traceElement): The element
 * 	start (traceElement): The time to start shifting (not including)
 */
func (t *Trace) ShiftConcurrentOrAfterToAfterStartingFromElement(element TraceElement, start int) {
	elemsToShift := make([]TraceElement, 0)
	minTime := -1
	maxNotMoved := 0

	for _, trace := range t.traces {
		for _, elem := range trace {
			if elem.GetTID() == element.GetTID() {
				continue
			}

			if !(clock.GetHappensBefore(elem.GetVC(), element.GetVC()) == clock.Before) {
				if elem.GetTPre() <= start {
					continue
				}

				elemsToShift = append(elemsToShift, elem)
				if minTime == -1 || elem.GetTPre() < minTime {
					minTime = elem.GetTPre()
				}
			} else {
				if maxNotMoved == 0 || elem.GetTPre() > maxNotMoved {
					maxNotMoved = elem.GetTPre()
				}
			}
		}
	}

	if element.GetTPost() == 0 {
		element.SetT(maxNotMoved + 1)
	}

	distance := element.GetTPre() - minTime + 1

	for _, elem := range elemsToShift {
		tSort := elem.GetTPre()
		elem.SetT(tSort + distance)
	}

}

/*
 * Shift the element to be after all elements, that are concurrent to it
 * Args:
 * 	element (traceElement): The element
 */
func (t *Trace) ShiftConcurrentToBefore(element TraceElement) {
	t.ShiftConcurrentOrAfterToAfterStartingFromElement(element, 0)
}

/*
 * Remove all elements that are concurrent to the element and have time greater or equal to tmin
 * Args:
 * 	element (traceElement): The element
 */
func (t *Trace) RemoveConcurrent(element TraceElement, tmin int) {
	for routine, trace := range t.traces {
		result := make([]TraceElement, 0)
		for _, elem := range trace {
			if elem.GetTSort() < tmin {
				result = append(result, elem)
				continue
			}

			if elem.GetTID() == element.GetTID() {
				result = append(result, elem)
				continue
			}

			if clock.GetHappensBefore(elem.GetVC(), element.GetVC()) != clock.Concurrent {
				result = append(result, elem)
			}
		}
		t.traces[routine] = result
	}
}

/*
 * Remove all elements that are concurrent to the element or must happen after the element
 * Args:
 * 	element (traceElement): The element
 */
func (t *Trace) RemoveConcurrentOrAfter(element TraceElement, tmin int) {
	for routine, trace := range t.traces {
		result := make([]TraceElement, 0)
		for _, elem := range trace {
			if elem.GetTSort() < tmin {
				result = append(result, elem)
				continue
			}

			if elem.GetTID() == element.GetTID() {
				result = append(result, elem)
				continue
			}

			if clock.GetHappensBefore(elem.GetVC(), element.GetVC()) != clock.Before {
				result = append(result, elem)
			}
		}
		t.traces[routine] = result
	}
}

/*
 * For each routine, get the earliest element that is concurrent to the element
 * Args:
 * 	element (traceElement): The element
 * Returns:
 * 	map[int]traceElement: The earliest concurrent element for each routine
 */
func (t *Trace) GetConcurrentEarliest(element TraceElement) map[int]TraceElement {
	concurrent := make(map[int]TraceElement)
	for routine, trace := range t.traces {
		for _, elem := range trace {
			if elem.GetTID() == element.GetTID() {
				continue
			}

			if clock.GetHappensBefore(element.GetVC(), elem.GetVC()) == clock.Concurrent {
				concurrent[routine] = elem
			}
		}
	}
	return concurrent
}

/*
 * Remove all elements that have a later tPost that the given tPost
 * Args:
 * 	tPost (int): Remove elements after tPost
 */
func (t *Trace) RemoveLater(tPost int) {
	for routine, trace := range t.traces {
		for i, elem := range trace {
			if elem.GetTPost() > tPost {
				t.traces[routine] = t.traces[routine][:i]
			}
		}
	}
}

/*
 * Shift all elements with time greater or equal to startTSort by shift
 * Only shift back
 * Args:
 * 	routine (int): The routine to shift
 * 	startTSort (int): The time to start shifting
 * 	shift (int): The shift
 * Returns:
 * 	bool: True if the shift was successful, false otherwise (shift <= 0)
 */
func (t *Trace) ShiftRoutine(routine int, startTSort int, shift int) bool {
	if shift <= 0 {
		return false
	}

	for index, elem := range t.traces[routine] {
		if elem.GetTPre() >= startTSort {
			t.traces[routine][index].SetTWithoutNotExecuted(elem.GetTSort() + shift)
		}
	}

	return true
}

/*
 * Get the partial trace of all element between startTime and endTime incluseve.
 * Args:
 *  startTime (int): The start time
 *  endTime (int): The end time
 * Returns:
 *  map[int][]TraceElement: The partial trace
 */
func (t *Trace) GetPartialTrace(startTime int, endTime int) map[int][]TraceElement {
	result := make(map[int][]TraceElement)
	for routine, trace := range t.traces {
		for index, elem := range trace {
			if _, ok := result[routine]; !ok {
				result[routine] = make([]TraceElement, 0)
			}
			time := elem.GetTSort()
			if time >= startTime && time <= endTime {
				result[routine] = append(result[routine], t.traces[routine][index])
			}
		}
	}

	return result
}

/*
 * Deep copy a trace
 * Returns:
 * 	The copy of the trace
 */
func (t *Trace) Copy() Trace {
	tracesCopy := make(map[int][]TraceElement)
	for routine, trace := range t.traces {
		tracesCopy[routine] = make([]TraceElement, len(trace))
		for i, elem := range trace {
			tracesCopy[routine][i] = elem.Copy()
		}
	}

	numberElemsInTraceCopy := make(map[int]int)
	for routine, elem := range t.numberElemsInTrace {
		numberElemsInTraceCopy[routine] = elem
	}

	currentIndexCopy := make(map[int]int)
	for routine, elem := range t.currentIndex {
		currentIndexCopy[routine] = elem
	}

	return Trace{
		traces:             tracesCopy,
		hbWasCalc:          t.hbWasCalc,
		numberOfRoutines:   t.numberOfRoutines,
		numberElemsInTrace: numberElemsInTraceCopy,
		currentIndex:       currentIndexCopy,
	}
}

/*
 * Print the trace sorted by tPost
 */
func (t *Trace) PrintTrace() {
	t.PrintTraceArgs([]string{}, false)
}

/*
* Print the trace sorted by tPost
* Args:
*   types: types of the elements to print. If empty, all elements will be printed
*   clocks: if true, the clocks will be printed
 */
func (t *Trace) PrintTraceArgs(types []string, clocks bool) {
	elements := make([]struct {
		string
		time   int
		thread int
		vc     *clock.VectorClock
		wVc    *clock.VectorClock
	}, 0)
	for _, tra := range t.traces {
		for _, elem := range tra {
			elemStr := elem.ToString()
			if len(types) == 0 || utils.Contains(types, elemStr[0:1]) {
				elements = append(elements, struct {
					string
					time   int
					thread int
					vc     *clock.VectorClock
					wVc    *clock.VectorClock
				}{elemStr, elem.GetTPost(), elem.GetRoutine(), elem.GetVC(), elem.GetwVc()})
			}
		}
	}

	// sort elements by timestamp
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].time < elements[j].time
	})

	for _, elem := range elements {
		if clocks {
			utils.LogInfo(elem.thread, elem.string, elem.vc.ToString(), elem.wVc.ToString())
			fmt.Println(elem.thread, elem.string, elem.vc.ToString(), elem.wVc.ToString())
		} else {
			utils.LogInfo(elem.thread, elem.string)
			fmt.Println(elem.thread, elem.string)
		}
	}
}

/*
 * Get all to element concurrent wait, broadcast and signal operations on the same condition variable
 * Args:
 * 	element (traceElement): The element
 * 	filter ([]string): The types of the elements to return
 * Returns:
 * 	[]*traceElement: The concurrent elements
 */
func (t *Trace) GetConcurrentWaitgroups(element TraceElement) map[string][]TraceElement {
	res := make(map[string][]TraceElement)
	res["broadcast"] = make([]TraceElement, 0)
	res["signal"] = make([]TraceElement, 0)
	res["wait"] = make([]TraceElement, 0)
	for _, trace := range t.traces {
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
 * Set the tSort for an element given by its index
 * Args:
 * 	tSort (int): the new tSort
 * 	routine (int): the routine of the element
 * 	index (int): the index of the element in its routine
 */
func (t *Trace) SetTSortAtIndex(tPost, routine, index int) {
	if len(t.traces[routine]) <= index {
		return
	}
	t.traces[routine][index].SetTSort(tPost)
}
