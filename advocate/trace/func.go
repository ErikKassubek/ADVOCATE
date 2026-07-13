// Copyright (c) 2024 Erik Kassubek
//
// File: func.go
// Brief: Records the execution and return of a function
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
	id      int
	index   int
	routine int
	name    string
	t       int
	file    string
	line    int
}

type ElementReturn struct {
	id      int
	index   int
	routine int
	t       int
	call    *ElementFunc
}

func (this *Trace) AddTaceElementFunc(routine int, t string, name string, op OperationType, pos string) error {
	tInt, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("t is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementFunc{
		index:   this.NumberElemInRoutine(routine),
		routine: routine,
		name:    name,
		t:       tInt,
		file:    file,
		line:    line,
	}

	if _, ok := lastCall[routine]; !ok {
		lastCall[routine] = types.NewStack[*ElementFunc]()
	}
	lastCall[routine].Push(&elem)

	this.AddElement(&elem)
	return nil
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

func (this *ElementFunc) setID(ID int) {
	this.id = ID
}

func (this *ElementFunc) GetID() int {
	return this.id
}

func (this *ElementFunc) GetObjId() int {
	return -1
}

func (this *ElementFunc) GetT(_ timeType) int {
	return this.t
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementFunc) Committed() bool {
	return true
}

func (this *ElementFunc) GetPos() string {
	return fmt.Sprintf("%s%s%d", this.file, consts.PosSep, this.line)
}

func (this *ElementFunc) GetFile() string {
	return this.file
}

func (this *ElementFunc) GetLine() int {
	return this.line
}

func (this *ElementFunc) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

func (this *ElementFunc) GetType(operation bool) OperationType {
	if operation {
		return FuncCall
	}

	return Func
}

func (this *ElementFunc) GetTID() string {
	return "F@" + this.GetPos() + "@" + strconv.Itoa(this.t)
}

func (this *ElementFunc) GetRoutine() int {
	return this.routine
}

func (this *ElementFunc) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

func (this *ElementFunc) IsSameElement(elem Element) bool {
	switch e := elem.(type) {
	case *ElementFunc:
		return this.name == e.name
	}

	return false
}

func (this *ElementFunc) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

func (this *ElementFunc) SetT(_ timeType, time int) {
	this.t = time
}

func (this *ElementFunc) SetTWithoutNotExecuted(t int) {
	this.t = t
}

func (this *ElementFunc) ToString() string {
	return fmt.Sprint("F,%d,%s,%s", this.t, this.name, this.GetPos())
}

func (this *ElementFunc) SetVc(_ *a_clock.VectorClock) {
}

func (this *ElementFunc) SetWVc(_ *a_clock.VectorClock) {
}

func (this *ElementFunc) GetVC() *a_clock.VectorClock {
	return nil
}

func (this *ElementFunc) GetWVC() *a_clock.VectorClock {
	return nil
}

func (this *ElementFunc) Copy(mapping map[string]Element, keep bool) Element {
	return &ElementFunc{
		id:      this.id,
		index:   this.index,
		routine: this.routine,
		t:       this.t,
		name:    this.name,
		file:    this.file,
		line:    this.line,
	}
}

func (this *ElementFunc) GetNumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementFunc) SetNumberConcurrent(_ int, _, _ bool) {
}

func (this *ElementFunc) IsValid() bool {
	return this != nil
}

func (this *ElementReturn) setID(ID int) {
	this.id = ID
}

func (this *ElementReturn) GetID() int {
	return this.id
}

func (this *ElementReturn) GetObjId() int {
	return -1
}

func (this *ElementReturn) GetT(_ timeType) int {
	return this.t
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementReturn) Committed() bool {
	return true
}

func (this *ElementReturn) GetPos() string {
	return ""
}

func (this *ElementReturn) GetFile() string {
	return ""
}

func (this *ElementReturn) GetLine() int {
	return -1
}

func (this *ElementReturn) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, "", -1)
}

func (this *ElementReturn) GetType(operation bool) OperationType {
	if operation {
		return FuncReturn
	}

	return Func
}

func (this *ElementReturn) GetTID() string {
	return "R@" + this.GetPos() + "@" + strconv.Itoa(this.t)
}

func (this *ElementReturn) GetRoutine() int {
	return this.routine
}

func (this *ElementReturn) IsEqual(elem Element) bool { // TODO: fix
	return this.routine == elem.GetRoutine()
}

func (this *ElementReturn) IsSameElement(elem Element) bool {
	switch e := elem.(type) {
	case *ElementReturn:
		return this.call.name == e.call.name
	}

	return false
}

func (this *ElementReturn) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

func (this *ElementReturn) SetTWithoutNotExecuted(t int) {
	this.t = t
}

func (this *ElementReturn) SetT(_ timeType, t int) {
	this.t = t
}

func (this *ElementReturn) ToString() string {
	return fmt.Sprint("R,%d", this.t)
}

func (this *ElementReturn) SetVc(_ *a_clock.VectorClock) {
}

func (this *ElementReturn) SetWVc(_ *a_clock.VectorClock) {
}

func (this *ElementReturn) GetVC() *a_clock.VectorClock {
	return nil
}

func (this *ElementReturn) GetWVC() *a_clock.VectorClock {
	return nil
}

func (this *ElementReturn) Copy(mapping map[string]Element, keep bool) Element {
	return &ElementReturn{
		id:      this.id,
		index:   this.index,
		routine: this.routine,
		t:       this.t,
	}
}

func (this *ElementReturn) GetNumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementReturn) SetNumberConcurrent(_ int, _, _ bool) {
}

func (this *ElementReturn) IsValid() bool {
	return this != nil
}
