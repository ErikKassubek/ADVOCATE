// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeInterface.go
// Brief: Make interface Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionMakeInterface struct {
	InstructionBase
}

func newMakeInterface(f *Function, inst *ssa.MakeInterface, i int) *InstructionMakeInterface {
	return &InstructionMakeInterface{InstructionBase: newInstructionBase(f, Ic_makeInterface, inst, i)}
}

func (this *InstructionMakeInterface) Instruction() *ssa.MakeInterface {
	return this.inst.(*ssa.MakeInterface)
}

func (this *InstructionMakeInterface) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
