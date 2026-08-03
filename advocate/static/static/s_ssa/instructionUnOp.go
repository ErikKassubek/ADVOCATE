// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionUnOp.go
// Brief: Unary instruction Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"go/token"

	"golang.org/x/tools/go/ssa"
)

type InstructionUnOp struct {
	InstructionBase
}

func newUnOp(f *Function, inst *ssa.UnOp, i int) *InstructionUnOp {
	return &InstructionUnOp{InstructionBase: newInstructionBase(f, Ic_unOp, inst, i)}
}

func (this *InstructionUnOp) Instruction() *ssa.UnOp {
	return this.inst.(*ssa.UnOp)
}

func (this *InstructionUnOp) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = this.Inst().(*ssa.UnOp).Op == token.ARROW // receive
}
