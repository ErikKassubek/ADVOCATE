// Copyright (c) 2024 Erik Kassubek
//
// File: /gocdr/trace/trace.go
// Brief: Functions and structs for the trace
//
// Author: Erik Kassubek
// Created: 2024-08-08
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"gocdr/utils/control"
	"gocdr/utils/log"
	"gocdr/utils/types"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Trace is a struct to represent a trace
// Fields:
//   - traces map[int][]TraceElement: the trace element, routineId -> list of elems
//   - hbWasCalc bool: set to true if the vector clock has been calculated for all elements
//   - channelWithoutPartner  map[int]map[int]*TraceElementChannel: channel for witch no partner has been found yet, id -> opId -> element
//   - channelIDs map[int]struct{}: all channel ids in the trace
type Trace struct {
	traces                map[int][]Element
	minTraceID            int
	channelWithoutPartner map[int]map[int]*ElementChannel
	channelIDs            map[int]struct{}
}

// NewTrace creates a new empty trace structure
//
// Returns:
//   - Trace: the new trace
func NewTrace() Trace {
	return Trace{
		traces:                make(map[int][]Element),
		minTraceID:            0,
		channelWithoutPartner: make(map[int]map[int]*ElementChannel),
	}
}

// Clear the trace
func (this *Trace) Clear() {
	this.traces = make(map[int][]Element)
	this.minTraceID = 0
}

// AddElement adds an element to the trace
//
// Parameter:
//   - elem TraceElement: Element to add
func (this *Trace) AddElement(elem Element) {
	routine := elem.GetRoutine()

	this.minTraceID++
	elem.setID(this.minTraceID)

	this.traces[routine] = append(this.traces[routine], elem)
}

// AddRoutine adds an empty routine if not exists
//
// Parameter:
//   - routine int: The routine
func (this *Trace) AddRoutine(routine int) {
	if _, ok := this.traces[routine]; !ok {
		this.traces[routine] = make([]Element, 0)
	}
}

// Helper functions to sort the trace by tSort
type sortByTSort []Element

// len function required for sorting
func (a sortByTSort) Len() int { return len(a) }

