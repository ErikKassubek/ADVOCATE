// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeSlice.go
// Brief: Make slice Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"
	"advocate/utils/log"

	"golang.org/x/tools/go/ssa"
)

type InstructionMakeSlice struct {
	InstructionBase
}

func newMakeSlice(f *Function, inst *ssa.MakeSlice, i int) *InstructionMakeSlice {
	return &InstructionMakeSlice{InstructionBase: newInstructionBase(f, Ic_makeSlice, inst, i)}
}

func (this *InstructionMakeSlice) Instruction() *ssa.MakeSlice {
	return this.inst.(*ssa.MakeSlice)
}

func (this *InstructionMakeSlice) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionMakeSlice) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionMakeSlice IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionMakeSlice) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
