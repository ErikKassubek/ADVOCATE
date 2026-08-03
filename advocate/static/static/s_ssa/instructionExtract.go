// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionExtrace.go
// Brief: Extract Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionExtract struct {
	InstructionBase
}

func NewExtract(f *Function, inst *ssa.Extract, i int) *InstructionExtract {
	return &InstructionExtract{InstructionBase: newInstructionBase(f, Ic_extract, inst, i)}
}

func (this *InstructionExtract) Instruction() *ssa.Extract {
	return this.inst.(*ssa.Extract)
}

func (this *InstructionExtract) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = false
}
