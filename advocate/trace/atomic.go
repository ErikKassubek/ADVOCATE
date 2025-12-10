// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementAtomic.go
// Brief: Struct and functions for atomic operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"advocate/utils/log"
	"errors"
	"fmt"
	"strconv"
)

// ElementAtomic is a struct to save an atomic event in the trace
// Fields:
//
//   - id: id of the element, should never be changed
//   - index int: index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp of the event
//   - objId int: The id of the atomic variable
//   - op ObjectType: The operation on the atomic variable
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//   - file string: the file of the operation
//   - line int: the line of the operation
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementAtomic struct {
	id                       int
	index                    int
	routine                  int
	tPost                    int
	objId                    int
	op                       OperationType
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	file                     string
	line                     int
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementAtomic adds a new atomic trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp of the event
//   - id string: The id of the atomic variable
//   - operation string: The operation on the atomic variable
//   - pos string: The position of the atomic
func (this Trace) AddTraceElementAtomic(routine int, tPost string,
	id string, operation string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	var opAInt OperationType
	switch operation {
	case "L":
		opAInt = AtomicLoad
	case "S":
		opAInt = AtomicStore
	case "A":
		opAInt = AtomicAdd
	case "W":
		opAInt = AtomicSwap
	case "C":
		opAInt = AtomicCompAndSwap
	case "N":
		opAInt = AtomicAnd
	case "O":
		opAInt = AtomicOr
	default:
		return fmt.Errorf("Atomic operation '%s' is not a valid operation", operation)
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		log.Error("Cannot read pos string ", pos)
		return err
	}

	elem := ElementAtomic{
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPost:                    tPostInt,
		objId:                    idInt,
		op:                       opAInt,
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

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementAtomic) GetObjId() int {
	return this.objId
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementAtomic) GetRoutine() int {
	return this.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (this *ElementAtomic) GetTPre() int {
	return this.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (this *ElementAtomic) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementAtomic) GetTSort() int {
	return this.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (this *ElementAtomic) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementAtomic) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementAtomic) GetFile() string {
	return this.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementAtomic) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form A@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementAtomic) GetTID() string {
	return "A@" + this.GetPos() + "@" + strconv.Itoa(this.tPost)
}

// SetVc sets the vector clock
//
// Parameter:
//   - cl *clock.VectorClock: the vector clock
func (this *ElementAtomic) SetVc(cl *clock.VectorClock) {
	this.vc = cl.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - cl *clock.VectorClock: the vector clock
func (this *ElementAtomic) SetWVc(cl *clock.VectorClock) {
	this.wVc = cl.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementAtomic) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The weak vector clock of the element
func (this *ElementAtomic) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementAtomic) GetType(operation bool) OperationType {
	if !operation {
		return Atomic
	}

	return this.op
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementAtomic) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same atomic variable
func (this *ElementAtomic) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Atomic {
		return false
	}

	return this.objId == elem.GetObjId()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementAtomic) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementAtomic) SetT(time int) {
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPost of the element
func (this *ElementAtomic) SetTPre(tPre int) {
	this.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementAtomic) SetTSort(tSort int) {
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementAtomic) SetTWithoutNotExecuted(tSort int) {
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementAtomic) ToString() string {
	opString := string(string(this.op)[1])

	return fmt.Sprintf("A,%d,%d,%s,%s", this.tPost, this.objId, opString, this.GetPos())
}

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementAtomic) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementAtomic) setID(ID int) {
	this.id = ID
}

// Copy the atomic element
//
// Parameter:
//   - c map[string]Element: map containing all already copied elements, if nil ignore all vc based values.
//     since atomics do not contain reference to other elements and no other
//     elements contain referents to atomics, this is not used
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementAtomic) Copy(_ map[string]Element, keep bool) Element {

	if !keep {
		return &ElementAtomic{
			id:                       this.id,
			index:                    0,
			routine:                  this.routine,
			tPost:                    0,
			objId:                    this.objId,
			op:                       this.op,
			vc:                       nil,
			wVc:                      nil,
			numberConcurrent:         0,
			numberConcurrentWeak:     0,
			numberConcurrentSame:     0,
			numberConcurrentWeakSame: 0,
			file:                     this.file,
			line:                     this.line,
		}
	}

	return &ElementAtomic{
		id:                       this.id,
		index:                    this.index,
		routine:                  this.routine,
		tPost:                    this.tPost,
		objId:                    this.objId,
		op:                       this.op,
		vc:                       this.vc.Copy(),
		wVc:                      this.wVc.Copy(),
		numberConcurrent:         this.numberConcurrent,
		numberConcurrentWeak:     this.numberConcurrentWeak,
		numberConcurrentSame:     this.numberConcurrentSame,
		numberConcurrentWeakSame: this.numberConcurrentWeakSame,
		file:                     this.file,
		line:                     this.line,
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
func (this *ElementAtomic) GetNumberConcurrent(weak, sameElem bool) int {
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
func (this *ElementAtomic) SetNumberConcurrent(c int, weak, sameElem bool) {
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
