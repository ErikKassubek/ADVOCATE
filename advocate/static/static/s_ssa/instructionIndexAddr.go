// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIndexAddr.go
// Brief: Index address Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionIndexAddr struct {
	InstructionBase
}

func NewIndexAddr(f *Function, inst *ssa.IndexAddr, i int) *InstructionIndexAddr {
	return &InstructionIndexAddr{InstructionBase: newInstructionBase(f, Ic_indexAddr, inst, i)}
}

func (this *InstructionIndexAddr) Instruction() *ssa.IndexAddr {
	return this.inst.(*ssa.IndexAddr)
}

func (this *InstructionIndexAddr) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
