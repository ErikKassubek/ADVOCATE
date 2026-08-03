// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionChangeType.go
// Brief: Change Type Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionChangeType struct {
	InstructionBase
}

func NewChangeType(f *Function, inst *ssa.ChangeType, i int) *InstructionChangeType {
	return &InstructionChangeType{InstructionBase: newInstructionBase(f, Ic_changeType, inst, i)}
}

func (this *InstructionChangeType) Instruction() *ssa.ChangeType {
	return this.inst.(*ssa.ChangeType)
}

func (this *InstructionChangeType) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
