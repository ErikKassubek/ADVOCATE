// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeChan.go
// Brief: Make channel Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionMakeChan struct {
	InstructionBase
}

func newMakeChan(f *Function, inst *ssa.MakeChan, i int) *InstructionMakeChan {
	return &InstructionMakeChan{InstructionBase: newInstructionBase(f, Ic_makeChan, inst, i)}
}

func (this *InstructionMakeChan) Instruction() *ssa.MakeChan {
	return this.inst.(*ssa.MakeChan)
}

func (this *InstructionMakeChan) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}
