// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/fork.go
// Brief: Struct and functions for fork operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"errors"
	"fmt"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementFork is a trace element for a go statement
// Fields:
//   - id: id of the element, should never be changed
//   - index int: the index of the fork in the routine
//   - routine int: The routine id of
//   - t int: The timestamp at the end of the event
//   - objId int: The id of the new go routine
//   - pos *position: code position
//   - ci *concInfo: concurrency info
//   - function *ElementFunc: the function the operation is in
type ElementFork struct {
	id       int
	index    int
	routine  int
	t        int
	objId    int
	pos      *position
	ci       *concInfo
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

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
		index:    this.NumberElemInRoutine(routine),
		routine:  routine,
		t:        tPostInt,
		objId:    idInt,
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

// GetObjId returns the ID of the newly created routine
//
// Returns:
//   - int: The id of the new routine
func (this *ElementFork) GetObjId() int {
	return this.objId
}

// ========================================================
// MARK: Timestamps
// ========================================================

// GetT returns the t of the element.
//
// Returns:
//   - int: The tPre of the element
func (this *ElementFork) GetT(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementFork) SetT(_ timeType, time int) {
	this.t = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementFork) SetTWithoutNotExecuted(tSort int) {
	if this.t != 0 {
		this.t = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementFork) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementFork) GetPos() string {
	return this.pos.toString()
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementFork) GetFile() string {
	return this.pos.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementFork) GetLine() int {
	return this.pos.line
}

// ========================================================
// MARK: Index
// ========================================================

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementFork) GetRoutine() int {
	return this.routine
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementFork) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// MARK: Operation
// ========================================================

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
func (this *ElementFork) IsEqual(elem Element) bool {
	return this.id == elem.GetID()
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

// ========================================================
// MARK: String
// ========================================================

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementFork) ToString() string {
	return "G" + "," + strconv.Itoa(this.t) + "," + strconv.Itoa(this.objId) +
		"," + this.GetPos()
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementFork) GetFunction() *ElementFunc {
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
func (this *ElementFork) SetVc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementFork) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementFork) GetNumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementFork) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementFork) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.pos.file, this.pos.line)
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy the element
//
// Parameter:
//   - mapping map[string]Element: map containing all already copied elements.
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementFork) Copy(mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementFork{
			id:       this.id,
			index:    0,
			routine:  this.routine,
			t:        0,
			objId:    this.objId,
			pos:      this.pos.copy(),
			ci:       newConcInfo(),
			function: this.function.Copy(mapping, keep).(*ElementFunc),
		}
	}

	return &ElementFork{
		id:       this.id,
		index:    this.index,
		routine:  this.routine,
		t:        this.t,
		objId:    this.objId,
		pos:      this.pos.copy(),
		ci:       this.ci.copy(),
		function: this.function.Copy(mapping, keep).(*ElementFunc),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementFork) IsValid() bool {
	return this != nil
}
