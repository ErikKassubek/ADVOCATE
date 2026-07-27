// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/cond.go
// Brief: Struct and functions for operations of conditional variables in the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementCond is a trace element for a condition variable
// Fields:
//   - id: id of the element, should never be changed
//   - index int: index in the routine
//   - routine int: The routine id
//   - tReq int: The timestamp at the start of the event
//   - tCom int: The timestamp at the end of the event
//   - objId int: The id of the condition variable
//   - op objectType: The operation on the condition variable
//   - pos position: code position
//   - ci *concInfo: concurrency infocalculated
//   - function *ElementFunc: the function the operation is in
type ElementCond struct {
	id       int
	index    int
	routine  int
	tReq     int
	tCom     int
	objId    int
	op       OperationType
	pos      position
	ci       *concInfo
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementCond adds a new condition variable element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the condition variable
//   - opC string: The operation on the condition variable
//   - pos string: The position of the condition variable operation in the code
func (this *Trace) AddTraceElementCond(routine int, tPre string, tPost string, id string, opN string, pos string) error {
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
	var op OperationType
	switch opN {
	case "W":
		op = CondWait
	case "S":
		op = CondSignal
	case "B":
		op = CondBroadcast
	default:
		return errors.New("op is not a valid operation")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementCond{
		index:    this.NumberElemInRoutine(routine),
		routine:  routine,
		tReq:     tPreInt,
		tCom:     tPostInt,
		objId:    idInt,
		op:       op,
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

// ID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementCond) ID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementCond) setID(ID int) {
	this.id = ID
}

// ObjID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementCond) ObjID() int {
	return this.objId
}

// ========================================================
// MARK: Timestamps
// ========================================================

// T returns the t of the element
//
// Parameter:
//   - t timeType: timer type
//
// Returns:
//   - int: The tPre of the element
func (this *ElementCond) T(t timeType) int {
	switch t {
	case Request:
		return this.tReq
	case Commit:
		return this.tCom
	case Sorting:
		t := this.tReq
		if this.op == CondWait {
			t = this.tCom
		}
		if t == 0 {
			// add at the end of the trace
			return math.MaxInt
		}
		return t
	}

	return this.tCom
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - t timeType: type of time to set
//   - time int: The tPre and tPost of the element
func (this *ElementCond) SetT(t timeType, time int) {
	switch t {
	case Request:
		this.tReq = time
		if this.tCom != 0 && this.tCom < time {
			this.tCom = time
		}
	case Commit, Sorting:
		this.tCom = time
		if time != 0 && this.tReq > time {
			this.tReq = time
		}
	case Both:
		this.SetT(Request, time)
		this.SetT(Commit, time)
	}
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementCond) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.op == CondWait {
		if this.tCom != 0 {
			this.tCom = tSort
		}
		return
	}
	if this.tReq != 0 {
		this.tReq = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementCond) Committed() bool {
	return this.tCom != 0
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementCond) Pos() string {
	return this.pos.toString()
}

// File returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementCond) File() string {
	return this.pos.file
}

// Line returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementCond) Line() int {
	return this.pos.line
}

// ========================================================
// MARK: Index
// ========================================================

// Routine returns the routine ID of the element.
//
// Returns:
//   - int: The routine id
func (this *ElementCond) Routine() int {
	return this.routine
}

// TraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementCond) TraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// MARK: Operation
// ========================================================

// Type returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementCond) Type(operation bool) OperationType {
	if !operation {
		return Cond
	}

	return this.op
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
func (this *ElementCond) IsEqual(elem Element) bool {
	return this.objId == elem.ObjID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same conditional variable
func (this *ElementCond) IsSameElement(elem Element) bool {
	if elem.Type(false) != Cond {
		return false
	}

	return this.objId == elem.ObjID()
}

// ========================================================
// MARK: String
// ========================================================

// String returns the string representation of the element
//
// Returns:
//   - string: The string representation of the element
func (this *ElementCond) String() string {
	res := "D,"
	res += strconv.Itoa(this.tReq) + "," + strconv.Itoa(this.tCom) + ","
	res += strconv.Itoa(this.objId) + ","
	switch this.op {
	case CondWait:
		res += "W"
	case CondSignal:
		res += "S"
	case CondBroadcast:
		res += "B"
	}
	res += "," + this.Pos()
	return res
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementCond) Function() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

// Vc sets the vector clock
//
// Parameter:
//   - weak bool: set the weak wv
//   - cl *clock.VectorClock: the vector clock
func (this *ElementCond) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementCond) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
	return this.ci.getVC(weak)
}

// NumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Parameter:
//   - weak bool: get number of weak concurrent
//   - sameElem bool: only operation on the same variable
//
// Returns:
//   - number of concurrent element, or -1
func (this *ElementCond) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementCond) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementCond) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.pos.file, this.pos.line)
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy the element
//
// Parameter:
//   - mapping map[string]Element: map containing all already copied elements.
//     since conds do not contain reference to other elements and no other
//     elements contain referents to conds, this is not used
//   - keep bool: if true, keep vc and order information

// Returns:
//   - TraceElement: The copy of the element
func (this *ElementCond) Copy(mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementCond{
			id:       this.id,
			index:    0,
			routine:  this.routine,
			tReq:     0,
			tCom:     0,
			objId:    this.objId,
			op:       this.op,
			pos:      this.pos.copy(),
			ci:       newConcInfo(),
			function: this.function.Copy(mapping, keep).(*ElementFunc),
		}
	}

	return &ElementCond{
		id:       this.id,
		index:    this.index,
		routine:  this.routine,
		tReq:     this.tReq,
		tCom:     this.tCom,
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

func (this *ElementCond) IsValid() bool {
	return this != nil
}
