// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeClosure.go
// Brief: Make closure Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionMakeClosure struct {
	InstructionBase
}

func newMakeClosure(f *Function, inst *ssa.MakeClosure, i int) *InstructionMakeClosure {
	return &InstructionMakeClosure{InstructionBase: newInstructionBase(f, Ic_makeClosure, inst, i)}
}

func (this *InstructionMakeClosure) Instruction() *ssa.MakeClosure {
	return this.inst.(*ssa.MakeClosure)
}

func (this *InstructionMakeClosure) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
