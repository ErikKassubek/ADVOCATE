// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementMutex.go
// Brief: Struct and functions for mutex operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"advocate/analysis/hb/clock"
	"advocate/utils/types"
)

// ElementMutex is a trace element for a mutex
//
// Fields:
//   - id: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - objId int: The id of the mutex
//   - rw bool: Whether the mutex is a read-write mutex
//   - op ObjectType: The operation on the mutex
//   - suc bool: Whether the operation was successful (only for trylock else always true)
//   - file string: The file of the mutex operation in the code
//   - line int: The line of the mutex operation in the code
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementMutex struct {
	id                       int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	objId                    int
	rw                       bool
	op                       OperationType
	suc                      bool
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementMutex adds a new mutex element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the mutex
//   - rw string: Whether the mutex is a read-write mutex
//   - opM string: The operation on the mutex
//   - suc string: Whether the operation was successful (only for trylock else always true)
//   - pos string: The position of the mutex operation in the code
func (this *Trace) AddTraceElementMutex(routine int, tPre string,
	tPost string, id string, rw string, opM string, suc string,
	pos string) error {
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

	rwBool := false
	if rw == "t" {
		rwBool = true
	}

	var opMInt OperationType
	switch opM {
	case "L":
		opMInt = MutexLock
	case "R":
		opMInt = MutexRLock
	case "T":
		opMInt = MutexTryLock
	case "Y":
		opMInt = MutexTryRLock
	case "U":
		opMInt = MutexUnlock
	case "N":
		opMInt = MutexRUnlock
	default:
		return errors.New("opM is not a valid operation")
	}

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementMutex{
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		objId:                    idInt,
		rw:                       rwBool,
		op:                       opMInt,
		suc:                      sucBool,
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
func (this *ElementMutex) GetElemMin() (ElemMin, bool) {
	return ElemMin{
		ID:      this.id,
		ObjID:   this.objId,
		Op:      this.op,
		Pos:     PosStringFromPos(this.file, this.line),
		Time:    types.NewPair(this.tPre, this.tPost),
		Routine: this.routine,
		Vc:      *this.vc.Copy(),
	}, true
}

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementMutex) GetObjId() int {
	return this.objId
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementMutex) GetRoutine() int {
	return this.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (this *ElementMutex) GetTPre() int {
	return this.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (this *ElementMutex) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementMutex) GetTSort() int {
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
func (this *ElementMutex) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementMutex) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementMutex) GetFile() string {
	return this.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementMutex) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form "M@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementMutex) GetTID() string {
	return "M@" + this.GetPos() + "@" + strconv.Itoa(this.tPre)
}

// IsLock returns if the element is a lock operation
//
// Returns:
//   - bool: If the element is a lock operation
func (this *ElementMutex) IsLock() bool {
	return this.op == MutexLock || this.op == MutexRLock || this.op == MutexTryLock || this.op == MutexTryRLock
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementMutex) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementMutex) SetWVc(vc *clock.VectorClock) {
	this.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementMutex) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementMutex) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementMutex) GetType(operation bool) OperationType {
	if !operation {
		return Mutex
	}

	return this.op
}

// IsSuc returns whether the locking was successful of the element
//
// Returns:
//   - For trylock wether it was successful, otherwise always true
func (this *ElementMutex) IsSuc() bool {
	return this.suc
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementMutex) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same mutex
func (this *ElementMutex) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Mutex {
		return false
	}

	return this.objId == elem.GetObjId()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementMutex) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementMutex) SetT(time int) {
	this.tPre = time
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementMutex) SetTPre(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementMutex) SetTSort(tSort int) {
	this.SetTPre(tSort)
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementMutex) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementMutex) ToString() string {
	res := "M,"
	res += strconv.Itoa(this.tPre) + "," + strconv.Itoa(this.tPost) + ","
	res += strconv.Itoa(this.objId) + ","

	if this.rw {
		res += "R,"
	} else {
		res += "-,"
	}

	res += string(string(this.op)[1])

	if this.suc {
		res += ",t"
	} else {
		res += ",f"
	}
	res += "," + this.GetPos()
	return res
}

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementMutex) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementMutex) setID(ID int) {
	this.id = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since mutex do not contain reference to other elements and no other
//     elements contain referents to mutex, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementMutex) Copy(_ map[string]Element) Element {
	return &ElementMutex{
		id:                       this.id,
		index:                    this.index,
		routine:                  this.routine,
		tPre:                     this.tPre,
		tPost:                    this.tPost,
		objId:                    this.objId,
		rw:                       this.rw,
		op:                       this.op,
		suc:                      this.suc,
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
func (this *ElementMutex) GetNumberConcurrent(weak, sameElem bool) int {
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
func (this *ElementMutex) SetNumberConcurrent(c int, weak, sameElem bool) {
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
