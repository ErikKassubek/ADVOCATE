// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionJump.go
// Brief: Jump Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionJump struct {
	InstructionBase

	to int
}

func newJump(f *Function, inst *ssa.Jump, i int) *InstructionJump {
	return &InstructionJump{InstructionBase: newInstructionBase(f, Ic_jump, inst, i), to: inst.Block().Succs[0].Index}
}

func (this *InstructionJump) Instruction() *ssa.Jump {
	return this.inst.(*ssa.Jump)
}

func (this *InstructionJump) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionJump) To() int {
	return this.to
}
