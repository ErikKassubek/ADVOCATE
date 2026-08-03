// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSlice.go
// Brief: Slice Instruction
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

type InstructionSlice struct {
	InstructionBase
}

func newSlice(f *Function, inst *ssa.Slice, i int) *InstructionSlice {
	return &InstructionSlice{InstructionBase: newInstructionBase(f, Ic_slice, inst, i)}
}

func (this *InstructionSlice) Instruction() *ssa.Slice {
	return this.inst.(*ssa.Slice)
}

func (this *InstructionSlice) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionSlice) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionSlice IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionSlice) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)

	return this.Next(), info
}
