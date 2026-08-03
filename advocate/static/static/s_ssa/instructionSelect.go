// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSelect.go
// Brief: Select Instruction
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

type InstructionSelect struct {
	InstructionBase
}

func newSelect(f *Function, inst *ssa.Select, i int) *InstructionSelect {
	return &InstructionSelect{InstructionBase: newInstructionBase(f, Ic_select, inst, i)}
}

func (this *InstructionSelect) Instruction() *ssa.Select {
	return this.inst.(*ssa.Select)
}

func (this *InstructionSelect) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionSelect) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionSelect IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionSelect) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)

	log.Error("InstructionSelect IMPLEMENTED YET")

	return this.Next(), info
}
