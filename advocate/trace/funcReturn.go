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

type ElementReturn struct {
	id      int
	index   int
	routine int
	t       int
	call    *ElementFunc
}

func (this *Trace) AddTaceElementReturn(routine int, t string) error {
	tInt, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("t is not an integer")
	}

	call := lastCall[routine].Pop()

	elem := ElementReturn{
		index:   this.NumberElemInRoutine(routine),
		routine: routine,
		t:       tInt,
		call:    call,
	}

	this.AddElement(&elem)
	return nil
}

// ========================================================
// ID
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
// Timestamps
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
// Position
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
// Index
// ========================================================

func (this *ElementReturn) GetRoutine() int {
	return this.routine
}

func (this *ElementReturn) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// Operation
// ========================================================

func (this *ElementReturn) GetType(operation bool) OperationType {
	if operation {
		return FuncReturn
	}

	return Func
}

// ========================================================
// Equal
// ========================================================

func (this *ElementReturn) IsEqual(elem Element) bool { // TODO: fix
	return this.id == elem.GetID()
}

func (this *ElementReturn) IsSameElement(elem Element) bool {
	switch e := elem.(type) {
	case *ElementReturn:
		return this.call.name == e.call.name
	}

	return false
}

// ========================================================
// String
// ========================================================

func (this *ElementReturn) ToString() string {
	return fmt.Sprint("R,%d", this.t)
}

func (this *ElementReturn) GetTID() string {
	return "R@" + this.GetPos() + "@" + strconv.Itoa(this.t)
}

// ========================================================
// VC
// ========================================================

func (this *ElementReturn) SetVc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

func (this *ElementReturn) GetVC(_ a_clock.VcType) *a_clock.VectorClock {
	return nil
}

// ========================================================
// Concurrent
// ========================================================

func (this *ElementReturn) GetNumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementReturn) SetNumberConcurrent(_ int, _, _ bool) {
}

// ========================================================
// Replay
// ========================================================

func (this *ElementReturn) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, "", -1)
}

// ========================================================
// Copy
// ========================================================

func (this *ElementReturn) Copy(mapping map[string]Element, keep bool) Element {
	return &ElementReturn{
		id:      this.id,
		index:   this.index,
		routine: this.routine,
		t:       this.t,
	}
}

// ========================================================
// Valid
// ========================================================

func (this *ElementReturn) IsValid() bool {
	return this != nil
}