// swap function required for sorting
func (a sortByTSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// order function required for sorting
func (a sortByTSort) Less(i, j int) bool {
	return a[i].GetTSort() < a[j].GetTSort()
}

// Sort each routine of the trace by tPost
func (this *Trace) Sort() {
	for routine, trace := range this.traces {
		sort.Sort(sortByTSort(trace))
		this.traces[routine] = trace
	}
}

// SortRoutines sort each routine of the trace by tPost
//
// Parameter:
//   - routines []int: List of routines to sort. For routines that are not in the trace, do nothing
func (this *Trace) SortRoutines(routines []int) {
	for _, routine := range routines {
		if trace, ok := this.traces[routine]; ok {
			sort.Sort(sortByTSort(trace))
			this.traces[routine] = trace
		}
	}
}

// GetTraces returns the traces
//
// Returns:
//   - map[int][]traceElement: The traces
func (this *Trace) GetTraces() map[int][]Element {
	return this.traces
}

// GetTraceSize returns the number of TraceElement with cap and len
func (this *Trace) GetTraceSize() (int, int) {
	capTot := 0
	lenTot := 0
	for _, elems := range this.traces {
		capTot += cap(elems)
		lenTot += len(elems)
	}
	return capTot, lenTot
}

// GetRoutineTrace returns the trace of the given routine
//
// Parameter:
//   - id int: The id of the routine
//
// Returns:
//   - []traceElement: The trace of the routine
func (this *Trace) GetRoutineTrace(id int) []Element {
	return this.traces[id]
}

// GetNumberElements returns the total number of elements in the trace
//
// Returns:
//   - int: total number of elems in trace
func (this *Trace) GetNumberElements() int {
	return this.NumberElemInRoutine(-1)
}

// GetTraceElementFromTID returns the routine and index of the element
// in trace, given the file and line info.
//
// Parameter:
//   - tID string: The tID of the element
//
// Returns:
//   - *TraceElement: The element
//   - error: An error if the element does not exist
func (this *Trace) GetTraceElementFromTID(tID string) (Element, error) {
	if tID == "" {
		return nil, errors.New("tID is empty")
	}

	for routine, trace := range this.traces {
		for index, elem := range trace {
			if elem.GetTID() == tID {
				return this.traces[routine][index], nil
			}
		}
	}
	return nil, errors.New("Element " + tID + " does not exist")
}

// GetTraceElementFromBugArg returns the element in the trace,
// given the bug info from the machine readable result file.
//
// Parameter:
//   - bugArg string: The bug info from the machine readable result file
//
// Returns:
//   - *TraceElement: The element
//   - error: An error if the element does not exist
func (this *Trace) GetTraceElementFromBugArg(bugArg string) (Element, error) {
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

	for index, elem := range this.traces[routine] {
		if elem.GetTPre() == tPre {
			return this.traces[routine][index], nil
		}
	}

	for routine, trace := range this.traces {
		for index, elem := range trace {
			if elem.GetTPre() == tPre {
				return this.traces[routine][index], nil
			}
		}
	}

	return nil, fmt.Errorf("Element %s not in trace", bugArg)
}

// ShortenTrace shortens the trace by removing all elements after the given time
//
// Parameter:
//   - time int: The time to shorten the trace to
//   - incl bool: True if an element with the same time should stay included in the trace
func (this *Trace) ShortenTrace(time int, incl bool) {
	for routine, trace := range this.traces {
		for index, elem := range trace {
			if incl && elem.GetTSort() > time {
				this.traces[routine] = this.traces[routine][:index]
				break
			}
			if !incl && elem.GetTSort() >= time {
				this.traces[routine] = this.traces[routine][:index]
				break
			}
		}
	}
}

// RemoveElementFromTrace removes the element with the given tID from the trace
//
// Parameter:
//   - tID string: The tID of the element to remove
func (this *Trace) RemoveElementFromTrace(tID string) {
	for routine, trace := range this.traces {
		for index, elem := range trace {
			if elem.GetTID() == tID {
				this.traces[routine] = append(this.traces[routine][:index], this.traces[routine][index+1:]...)
				break
			}
		}
	}
}

// ShortenRoutine shorten the trace of the given routine by removing all
// elements after and equal the given time
//
// Parameter:
//   - routine int: The routine to shorten
//   - time int: The time to shorten the trace to
func (this *Trace) ShortenRoutine(routine int, time int) {
	for index, elem := range this.traces[routine] {
		if elem.GetTSort() >= time {
			this.traces[routine] = this.traces[routine][:index]
			break
		}
	}
}

// ShortenRoutineIndex shorten a given a routine to index
//
// Parameter:
//   - routine int: the routine to shorten
//   - index int: the index to which it should be shortened
//   - incl bool: if true, the value a index will remain in the routine, otherwise it will be removed
func (this *Trace) ShortenRoutineIndex(routine, index int, incl bool) {
	if incl {
		this.traces[routine] = this.traces[routine][:index+1]
	} else {
		this.traces[routine] = this.traces[routine][:index]
	}
}

// GetNoRoutines returns the number of routines
//
// Returns:
//   - int: The number of routines
func (this *Trace) GetNoRoutines() int {
	return len(this.traces)
}

// GetTraceLength returns the number of element in a given routine
//
// Parameter:
//   - routine int: the routine id
//
// Returns:
//   - int: number of elements in the routine.
func (this *Trace) GetTraceLength(routine int) int {
	return len(this.GetTraces()[routine])
}

// GetTraceLengths returns a slice containing the number of elements in the
// routines
//
// Returns:
//   - []int: number of elements in routines.
func (this *Trace) GetTraceLengths() []int {
	l := make([]int, this.GetNoRoutines()+1)
	for i, trace := range this.GetTraces() {
		l[i] = len(trace)
	}

	return l
}

// NumberElemInRoutine returns the number of elements in the trace.
//
// Parameter:
//   - routine: return the number of elements in this routine, if -1, return the number of all elements
//
// Returns:
//   - int: the number of element in a routine or the complete trace
func (this *Trace) NumberElemInRoutine(routine int) int {
	if routine == -1 {
		total := 0
		for _, n := range this.traces {
			total += len(n)
		}
		return total
	}

	return len(this.traces[routine])
}

// GetLastElemPerRout returns the last elements in each routine
// Returns
//
//   - []TraceElements: List of elements that are the last element in a routine
func (this *Trace) GetLastElemPerRout() []Element {
	res := make([]Element, 0)
	for _, trace := range this.traces {
		if len(trace) == 0 {
			continue
		}

		res = append(res, trace[len(trace)-1])
	}

	return res
}

// RemoveLater removes all elements that have a later tPost that the given tPost
//
// Parameter:
//   - tPost int: Remove elements after tPost
func (this *Trace) RemoveLater(tPost int) {
	mapping := make(map[string]Element)
	for routine, trace := range this.traces {
		newElems := make([]Element, 0)
		for _, elem := range trace {
			if elem.GetTPost() > tPost {
				newElems = append(newElems, elem.Copy(mapping, true))
			}
		}
		this.traces[routine] = newElems
	}
}

// ShiftRoutine shifts all elements in a routine with time greater or equal to
// startTSort by shift. Only shift back (shift > 0).
//
// Parameter:
//   - routine int: The routine to shift
//   - startTSort int: The time to start shifting
//   - shift int: The shift, must be > 0
//
// Returns:
//   - bool: True if the shift was successful, false otherwise (shift <= 0)
func (this *Trace) ShiftRoutine(routine int, startTSort int, shift int) bool {
	if shift <= 0 {
		return false
	}

	for index, elem := range this.traces[routine] {
		if elem.GetTPre() >= startTSort {
			this.traces[routine][index].SetTWithoutNotExecuted(elem.GetTSort() + shift)
		}
	}

	return true
}

// GetPartialTrace returns the partial trace of all element between startTime
// and endTime inclusive.
//
// Parameter:
//   - startTime int: The start time
//   - endTime int: The end time
//
// Returns:
//   - map[int][]TraceElement: The partial trace
func (this *Trace) GetPartialTrace(startTime int, endTime int) map[int][]Element {
	result := make(map[int][]Element)
	for routine, trace := range this.traces {
		for index, elem := range trace {
			if _, ok := result[routine]; !ok {
				result[routine] = make([]Element, 0)
			}
			time := elem.GetTSort()
			if time >= startTime && time <= endTime {
				result[routine] = append(result[routine], this.traces[routine][index])
			}
		}
	}

	return result
}

// Copy returns a deep copy a trace
//
// Parameter:
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - Trace: The copy of the trace
//   - error
func (this *Trace) Copy(keep bool) (Trace, error) {
	mapping := make(map[string]Element)
	tracesCopy := make(map[int][]Element)
	for routine, trace := range this.traces {
		tracesCopy[routine] = make([]Element, len(trace))
		for i, elem := range trace {
			tracesCopy[routine][i] = elem.Copy(mapping, keep)

			if control.WasCanceled() {
				return Trace{}, fmt.Errorf("Analysis was canceled due to insufficient RAM")
			}
		}
	}

	return Trace{
		traces:     tracesCopy,
		minTraceID: this.minTraceID,
	}, nil
}

// PrintTrace prints the trace sorted by tPost
func (this *Trace) PrintTrace() {
	this.PrintTraceArgs([]string{}, false)
}

// PrintTraceArgs print the elements of given types sorted by tPost
//
// Parameter:
//   - ty: types of the elements to print. If empty, all elements will be printed
//   - clocks: if true, the clocks will be printed
func (this *Trace) PrintTraceArgs(ty []string, clocks bool) {
	elements := make([]struct {
		string
		time   int
		thread int
	}, 0)
	for _, tra := range this.traces {
		for _, elem := range tra {
			elemStr := elem.ToString()
			if len(ty) == 0 || types.Contains(ty, elemStr[0:1]) {
				elements = append(elements, struct {
					string
					time   int
					thread int
				}{elemStr, elem.GetTPost(), elem.GetRoutine()})
			}
		}
	}

	// sort elements by timestamp
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].time < elements[j].time
	})

	if len(elements) == 0 {
		log.Info("Trace contains no elements")
	} else {
		log.Infof("Trace contains %d elements", len(elements))
	}

	for _, elem := range elements {
		if clocks {
			fmt.Println(elem.thread, elem.string)
		} else {
			fmt.Println(elem.thread, elem.string)
		}
	}
}

