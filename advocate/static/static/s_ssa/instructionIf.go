// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIf.go
// Brief: If Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionIf struct {
	InstructionBase

	case_if   int
	case_else int
}

func newIf(f *Function, inst *ssa.If, i int) *InstructionIf {
	ind := inst.Block().Succs[0].Index
	e := inst.Block().Succs[1].Index
	return &InstructionIf{InstructionBase: newInstructionBase(f, Ic_if, inst, i), case_if: ind, case_else: e}
}

func (this *InstructionIf) Instruction() *ssa.If {
	return this.inst.(*ssa.If)
}

func (this *InstructionIf) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionIf) If() int {
	return this.case_if
}

func (this *InstructionIf) Else() int {
	return this.case_else
}
