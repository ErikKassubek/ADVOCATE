// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionAlloc.go
// Brief: Alloc Instruciton
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionAlloc struct {
	InstructionBase
}

func NewAlloc(f *Function, inst *ssa.Alloc, i int) *InstructionAlloc {
	return &InstructionAlloc{InstructionBase: newInstructionBase(f, Ic_alloc, inst, i)}
}

func (this *InstructionAlloc) Instruction() *ssa.Alloc {
	return this.inst.(*ssa.Alloc)
}

func (this *InstructionAlloc) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = !this.Conc().Channel()
}
