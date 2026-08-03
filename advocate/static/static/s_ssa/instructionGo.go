// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionGo.go
// Brief: Go Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionGo struct {
	InstructionBase
}

func NewGo(f *Function, inst *ssa.Go, i int) *InstructionGo {
	return &InstructionGo{InstructionBase: newInstructionBase(f, Ic_go, inst, i)}
}

func (this *InstructionGo) Instruction() *ssa.Go {
	return this.inst.(*ssa.Go)
}

func (this *InstructionGo) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}
