// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionConvert.go
// Brief: convert Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionConvert struct {
	InstructionBase
}

func NewConvert(f *Function, inst *ssa.Convert, i int) *InstructionConvert {
	return &InstructionConvert{InstructionBase: newInstructionBase(f, Ic_convert, inst, i)}
}

func (this *InstructionConvert) Instruction() *ssa.ChangeType {
	return this.inst.(*ssa.ChangeType)
}

func (this *InstructionConvert) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
