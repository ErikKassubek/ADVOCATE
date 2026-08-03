// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionField.go
// Brief: Field Instruction
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

type InstructionField struct {
	InstructionBase
}

func NewField(f *Function, inst *ssa.Field, i int) *InstructionField {
	return &InstructionField{InstructionBase: newInstructionBase(f, Ic_field, inst, i)}
}

func (this *InstructionField) Instruction() *ssa.Field {
	return this.inst.(*ssa.Field)
}

func (this *InstructionField) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionField) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionField IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionField) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
