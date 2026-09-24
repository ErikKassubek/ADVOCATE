// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSlice.go
// Brief: Slice Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionSlice struct {
	InstructionBase
}

func newSlice(f *Function, inst *ssa.Slice, i int) *InstructionSlice {
	return &InstructionSlice{InstructionBase: newInstructionBase(f, Ic_slice, inst, i)}
}

func (this *InstructionSlice) Instruction() *ssa.Slice {
	return this.inst.(*ssa.Slice)
}

func (this *InstructionSlice) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
