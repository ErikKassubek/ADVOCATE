// Copyright (c) 2026 Erik Kassubek
//
// File: InstructioRunDefers.go
// Brief: Run defers Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionRunDefers struct {
	InstructionBase
}

func newRunDefers(f *Function, inst *ssa.RunDefers, i int) *InstructionRunDefers {
	return &InstructionRunDefers{InstructionBase: newInstructionBase(f, Ic_runDefers, inst, i)}
}

func (this *InstructionRunDefers) Instruction() *ssa.RunDefers {
	return this.inst.(*ssa.RunDefers)
}

func (this *InstructionRunDefers) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
