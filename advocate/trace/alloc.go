// Copyright (c) 2026 Erik Kassubek
//
// File: alloc.go
// Brief: Trace element to store the creation (new) of relevant operations. For now this is only creates the new for channel. This may be expanded later.
//
// Author: Erik Kassubek
// Created: 2024-11-29
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

// ElementAlloc is a trace element for the creation of an object / new
// Fields:
//   - id: id of the element, should never be changed
//   - objId int: The id of the underlying operation
//   - index int: Index in the routine
//   - routine int: The routine id
//   - elemType newOpType: The type of the created object
//   - t int: The timestamp of the new
//   - pos *position: code position
//   - ci *concInfo: concurrency info
//   - num int: variable field for additional information
//   - function *ElementFunc: the function the operation is in
//
// For now this is only creates the new for channel. This may be expanded later.
type ElementAlloc struct {
	id       int
	objId    int
	index    int
	routine  int
	elemType OperationType
	t        int
	pos      *position
	ci       *concInfo
	num      int
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementAlloc adds a make trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - t string: The timestamp at the event
//   - id string: The id of the channel
//   - elemType string: Type of the created primitive
//   - num string: Variable field for additional information
//   - pos string: position
func (this *Trace) AddTraceElementAlloc(routine int, t string, id string, elemType string, num string, pos string) error {
	tInt, err := strconv.Atoi(t)
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

	elem := ElementAlloc{
		index:    this.NumberElemInRoutine(routine),
		routine:  routine,
		t:        tInt,
		objId:    idInt,
		elemType: et,
		num:      numInt,
		pos:      newPosition(file, line),
		ci:       newConcInfo(),
		function: getLastCall(routine),
	}

	this.AddElement(&elem)

	this.allocs[elem.objId] = &elem

	return nil
}

// ========================================================
// MARK: ID
// ========================================================

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementAlloc) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementAlloc) setID(ID int) {
	this.id = ID
}

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementAlloc) GetObjId() int {
	return this.objId
}

// setObjId sets the object id
//
// Parameter:
//   - id int: the object id
func (this *ElementAlloc) setObjId(id int) {
	this.objId = id
}

// ========================================================
// MARK: Index
// ========================================================

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementAlloc) GetRoutine() int {
	return this.routine
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementAlloc) GetTraceIndex() (int, int) {
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
func (this *ElementAlloc) GetType(operation bool) OperationType {
	if !operation {
		return New
	}

	return this.elemType
}

// ========================================================
// MARK: Timestamps
// ========================================================

// GetT returns the timestamp of the element
//
// Returns:
//   - int: The tPre of the element
func (this *ElementAlloc) GetT(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementAlloc) SetT(_ timeType, tSort int) {
	this.t = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementAlloc) SetTWithoutNotExecuted(tSort int) {
	if this.t == 0 {
		return
	}
	this.t = tSort
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementAlloc) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementAlloc) GetPos() string {
	return this.pos.toString()
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - int: The file of the element
func (this *ElementAlloc) GetFile() string {
	return this.pos.file
}

// GetTraceID sets the file
//
// Parameter:
//   - f string: the file
func (this *ElementAlloc) setFile(f string) {
	this.pos.file = f
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - int: The line of the element
func (this *ElementAlloc) GetLine() int {
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
func (this *ElementAlloc) IsEqual(elem Element) bool {
	return this.id == elem.GetID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementAlloc) IsSameElement(elem Element) bool {
	return this.objId == elem.GetObjId()
}

// ========================================================
// MARK: String
// ========================================================

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementAlloc) ToString() string {
	return fmt.Sprintf("N,%d,%d,%s,%d,%s", this.t, this.objId, string(this.elemType), this.num, this.GetPos())
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementAlloc) GetFunction() *ElementFunc {
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
func (this *ElementAlloc) SetVc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementAlloc) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementAlloc) GetNumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementAlloc) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// GetReplayID returns the replay ID of the element
//
// Returns:
//   - int: The replayId of the element
func (this *ElementAlloc) GetReplayID() string {
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
func (this *ElementAlloc) Copy(mapping map[int]Element, keep bool) Element {

	return &ElementAlloc{
		id:       this.id,
		index:    0,
		routine:  this.routine,
		t:        0,
		objId:    this.objId,
		elemType: this.elemType,
		pos:      this.pos.copy(),
		ci:       this.ci.copy(),
		function: this.function.Copy(mapping, keep).(*ElementFunc),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementAlloc) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

// GetNum returns the num field of the element
//
// Returns:
//   - VectorClock: The num field of the element
func (this *ElementAlloc) GetNum() int {
	return this.num
}

// IsRequest determines if the element is a request
// Returns:
//   - bool: element is request
func (this *ElementAlloc) IsRequest() bool {
	return false
}

// IsRequest determines if the element can be a request
// Returns:
//   - bool: element can be request
func (this *ElementAlloc) CanBeRequest() bool {
	return false
}

// SetRequest set request
// Argument:
//   - bool: element is request
func (this *ElementAlloc) SetRequest(_ bool) {
	return
}

// SetRequest sets the routine id
// Argument:
//   - int: new routine id
func (this *ElementAlloc) SetRoutine(id int) {
	this.routine = id
}
