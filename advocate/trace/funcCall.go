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

type ElementFunc struct {
	id       int
	index    int
	routine  int
	name     string
	t        int
	fileDef  string
	lineDef  int
	fileCall string
	lineCall int
}

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
		fileDef:  fileDef,
		lineDef:  lineDef,
		fileCall: fileCall,
		lineCall: lineCall,
	}

	if _, ok := lastCall[routine]; !ok {
		lastCall[routine] = types.NewStack[*ElementFunc]()
	}
	lastCall[routine].Push(&elem)

	this.AddElement(&elem)
	return nil
}

// ========================================================
// ID
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
// Timestamps
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
// Position
// ========================================================

func (this *ElementFunc) GetPos() string {
	return fmt.Sprintf("%s%s%d", this.fileCall, consts.PosSep, this.lineCall)
}

func (this *ElementFunc) GetFile() string {
	return this.fileCall
}

func (this *ElementFunc) GetLine() int {
	return this.lineCall
}

func (this *ElementFunc) GetPosDef() string {
	return fmt.Sprintf("%s%s%d", this.fileDef, consts.PosSep, this.lineDef)
}

// ========================================================
// Index
// ========================================================

func (this *ElementFunc) GetRoutine() int {
	return this.routine
}

func (this *ElementFunc) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// Operation
// ========================================================

func (this *ElementFunc) GetType(operation bool) OperationType {
	if operation {
		return FuncCall
	}

	return Func
}

// ========================================================
// Equal
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
// String
// ========================================================

func (this *ElementFunc) ToString() string {
	return fmt.Sprint("F,%d,%s,%s,%s", this.t, this.name, this.GetPosDef(), this.GetPos())
}

func (this *ElementFunc) GetTID() string {
	return "F@" + this.GetPos() + "@" + strconv.Itoa(this.t)
}

// ========================================================
// VC
// ========================================================

func (this *ElementFunc) SetVc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

func (this *ElementFunc) GetVC(a_clock.VcType) *a_clock.VectorClock {
	return &a_clock.VectorClock{}
}

// ========================================================
// Concurrent
// ========================================================

func (this *ElementFunc) GetNumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementFunc) SetNumberConcurrent(_ int, _, _ bool) {
}

// ========================================================
// Replay
// ========================================================

func (this *ElementFunc) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.fileCall, this.lineCall)
}

// ========================================================
// Copy
// ========================================================

func (this *ElementFunc) Copy(mapping map[string]Element, keep bool) Element {
	return &ElementFunc{
		id:       this.id,
		index:    this.index,
		routine:  this.routine,
		t:        this.t,
		name:     this.name,
		fileDef:  this.fileDef,
		lineDef:  this.lineDef,
		fileCall: this.fileCall,
		lineCall: this.lineCall,
	}
}

// ========================================================
// Valid
// ========================================================

func (this *ElementFunc) IsValid() bool {
	return this != nil
}
