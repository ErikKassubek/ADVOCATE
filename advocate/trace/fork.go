// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementFork.go
// Brief: Struct and functions for fork operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"errors"
	"fmt"
	"strconv"
)

// ElementFork is a trace element for a go statement
// Fields:
//   - traceID: id of the element, should never be changed
//   - index int: the index of the fork in the routine
//   - routine int: The routine id of
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the new go routine
//   - file (string), line int: The position of the trace element in the file
//   - vc *clock.VectorClock: the vector clock of the element
//   - wVc *clock.VectorClock: the weak vector clock of the element
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementFork struct {
	traceID                  int
	index                    int
	routine                  int
	tPost                    int
	id                       int
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	concurrent               []Element
	concurrentWeak           []Element
	numberConcurrentSame     int
	numberConcurrentWeakSame int
	concurrentSame           []Element
	concurrentWeakSame       []Element
}

// AddTraceElementFork adds a new go statement element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the new routine
//   - pos string: The position of the trace element in the file
func (t *Trace) AddTraceElementFork(routine int, tPost string, id string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementFork{
		index:                    t.numberElemsInTrace[routine],
		routine:                  routine,
		tPost:                    tPostInt,
		id:                       idInt,
		file:                     file,
		line:                     line,
		vc:                       nil,
		wVc:                      nil,
		numberConcurrent:         -1,
		numberConcurrentWeak:     -1,
		concurrent:               make([]Element, 0),
		concurrentWeak:           make([]Element, 0),
		numberConcurrentSame:     -1,
		numberConcurrentWeakSame: -1,
		concurrentSame:           make([]Element, 0),
		concurrentWeakSame:       make([]Element, 0),
	}

	t.AddElement(&elem)
	return nil
}

// GetID returns the ID of the newly created routine
//
// Returns:
//   - int: The id of the new routine
func (fo *ElementFork) GetID() int {
	return fo.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (fo *ElementFork) GetRoutine() int {
	return fo.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (fo *ElementFork) GetTPre() int {
	return fo.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (fo *ElementFork) GetTPost() int {
	return fo.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (fo *ElementFork) GetTSort() int {
	return fo.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (fo *ElementFork) GetPos() string {
	return fmt.Sprintf("%s:%d", fo.file, fo.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (fo *ElementFork) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", fo.routine, fo.file, fo.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (fo *ElementFork) GetFile() string {
	return fo.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (fo *ElementFork) GetLine() int {
	return fo.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (fo *ElementFork) GetTID() string {
	return fo.GetPos() + "@" + strconv.Itoa(fo.tPost)
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (fo *ElementFork) SetVc(vc *clock.VectorClock) {
	fo.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (fo *ElementFork) SetWVc(vc *clock.VectorClock) {
	fo.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (fo *ElementFork) GetVC() *clock.VectorClock {
	return fo.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (fo *ElementFork) GetWVC() *clock.VectorClock {
	return fo.wVc
}

// GetObjType returns the string representation of the object type
func (fo *ElementFork) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeFork + "F"
	}
	return ObjectTypeFork
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (fo *ElementFork) IsEqual(elem Element) bool {
	return fo.routine == elem.GetRoutine() && fo.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (fo *ElementFork) GetTraceIndex() (int, int) {
	return fo.routine, fo.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (fo *ElementFork) SetT(time int) {
	fo.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (fo *ElementFork) SetTPre(tPre int) {
	fo.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (fo *ElementFork) SetTSort(tPost int) {
	fo.SetTPre(tPost)
	fo.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (fo *ElementFork) SetTWithoutNotExecuted(tSort int) {
	fo.SetTPre(tSort)
	if fo.tPost != 0 {
		fo.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (fo *ElementFork) ToString() string {
	return "G" + "," + strconv.Itoa(fo.tPost) + "," + strconv.Itoa(fo.id) +
		"," + fo.GetPos()
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (fo *ElementFork) GetTraceID() int {
	return fo.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (fo *ElementFork) setTraceID(ID int) {
	fo.traceID = ID
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (fo *ElementFork) Copy() Element {
	copyConcurrent := make([]Element, 0)
	copy(copyConcurrent, fo.concurrent)
	copyConcurrentWeak := make([]Element, 0)
	copy(copyConcurrentWeak, fo.concurrentWeak)
	copyConcurrentSame := make([]Element, 0)
	copy(copyConcurrentSame, fo.concurrentSame)
	copyConcurrentWeakSame := make([]Element, 0)
	copy(copyConcurrentWeakSame, fo.concurrentWeakSame)

	return &ElementFork{
		traceID:                  fo.traceID,
		index:                    fo.index,
		routine:                  fo.routine,
		tPost:                    fo.tPost,
		id:                       fo.id,
		file:                     fo.file,
		line:                     fo.line,
		vc:                       fo.vc.Copy(),
		wVc:                      fo.wVc.Copy(),
		numberConcurrent:         fo.numberConcurrent,
		numberConcurrentWeak:     fo.numberConcurrentWeak,
		concurrent:               copyConcurrent,
		concurrentWeak:           copyConcurrentWeak,
		numberConcurrentSame:     fo.numberConcurrentSame,
		numberConcurrentWeakSame: fo.numberConcurrentWeakSame,
		concurrentSame:           copyConcurrentSame,
		concurrentWeakSame:       copyConcurrentWeakSame,
	}
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Parameter:
//   - weak bool: get number of weak concurrent
//   - sameElem bool: only operation on the same variable
//
// Returns:
//   - number of concurrent element, or -1
func (fo *ElementFork) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return fo.numberConcurrentWeakSame
		}
		return fo.numberConcurrentWeak
	}
	if sameElem {
		return fo.numberConcurrentSame
	}
	return fo.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (fo *ElementFork) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			fo.numberConcurrentWeakSame = c
		} else {
			fo.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			fo.numberConcurrentSame = c
		} else {
			fo.numberConcurrent = c
		}
	}
}

// GetConcurrent returns the elements that are concurrent to the element
//
// Parameter:
//   - weak bool: get number of weak concurrent
//   - sameElem bool: only operation on the same variable
//
// Returns:
//   - []Element: the concurrent elements
func (fo *ElementFork) GetConcurrent(weak, sameElem bool) []Element {
	if weak {
		if sameElem {
			return fo.concurrentWeakSame
		}
		return fo.concurrentWeak
	}
	if sameElem {
		return fo.concurrentSame
	}
	return fo.concurrent
}

// SetConcurrent sets the concurrent elements
//
// Parameter:
//   - []Element: the concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (fo *ElementFork) SetConcurrent(elem []Element, weak, sameElem bool) {
	fo.SetNumberConcurrent(len(elem), weak, sameElem)

	if weak {
		if sameElem {
			fo.concurrentWeakSame = elem
		} else {
			fo.concurrentWeak = elem
		}
	} else {
		if sameElem {
			fo.concurrentSame = elem
		} else {
			fo.concurrent = elem
		}
	}
}
