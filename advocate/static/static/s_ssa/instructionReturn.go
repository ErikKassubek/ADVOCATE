// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionReturn.go
// Brief: Return Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

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
