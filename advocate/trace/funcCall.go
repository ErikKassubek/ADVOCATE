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
	"advocate/utils/consts"
	"advocate/utils/types"
	"errors"
	"fmt"
	"strconv"
)

var lastCall = make(map[int]*types.Stack[*ElementFunc])

func getLastCall(rout int) *ElementFunc {
	rStack, ok := lastCall[rout]
	if !ok {
		return nil
	}

	if rStack.IsEmpty() {
		return nil
	}

	return rStack.Peek()
}

// ========================================================
// MARK: Data
// ========================================================

// ElementFunc is a struct to save a function call in the trace
// Fields:
//
//   - id: id of the element, should never be changed
//   - index int: index in the routine
//   - routine int: The routine id
//   - name: name of the called function
//   - t int: The timestamp of the event
//   - posDef *position: code position of the function declaration
//   - posCall *position: code position of the function call
//   - function *ElementFunc: the function the operation is in
type ElementFunc struct {
	id       int
	index    int
	routine  int
	name     string
	t        int
	posDef   *position
	posCall  *position
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

func (this *Trace) AddTaceElementFunc(routine int, t string, name string, op OperationType, posDef, posCall string) error {
	tInt, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("t is not an integer")
	}

	fileDef, lineDef, err := PosFromPosString(posDef)
	if err != nil {
		return err
	}

	fileCall, lineCall, err := PosFromPosString(posCall)
	if err != nil {
		return err
	}

	elem := ElementFunc{
		index:    this.NumberElemInRoutine(routine),
		routine:  routine,
		name:     name,
		t:        tInt,
		posDef:   newPosition(fileDef, lineDef),
		posCall:  newPosition(fileCall, lineCall),
		function: getLastCall(routine),
	}

	if _, ok := lastCall[routine]; !ok {
		lastCall[routine] = types.NewStack[*ElementFunc]()
	}
	lastCall[routine].Push(&elem)

	this.AddElement(&elem)
	return nil
}

// ========================================================
// MARK: ID
// ========================================================

func (this *ElementFunc) GetID() int {
	return this.id
}

func (this *ElementFunc) setID(ID int) {
	this.id = ID
}

func (this *ElementFunc) GetObjId() int {
	return -1
}

// ========================================================
// MARK: Timestamps
// ========================================================

func (this *ElementFunc) GetT(_ timeType) int {
	return this.t
}

func (this *ElementFunc) SetT(_ timeType, time int) {
	this.t = time
}

func (this *ElementFunc) SetTWithoutNotExecuted(t int) {
	this.t = t
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementFunc) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

func (this *ElementFunc) GetPos() string {
	return this.posCall.toString()
}

func (this *ElementFunc) GetFile() string {
	return this.posCall.file
}

func (this *ElementFunc) GetLine() int {
	return this.posCall.line
}

func (this *ElementFunc) GetPosDef() string {
	return fmt.Sprintf("%s%s%d", this.posCall.file, consts.PosSep, this.posCall.line)
}

// ========================================================
// MARK: Index
// ========================================================

func (this *ElementFunc) GetRoutine() int {
	return this.routine
}

func (this *ElementFunc) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// MARK: Operation
// ========================================================

func (this *ElementFunc) GetType(operation bool) OperationType {
	if operation {
		return FuncCall
	}

	return Func
}

// ========================================================
// MARK: Equal
// ========================================================

func (this *ElementFunc) IsEqual(elem Element) bool {
	return this.id == elem.GetID()
}

func (this *ElementFunc) IsSameElement(elem Element) bool {
	switch e := elem.(type) {
	case *ElementFunc:
		return this.name == e.name
	}

	return false
}

// ========================================================
// MARK: String
// ========================================================

func (this *ElementFunc) ToString() string {
	return fmt.Sprint("F,%d,%s,%s,%s", this.t, this.name, this.GetPosDef(), this.GetPos())
}

func (this *ElementFunc) GetTID() string {
	return "F@" + this.GetPos() + "@" + strconv.Itoa(this.t)
}

// ========================================================
// MARK: VC
// ========================================================

func (this *ElementFunc) SetVc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

func (this *ElementFunc) GetVC(a_clock.VcType) *a_clock.VectorClock {
	return &a_clock.VectorClock{}
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementFunc) GetFunction() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

func (this *ElementFunc) GetNumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementFunc) SetNumberConcurrent(_ int, _, _ bool) {
}

// ========================================================
// MARK: Replay
// ========================================================

func (this *ElementFunc) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.posCall.file, this.posCall.line)
}

// ========================================================
// MARK: Copy
// ========================================================

func (this *ElementFunc) Copy(mapping map[int]Element, keep bool) Element {
	id := this.GetID()

	if existing, ok := mapping[id]; ok {
		return existing
	}

	elem := &ElementFunc{
		id:      this.id,
		index:   this.index,
		routine: this.routine,
		t:       this.t,
		name:    this.name,
		posDef:  this.posDef.copy(),
		posCall: this.posCall.copy(),
	}

	mapping[id] = elem

	return elem
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementFunc) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

func (this *ElementFunc) GetName() string {
	return this.name
}
