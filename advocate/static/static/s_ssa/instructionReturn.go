// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionReturn.go
// Brief: Return Instruction
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

type InstructionReturn struct {
	InstructionBase
}

func newReturn(f *Function, inst *ssa.Return, i int) *InstructionReturn {
	return &InstructionReturn{InstructionBase: newInstructionBase(f, Ic_return, inst, i)}
}

func (this *InstructionReturn) Instruction() *ssa.Return {
	return this.inst.(*ssa.Return)
}

func (this *InstructionReturn) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionReturn) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionReturn IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionReturn) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)

	inst := data.Blocking.JumpBackPos[rout].Pop()

	return inst, info
}
