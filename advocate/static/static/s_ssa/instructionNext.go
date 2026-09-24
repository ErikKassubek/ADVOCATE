// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionNext.go
// Brief: Next Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionNext struct {
	InstructionBase
}

func newNext(f *Function, inst *ssa.Next, i int) *InstructionNext {
	return &InstructionNext{InstructionBase: newInstructionBase(f, Ic_next, inst, i)}
}

func (this *InstructionNext) Instruction() *ssa.Next {
	return this.inst.(*ssa.Next)
}

func (this *InstructionNext) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
