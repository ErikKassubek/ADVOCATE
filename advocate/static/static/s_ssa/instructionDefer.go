// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionDefer.go
// Brief: Defer Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionDefer struct {
	InstructionBase
}

func NewDefer(f *Function, inst *ssa.Defer, i int) *InstructionDefer {
	return &InstructionDefer{InstructionBase: newInstructionBase(f, Ic_defer, inst, i)}
}

func (this *InstructionDefer) Instruction() *ssa.Defer {
	return this.inst.(*ssa.Defer)
}

func (this *InstructionDefer) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
