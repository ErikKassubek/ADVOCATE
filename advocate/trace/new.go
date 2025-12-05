// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementNew.go
// Brief: Trace element to store the creation (new) of relevant operations. For now this is only creates the new for channel. This may be expanded later.
//
// Author: Erik Kassubek
// Created: 2024-11-29
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"advocate/utils/types"
	"errors"
	"fmt"
	"strconv"
)

// ElementNew is a trace element for the creation of an object / new
// Fields:
//   - id: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp of the new
//   - objId int: The id of the underlying operation
//   - elemType newOpType: The type of the created object
//   - num int: Variable field for additional information
//   - file string: The file of the new
//   - line int: The line of the new
//   - children []TraceElement: children in partial order graph
//   - parents []TraceElement: parents in partial order graph
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
//
// For now this is only creates the new for channel. This may be expanded later.
type ElementNew struct {
	id                       int
	index                    int
	routine                  int
	tPost                    int
	objId                    int
	elemType                 OperationType
	num                      int
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementNew adds a make trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the channel
//   - elemType string: Type of the created primitive
//   - num string: Variable field for additional information
//   - pos string: position
func (this *Trace) AddTraceElementNew(routine int, tPost string, id string, elemType string, num string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	numInt, err := strconv.Atoi(num)
	if err != nil {
		return errors.New("num is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	et := None
	switch elemType {
	case "NA":
		et = NewAtomic
	case "NC":
		et = NewChannel
	case "ND":
		et = NewCond
	case "NM":
		et = NewMutex
	case "NO":
		et = NewOnce
	case "NW":
		et = NewWait
	}

	elem := ElementNew{
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPost:                    tPostInt,
		objId:                    idInt,
		elemType:                 et,
		num:                      numInt,
		file:                     file,
		line:                     line,
		vc:                       nil,
		wVc:                      nil,
		numberConcurrent:         -1,
		numberConcurrentWeak:     -1,
		numberConcurrentSame:     -1,
		numberConcurrentWeakSame: -1,
	}

	this.AddElement(&elem)
	return nil
}

// Get the ElemMin representation of the operation
//
// Returns:
//   - ElemMin: the ElemMin representations of the operation
//   - bool: true if it should be part of a min trace, false otherwise
func (this *ElementNew) GetElemMin() (ElemMin, bool) {
	return ElemMin{
		ID:      this.id,
		ObjID:   this.objId,
		Op:      this.elemType,
		Pos:     PosStringFromPos(this.file, this.line),
		Time:    types.NewPair(this.tPost, this.tPost),
		Routine: this.routine,
		Vc:      *this.vc.Copy(),
	}, false
}

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementNew) GetObjId() int {
	return this.objId
}

// GetTPre returns the tPre of the element
//
// Returns:
//   - int: The tPre of the element
func (this *ElementNew) GetTPre() int {
	return this.tPost
}

// GetTPost returns the tPost of the operation.
//
// Returns:
//   - string: The position of the element
func (this *ElementNew) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - float32: The time of the element
func (this *ElementNew) GetTSort() int {
	return this.tPost
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementNew) GetRoutine() int {
	return this.routine
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementNew) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay ID of the element
//
// Returns:
//   - int: The replayId of the element
func (this *ElementNew) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - int: The file of the element
func (this *ElementNew) GetFile() string {
	return this.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - int: The line of the element
func (this *ElementNew) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form "N@[file]:[line]@[tPre]
//
// Returns:
//   - int: The tID of the element
func (this *ElementNew) GetTID() string {
	return "N@" + this.GetPos() + "@" + strconv.Itoa(this.tPost)
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementNew) GetType(operation bool) OperationType {
	if !operation {
		return New
	}

	return this.elemType
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementNew) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementNew) SetWVc(vc *clock.VectorClock) {
	this.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementNew) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementNew) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetNum returns the num field of the element
//
// Returns:
//   - VectorClock: The num field of the element
func (this *ElementNew) GetNum() int {
	return this.num
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementNew) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementNew) ToString() string {
	return fmt.Sprintf("N,%d,%d,%s,%d,%s", this.tPost, this.objId, string(this.elemType), this.num, this.GetPos())
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementNew) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementNew) IsSameElement(elem Element) bool {
	return false
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementNew) SetTPre(tSort int) {
	this.tPost = tSort
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementNew) SetT(tSort int) {
	this.tPost = tSort
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementNew) SetTSort(tSort int) {
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementNew) SetTWithoutNotExecuted(tSort int) {
	if this.tPost == 0 {
		return
	}
	this.tPost = tSort
}

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementNew) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementNew) setID(ID int) {
	this.id = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since New do not contain reference to other elements and no other
//     elements contain referents to New, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementNew) Copy(_ map[string]Element) Element {

	return &ElementNew{
		id:                       this.id,
		index:                    this.index,
		routine:                  this.routine,
		tPost:                    this.tPost,
		objId:                    this.objId,
		elemType:                 this.elemType,
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
func (this *ElementNew) GetNumberConcurrent(weak, sameElem bool) int {
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
func (this *ElementNew) SetNumberConcurrent(c int, weak, sameElem bool) {
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
