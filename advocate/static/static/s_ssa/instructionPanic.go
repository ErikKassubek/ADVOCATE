// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionPanic.go
// Brief: Panic Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"

	"golang.org/x/tools/go/ssa"
)

type InstructionPanic struct {
	InstructionBase
}

func newPanic(f *Function, inst *ssa.Panic, i int) *InstructionPanic {
	return &InstructionPanic{InstructionBase: newInstructionBase(f, Ic_panic, inst, i)}
}

func (this *InstructionPanic) Instruction() *ssa.Panic {
	return this.inst.(*ssa.Panic)
}

func (this *InstructionPanic) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionPanic) Parse(_ *Data, _ int, _ trace.Element) (Instruction, *InstructionWithInfo) {
	return this.Next(), nil
}
