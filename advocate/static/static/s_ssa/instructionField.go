// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionField.go
// Brief: Field Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

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
