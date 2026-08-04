// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionTypeAssert.go
// Brief: TypeAssert Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionTypeAssert struct {
	InstructionBase
}

func newTypeAssert(f *Function, inst *ssa.TypeAssert, i int) *InstructionTypeAssert {
	return &InstructionTypeAssert{InstructionBase: newInstructionBase(f, Ic_typeAssert, inst, i)}
}

func (this *InstructionTypeAssert) Instruction() *ssa.TypeAssert {
	return this.inst.(*ssa.TypeAssert)
}

func (this *InstructionTypeAssert) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
