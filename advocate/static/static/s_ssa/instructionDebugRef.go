// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionDebugRef.go
// Brief: Debug Ref Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionDebugRef struct {
	InstructionBase
}

func NewDebugRef(f *Function, inst *ssa.DebugRef, i int) *InstructionDebugRef {
	return &InstructionDebugRef{InstructionBase: newInstructionBase(f, Ic_debugRef, inst, i)}
}

func (this *InstructionDebugRef) Instruction() *ssa.DebugRef {
	return this.inst.(*ssa.DebugRef)
}

func (this *InstructionDebugRef) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
