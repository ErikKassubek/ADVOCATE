// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementWait.go
// Brief: Struct and functions for wait group operations in the trace
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
	"math"
	"strconv"
)

// OpWait enum
type OpWait int

// ElementWait is a trace element for a wait group statement
//
// Fields:
//   - traceID: id of the element, should never be changed
//   - index int: Index in the routine
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the wait group
//   - opW opW: The operation on the wait group
//   - delta int: The delta of the wait group
//   - val int: The value of the wait group
//   - file string: The file of the wait group in the code
//   - line int: The line of the wait group in the code
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementWait struct {
	traceID                  int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	id                       int
	op                       OperationType
	delta                    int
	val                      int
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementWait adds a new wait group element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the wait group
//   - opW string: The operation on the wait group
//   - delta string: The delta of the wait group
//   - val string: The value of the wait group
//   - pos string: The position of the wait group in the code
func (this *Trace) AddTraceElementWait(routine int, tPre,
	tPost, id, opW, delta, val, pos string) error {
	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	deltaInt, err := strconv.Atoi(delta)
	if err != nil {
		return errors.New("delta is not an integer")
	}

	opWOp := None
	if opW == "W" {
		opWOp = WaitWait
	} else if deltaInt > 0 {
		opWOp = WaitAdd
	} else {
		opWOp = WaitDone
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("val is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementWait{
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		id:                       idInt,
		op:                       opWOp,
		delta:                    deltaInt,
		val:                      valInt,
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
func (this *ElementWait) GetElemMin() (ElemMin, bool) {
	return ElemMin{
		ID:      this.id,
		Op:      this.op,
		Pos:     PosStringFromPos(this.file, this.line),
		Routine: this.routine,
		Vc:      *this.vc.Copy(),
	}, true
}

// Return an empty wait element with an id. Mainly used for source/drain in
// st-graph to detect potential negative wait group
//
// Parameter:
//   - id int: the id of the element
//
// Returns:
//   - ElementWait: the wait element
func EmptyWait(id int) ElementWait {
	return ElementWait{
		id: id,
	}
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementWait) GetID() int {
	return this.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementWait) GetRoutine() int {
	return this.routine
}

// GetTPre returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the start of the event
func (this *ElementWait) GetTPre() int {
	return this.tPre
}

// GetTPost returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the end of the event
func (this *ElementWait) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementWait) GetTSort() int {
	if this.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return this.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementWait) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementWait) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementWait) GetFile() string {
	return this.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementWait) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementWait) GetTID() string {
	return "W@" + this.GetPos() + "@" + strconv.Itoa(this.tPre)
}

// IsWait returns if the operation is a wait op
//
// Returns:
//   - bool: True if the operation is a wait op
func (this *ElementWait) IsWait() bool {
	return this.op == WaitWait
}

// GetOpW returns the operation type
//
// Returns:
//   - objectType: the wait operations
func (this *ElementWait) GetOpW() OperationType {
	return this.op
}

// GetDelta returns the delta of the element. The delta is the value by which the counter
// of the wait has been changed. For Add the delta is > 0, for Done it is -1,
// for Wait it is 0
//
// Returns:
//   - int: the delta of the wait element
func (this *ElementWait) GetDelta() int {
	return this.delta
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementWait) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementWait) SetWVc(vc *clock.VectorClock) {
	this.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementWait) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementWait) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementWait) GetType(operation bool) OperationType {
	if !operation {
		return Wait
	}

	if this.delta > 0 {
		return WaitAdd
	} else if this.delta < 0 {
		return WaitDone
	}
	return WaitWait
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementWait) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same w3ait group
func (this *ElementWait) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Wait {
		return false
	}

	return this.id == elem.GetID()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementWait) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementWait) SetT(time int) {
	this.tPre = time
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementWait) SetTPre(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementWait) SetTSort(tSort int) {
	this.SetTPre(tSort)
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementWait) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementWait) ToString() string {
	res := "W,"
	res += strconv.Itoa(this.tPre) + "," + strconv.Itoa(this.tPost) + ","
	res += strconv.Itoa(this.id) + ","
	switch this.op {
	case WaitAdd, WaitDone:
		res += "A,"
	case WaitWait:
		res += "W,"
	}

	res += strconv.Itoa(this.delta) + "," + strconv.Itoa(this.val)
	res += "," + this.GetPos()
	return res
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementWait) GetTraceID() int {
	return this.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementWait) setTraceID(ID int) {
	this.traceID = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since atomics do not contain reference to other elements and no other
//     elements contain referents to atomics, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementWait) Copy(_ map[string]Element) Element {
	return &ElementWait{
		traceID:                  this.traceID,
		index:                    this.index,
		routine:                  this.routine,
		tPre:                     this.tPre,
		tPost:                    this.tPost,
		id:                       this.id,
		op:                       this.op,
		delta:                    this.delta,
		val:                      this.val,
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
func (this *ElementWait) GetNumberConcurrent(weak, sameElem bool) int {
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
func (this *ElementWait) SetNumberConcurrent(c int, weak, sameElem bool) {
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
