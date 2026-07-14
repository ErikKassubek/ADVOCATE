// Copyright (c) 2024 Erik Kassubek
//
// File: funcCall.go
// Brief: Records the call of a function
//
// Author: Erik Kassubek
// Created: 2026-07-13
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

// ElementReturn is a struct to save a function return in the trace
// Fields:
//
//   - id: id of the element, should never be changed
//   - index int: index in the routine
//   - routine int: The routine id
//   - t int: The timestamp of the event
//   - function *ElementFunc: corresponding function call
type ElementReturn struct {
	id       int
	index    int
	routine  int
	t        int
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

func (this *Trace) AddTaceElementReturn(routine int, t string) error {
	tInt, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("t is not an integer")
	}

	var call *ElementFunc = nil
	if _, ok := lastCall[routine]; ok {
		call = lastCall[routine].Pop()
	}

	elem := ElementReturn{
		index:    this.NumberElemInRoutine(routine),
		routine:  routine,
		t:        tInt,
		function: call,
	}

	this.AddElement(&elem)
	return nil
}

// ========================================================
// MARK: ID
// ========================================================

func (this *ElementReturn) GetID() int {
	return this.id
}

func (this *ElementReturn) setID(ID int) {
	this.id = ID
}

func (this *ElementReturn) GetObjId() int {
	return -1
}

// ========================================================
// MARK: Timestamps
// ========================================================

func (this *ElementReturn) GetT(_ timeType) int {
	return this.t
}

func (this *ElementReturn) SetT(_ timeType, t int) {
	this.t = t
}

func (this *ElementReturn) SetTWithoutNotExecuted(t int) {
	this.t = t
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementReturn) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

func (this *ElementReturn) GetPos() string {
	return ""
}

func (this *ElementReturn) GetFile() string {
	return ""
}

func (this *ElementReturn) GetLine() int {
	return -1
}

// ========================================================
// MARK: Index
// ========================================================

func (this *ElementReturn) GetRoutine() int {
	return this.routine
}

func (this *ElementReturn) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// MARK: Operation
// ========================================================

func (this *ElementReturn) GetType(operation bool) OperationType {
	if operation {
		return FuncReturn
	}

	return Func
}

// ========================================================
// MARK: Equal
// ========================================================

func (this *ElementReturn) IsEqual(elem Element) bool { // TODO: fix
	return this.id == elem.GetID()
}

func (this *ElementReturn) IsSameElement(elem Element) bool {
	switch e := elem.(type) {
	case *ElementReturn:
		return this.function.name == e.function.name
	}

	return false
}

// ========================================================
// MARK: String
// ========================================================

func (this *ElementReturn) ToString() string {
	return fmt.Sprintf("R,%d", this.t)
}

// ========================================================
// MARK: VC
// ========================================================

func (this *ElementReturn) SetVc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

func (this *ElementReturn) GetVC(_ a_clock.VcType) *a_clock.VectorClock {
	return nil
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementReturn) GetFunction() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

func (this *ElementReturn) GetNumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementReturn) SetNumberConcurrent(_ int, _, _ bool) {
}

// ========================================================
// MARK: Replay
// ========================================================

func (this *ElementReturn) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, "", -1)
}

// ========================================================
// MARK: Copy
// ========================================================

func (this *ElementReturn) Copy(mapping map[int]Element, keep bool) Element {
	return &ElementReturn{
		id:       this.id,
		index:    this.index,
		routine:  this.routine,
		t:        this.t,
		function: this.function.Copy(mapping, keep).(*ElementFunc),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementReturn) IsValid() bool {
	return this != nil
}
