// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSliceToArrayPointer.go
// Brief: Slice to array pointer Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionSliceToArrayPointer struct {
	InstructionBase
}

func newSliceToArrayPointer(f *Function, inst *ssa.SliceToArrayPointer, i int) *InstructionSliceToArrayPointer {
	return &InstructionSliceToArrayPointer{InstructionBase: newInstructionBase(f, Ic_sliceToArrayPointer, inst, i)}
}

func (this *InstructionSliceToArrayPointer) Instruction() *ssa.SliceToArrayPointer {
	return this.inst.(*ssa.SliceToArrayPointer)
}

func (this *InstructionSliceToArrayPointer) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