// SetTSortAtIndex sets the tSort for an element given by its index
//
// Parameter:
//   - tSort int: the new tSort
//   - routine int: the routine of the element
//   - index int: the index of the element in its routine
func (this *Trace) SetTSortAtIndex(tPost, routine, index int) {
	if len(this.traces[routine]) <= index {
		return
	}
	this.traces[routine][index].SetTSort(tPost)
}

// Iterator is an iterator to iterate over the element in the trace
// sorted by tSort
type Iterator struct {
	t            *Trace
	currentIndex map[int]int
}

// AsIterator returns a new iterator for a trace
//
// Returns:
//   - the iterator
func (this *Trace) AsIterator() Iterator {
	return Iterator{this, make(map[int]int)}
}

// GetTraceSection returns a copy of a section of the trace given by index
//
// Parameter:
//   - start int: start index
//   - end int: end index
//
// Returns:
//   - []trace.Element: the elements in the trace between start and end (including)
//     if start >= end, the result is empty. If start < 0, start is set to 0,
//     if end > len(trace), end is set to len(trace)
func (this *Trace) GetTraceSection(start, end int) []Element {
	if end <= start {
		return make([]Element, 0)
	}

	start = max(0, start)
	end = min(end, this.GetNumberElements()-1)
	numElems := end - start

	res := make([]Element, numElems)

	traceIter := this.AsIterator()

	counter := 0
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		if counter >= start {
			res = append(res, elem)
		}
		if counter >= end {
			return res
		}
		counter++
	}
	return res
}

