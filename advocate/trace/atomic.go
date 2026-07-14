// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/atomic.go
// Brief: Struct and functions for atomic operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"advocate/utils/log"
	"errors"
	"fmt"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementAtomic is a struct to save an atomic event in the trace
// Fields:
//
//   - id: id of the element, should never be changed
//   - objId int: The id of the atomic variable
//   - index int: index in the routine
//   - routine int: The routine id
//   - op ObjectType: The operation on the atomic variable
//   - t int: The timestamp of the event
//   - pos *position: code position
//   - ci *concInfo: concurrency info
//   - function *ElementFunc: the function the operation is in
type ElementAtomic struct {
	id       int
	objId    int
	index    int
	routine  int
	op       OperationType
	t        int
	pos      *position
	ci       *concInfo
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

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
		index:    this.NumberElemInRoutine(routine),
		routine:  routine,
		t:        tPostInt,
		objId:    idInt,
		op:       opAInt,
		pos:      newPosition(file, line),
		ci:       newConcInfo(),
		function: getLastCall(routine),
	}

	this.AddElement(&elem)
	return nil
}

// ========================================================
// MARK: ID
// ========================================================

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

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementAtomic) GetObjId() int {
	return this.objId
}

// ========================================================
// MARK: Index
// ========================================================

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementAtomic) GetRoutine() int {
	return this.routine
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementAtomic) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// MARK: Operation
// ========================================================

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

// ========================================================
// MARK: Timestamps
// ========================================================

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (this *ElementAtomic) GetT(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementAtomic) SetT(_ timeType, time int) {
	this.t = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementAtomic) SetTWithoutNotExecuted(tSort int) {
	if this.t != 0 {
		this.t = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementAtomic) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (this *ElementAtomic) GetPos() string {
	return this.pos.toString()
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementAtomic) GetFile() string {
	return this.pos.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementAtomic) GetLine() int {
	return this.pos.line
}

// ========================================================
// MARK: Equal
// ========================================================

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementAtomic) IsEqual(elem Element) bool {
	return this.id == elem.GetID()
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
	return this.objId == elem.GetObjId()
}

// ========================================================
// MARK: String
// ========================================================

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementAtomic) ToString() string {
	opString := string(string(this.op)[1])

	return fmt.Sprintf("A,%d,%d,%s,%s", this.t, this.objId, opString, this.GetPos())
}

// GetTID returns the tID of the element.
// The tID is a string of form A@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementAtomic) GetTID() string {
	return "A@" + this.GetPos() + "@" + strconv.Itoa(this.t)
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementAtomic) GetFunction() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

// SetVc sets the vector clock
//
// Parameter:
//   - weak bool: set the weak wv
//   - cl *clock.VectorClock: the vector clock
func (this *ElementAtomic) SetVc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementAtomic) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
	return this.ci.getVC(weak)
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
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementAtomic) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementAtomic) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.pos.file, this.pos.line)
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy the atomic element
//
// Parameter:
//   - mapping map[int]Element: map containing all already copied elements, if nil ignore all vc based values.
//     since atomics do not contain reference to other elements and no other
//     elements contain referents to atomics, this is not used
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementAtomic) Copy(mapping map[int]Element, keep bool) Element {

	if !keep {
		return &ElementAtomic{
			id:       this.id,
			index:    0,
			routine:  this.routine,
			t:        0,
			objId:    this.objId,
			op:       this.op,
			pos:      this.pos.copy(),
			ci:       newConcInfo(),
			function: this.function.Copy(mapping, keep).(*ElementFunc),
		}
	}

	return &ElementAtomic{
		id:       this.id,
		index:    this.index,
		routine:  this.routine,
		t:        this.t,
		objId:    this.objId,
		op:       this.op,
		pos:      this.pos.copy(),
		ci:       this.ci.copy(),
		function: this.function.Copy(mapping, keep).(*ElementFunc),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementAtomic) IsValid() bool {
	return this != nil
}
