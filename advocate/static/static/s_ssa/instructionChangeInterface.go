// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionChangeInterface.go
// Brief: Change Interface Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionChangeInterface struct {
	InstructionBase
}

func NewChangeInterface(f *Function, inst *ssa.ChangeInterface, i int) *InstructionChangeInterface {
	return &InstructionChangeInterface{InstructionBase: newInstructionBase(f, Ic_changeInterface, inst, i)}
}

func (this *InstructionChangeInterface) Instruction() *ssa.ChangeInterface {
	return this.inst.(*ssa.ChangeInterface)
}

func (this *InstructionChangeInterface) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
