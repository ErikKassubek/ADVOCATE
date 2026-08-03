// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeSlice.go
// Brief: Make slice Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionMakeSlice struct {
	InstructionBase
}

func newMakeSlice(f *Function, inst *ssa.MakeSlice, i int) *InstructionMakeSlice {
	return &InstructionMakeSlice{InstructionBase: newInstructionBase(f, Ic_makeSlice, inst, i)}
}

func (this *InstructionMakeSlice) Instruction() *ssa.MakeSlice {
	return this.inst.(*ssa.MakeSlice)
}

func (this *InstructionMakeSlice) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
