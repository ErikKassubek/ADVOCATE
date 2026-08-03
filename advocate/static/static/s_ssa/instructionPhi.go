// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionPhi.go
// Brief: Phi Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionPhi struct {
	InstructionBase
}

func newPhi(f *Function, inst *ssa.Phi, i int) *InstructionPhi {
	return &InstructionPhi{InstructionBase: newInstructionBase(f, Ic_phi, inst, i)}
}

func (this *InstructionPhi) Instruction() *ssa.Phi {
	return this.inst.(*ssa.Phi)
}

func (this *InstructionPhi) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