// Next returns the next element from the iterator. If all elements have been returned
// already, return nul
//
// Returns:
//   - TraceElement: the next element, or nil if no element are left
func (this *Iterator) Next() Element {
	// find the local trace, where the element on which currentIndex points to
	// has the smallest tPost
	minTSort := -1
	minRoutine := -1
	for routine, trace := range this.t.traces {
		// no more elements in the routine trace
		if this.currentIndex[routine] == -1 {
			continue
		}

		// ignore empty routines
		if len(trace) == 0 {
			this.currentIndex[routine] = -1
			continue
		}

		// ignore non executed operations
		tSort := trace[this.currentIndex[routine]].GetTSort()
		if tSort == 0 || tSort == math.MaxInt {
			continue
		}
		if minTSort == -1 || trace[this.currentIndex[routine]].GetTSort() < minTSort {
			minTSort = trace[this.currentIndex[routine]].GetTSort()
			minRoutine = routine
		}
	}

	// all executed elements have been processed
	// check for elements with just a pre but no post
	if minRoutine == -1 {
		for routine := range this.t.traces {
			if this.currentIndex[routine] == -1 {
				continue
			}

			element := this.t.traces[routine][this.currentIndex[routine]]
			this.IncreaseIndex(routine)

			return element
		}

		// all elements have been processed
		return nil
	}

	// return the element and increase the index
	element := this.t.traces[minRoutine][this.currentIndex[minRoutine]]
	this.IncreaseIndex(minRoutine)

	return element
}

// Reset resets the iterator
func (this *Iterator) Reset() {
	this.currentIndex = make(map[int]int)
}

// IncreaseIndex the currentIndex value of a trace iterator for a routine
//
// Parameter:
//   - routine int: the routine to update
func (this *Iterator) IncreaseIndex(routine int) {
	if this.currentIndex[routine] == -1 {
		log.Error("Tried to increase index of -1 at routine ", routine)
	}
	this.currentIndex[routine]++
	if this.currentIndex[routine] >= len(this.t.traces[routine]) {
		this.currentIndex[routine] = -1
	}
}
