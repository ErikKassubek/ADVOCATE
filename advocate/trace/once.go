// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/once.go
// Brief: Struct and functions for once operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-09-25
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"advocate/analysis/hb/a_clock"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementOnce is a trace element for a once
// Fields:
//   - id: id of the element, should never be changed
//   - index int: index in the routine
//   - routine int: The routine id
//   - tReq int: The timestamp at the start of the event
//   - tCom int: The timestamp at the end of the event
//   - objId int: The id of the mutex
//   - pos *position: code position
//   - ci *concInfo: concurrency info
//   - suc bool: Whether the operation was successful
//   - function *ElementFunc: the function the operation is in
type ElementOnce struct {
	id       int
	index    int
	routine  int
	tReq     int
	tCom     int
	objId    int
	pos      *position
	ci       *concInfo
	suc      bool
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementOnce adds a new mutex trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tReq string: The timestamp at the start of the event
//   - tCom string: The timestamp at the end of the event
//   - id string: The id of the mutex
//   - suc string: Whether the operation was successful (only for trylock else always true)
//   - pos string: The position of the mutex operation in the code
func (this *Trace) AddTraceElementOnce(routine int, tReq string,
	tCom string, id string, suc string, pos string) error {
	tReqInt, err := strconv.Atoi(tReq)
	if err != nil {
		return errors.New("tReq is not an integer")
	}

	tComInt, err := strconv.Atoi(tCom)
	if err != nil {
		return errors.New("tCom is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementOnce{
		index:    this.NumberElemInRoutine(routine),
		routine:  routine,
		tReq:     tReqInt,
		tCom:     tComInt,
		objId:    idInt,
		suc:      sucBool,
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
func (this *ElementOnce) ID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementOnce) setID(ID int) {
	this.id = ID
}

// ObjID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementOnce) ObjID() int {
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
func (this *ElementOnce) T(t timeType) int {
	switch t {
	case Request:
		return this.tReq
	case Commit:
		return this.tCom
	case Sorting:
		if this.tCom == 0 {
			return math.MaxInt
		}
		return this.tCom
	}

	return this.tCom
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - t timeType: type of time to set
//   - time int: The tPre and tPost of the element
func (this *ElementOnce) SetT(t timeType, time int) {
	switch t {
	case Request:
		this.tReq = time
		if this.tCom != 0 && this.tCom < time {
			this.tCom = time
		}
	case Commit:
		this.tCom = time
		if time != 0 && this.tReq > time {
			this.tReq = time
		}
	case Sorting, Both:
		this.SetT(Request, time)
		this.SetT(Commit, time)
	}
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementOnce) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.tCom != 0 {
		this.tCom = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementOnce) Committed() bool {
	return this.tCom != 0
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementOnce) Pos() string {
	return this.pos.toString()
}

// File returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementOnce) File() string {
	return this.pos.file
}

// Line returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementOnce) Line() int {
	return this.pos.line
}

// ========================================================
// MARK: Index
// ========================================================

// Routine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementOnce) Routine() int {
	return this.routine
}

// TraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementOnce) TraceIndex() (int, int) {
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
func (this *ElementOnce) Type(operation bool) OperationType {
	if !operation {
		return Once
	}

	if this.suc {
		return OnceSuc
	}
	return OnceFail
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
func (this *ElementOnce) IsEqual(elem Element) bool {
	return this.objId == elem.ObjID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same once
func (this *ElementOnce) IsSameElement(elem Element) bool {
	if elem.Type(false) != Once {
		return false
	}

	return this.objId == elem.ObjID()
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementOnce) String() string {
	res := "O,"
	res += strconv.Itoa(this.tReq) + ","
	res += strconv.Itoa(this.tCom) + ","
	res += strconv.Itoa(this.objId) + ","
	if this.suc {
		res += "t"
	} else {
		res += "f"
	}
	res += "," + this.Pos()
	return res
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementOnce) Function() *ElementFunc {
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
func (this *ElementOnce) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementOnce) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementOnce) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementOnce) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementOnce) ReplayID() string {
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
func (this *ElementOnce) Copy(mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementOnce{
			id:       this.id,
			index:    0,
			routine:  this.routine,
			tReq:     0,
			tCom:     0,
			objId:    this.objId,
			suc:      false,
			pos:      this.pos.copy(),
			ci:       newConcInfo(),
			function: this.function.Copy(mapping, keep).(*ElementFunc),
		}
	}

	return &ElementOnce{
		id:       this.id,
		index:    this.index,
		routine:  this.routine,
		tReq:     this.tReq,
		tCom:     this.tCom,
		objId:    this.objId,
		suc:      this.suc,
		pos:      this.pos.copy(),
		ci:       this.ci.copy(),
		function: this.function.Copy(mapping, keep).(*ElementFunc),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementOnce) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

// GetSuc returns whether the once do was executed (successful)
//
// Returns:
//   - bool: true if function in Do was executed, false otherwise
func (this *ElementOnce) GetSuc() bool {
	return this.suc
}

// SetSuc sets whether the once do was executed successful
//
// Parameter:
//   - bool: true if function in Do was executed, false otherwise
func (this *ElementOnce) SetSuc(s bool) {
	this.suc = s
}
