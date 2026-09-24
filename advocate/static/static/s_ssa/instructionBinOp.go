// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionBinOp.go
// Brief: Binary Operation Instruciton
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionBinOp struct {
	InstructionBase
}

func NewBinOp(f *Function, inst *ssa.BinOp, i int) *InstructionBinOp {
	return &InstructionBinOp{InstructionBase: newInstructionBase(f, Ic_binOp, inst, i)}
}

func (this *InstructionBinOp) Instruction() *ssa.BinOp {
	return this.inst.(*ssa.BinOp)
}

func (this *InstructionBinOp) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
