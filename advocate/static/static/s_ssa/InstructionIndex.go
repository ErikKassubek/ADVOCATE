// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIndex.go
// Brief: Index Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionIndex struct {
	InstructionBase
}

func NewIndex(f *Function, inst *ssa.Index, i int) *InstructionIndex {
	return &InstructionIndex{InstructionBase: newInstructionBase(f, Ic_index, inst, i)}
}

func (this *InstructionIndex) Instruction() *ssa.Index {
	return this.inst.(*ssa.Index)
}

func (this *InstructionIndex) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
