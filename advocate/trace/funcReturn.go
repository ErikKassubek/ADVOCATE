// Copyright (c) 2024 Erik Kassubek
//
// File: funcCall.go
// Brief: Records the call of a function
//
// Author: Erik Kassubek
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
//   - t int: The timestamp of the event
//   - function *ElementFunc: corresponding function call
type ElementReturn struct {
	ElementBase

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
		ElementBase: this.newElementBase(routine),
		t:           tInt,
		function:    call,
	}

	this.AddElement(&elem)
	return nil
}

// ========================================================
// MARK: ID
// ========================================================

func (this *ElementReturn) ResourceID() int {
	return -1
}

// ========================================================
// MARK: Timestamps
// ========================================================

func (this *ElementReturn) T(_ timeType) int {
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

func (this *ElementReturn) Pos() Position {
	return newPosition("", 0)
}

func (this *ElementReturn) File() string {
	return ""
}

func (this *ElementReturn) Line() int {
	return -1
}

// ========================================================
// MARK: Operation
// ========================================================

func (this *ElementReturn) Type(operation bool) OperationType {
	if operation {
		return FuncReturn
	}

	return Func
}

// ========================================================
// MARK: Equal
// ========================================================

func (this *ElementReturn) IsEqual(elem Element) bool {
	return this.id == elem.ID()
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

func (this *ElementReturn) String() string {
	return fmt.Sprintf("R,%d", this.t)
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementReturn) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// ========================================================
// MARK: VC
// ========================================================

func (this *ElementReturn) Vc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

func (this *ElementReturn) GetVC(_ a_clock.VcType) *a_clock.VectorClock {
	return nil
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementReturn) Function() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

func (this *ElementReturn) NumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementReturn) SetNumberConcurrent(_ int, _, _ bool) {
}

// ========================================================
// MARK: Replay
// ========================================================

func (this *ElementReturn) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, "", -1)
}

// ========================================================
// MARK: Copy
// ========================================================

func (this *ElementReturn) Copy(mapping map[int]Element, keep bool) Element {
	return &ElementReturn{
		ElementBase: this.ElementBase.Copy(),
		t:           this.t,
		function:    this.function.CopyFunc(mapping, keep),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementReturn) IsValid() bool {
	return this != nil
}
