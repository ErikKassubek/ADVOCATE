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
//   - id: id of the element, should never be changed
//   - index int: the index of the fork in the routine
//   - routine int: The routine id of
//   - tPost int: The timestamp at the end of the event
//   - objId int: The id of the new go routine
//   - file (string), line int: The position of the trace element in the file
//   - vc *clock.VectorClock: the vector clock of the element
//   - wVc *clock.VectorClock: the weak vector clock of the element
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementFork struct {
	id                       int
	index                    int
	routine                  int
	tPost                    int
	objId                    int
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
func (this *Trace) AddTraceElementFork(routine int, tPost string, id string, pos string) error {
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
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPost:                    tPostInt,
		objId:                    idInt,
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

	this.AddElement(&elem)
	return nil
}

// GetObjId returns the ID of the newly created routine
//
// Returns:
//   - int: The id of the new routine
func (this *ElementFork) GetObjId() int {
	return this.objId
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementFork) GetRoutine() int {
	return this.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (this *ElementFork) GetTPre() int {
	return this.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (this *ElementFork) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementFork) GetTSort() int {
	return this.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementFork) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementFork) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementFork) GetFile() string {
	return this.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementFork) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form F@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementFork) GetTID() string {
	return "F@" + this.GetPos() + "@" + strconv.Itoa(this.tPost)
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementFork) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementFork) SetWVc(vc *clock.VectorClock) {
	this.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementFork) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementFork) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetObjType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementFork) GetType(operation bool) OperationType {
	if !operation {
		return Fork
	}
	return ForkOp
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementFork) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same. For fork, all forks are  considered
// to be on the same element
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same channel
func (this *ElementFork) IsSameElement(elem Element) bool {
	return elem.GetType(false) == Fork
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementFork) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementFork) SetT(time int) {
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementFork) SetTPre(tPre int) {
	this.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementFork) SetTSort(tPost int) {
	this.SetTPre(tPost)
	this.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementFork) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementFork) ToString() string {
	return "G" + "," + strconv.Itoa(this.tPost) + "," + strconv.Itoa(this.objId) +
		"," + this.GetPos()
}

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementFork) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementFork) setID(ID int) {
	this.id = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since forks do not contain reference to other elements and no other
//     elements contain referents to forks, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementFork) Copy(_ map[string]Element) Element {

	return &ElementFork{
		id:                       this.id,
		index:                    this.index,
		routine:                  this.routine,
		tPost:                    this.tPost,
		objId:                    this.objId,
		file:                     this.file,
		line:                     this.line,
		vc:                       this.vc.Copy(),
		wVc:                      this.wVc.Copy(),
		numberConcurrent:         this.numberConcurrent,
		numberConcurrentWeak:     this.numberConcurrentWeak,
		numberConcurrentSame:     this.numberConcurrentSame,
		numberConcurrentWeakSame: this.numberConcurrentWeakSame,
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
func (this *ElementFork) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return this.numberConcurrentWeakSame
		}
		return this.numberConcurrentWeak
	}
	if sameElem {
		return this.numberConcurrentSame
	}
	return this.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementFork) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			this.numberConcurrentWeakSame = c
		} else {
			this.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			this.numberConcurrentSame = c
		} else {
			this.numberConcurrent = c
		}
	}
}
