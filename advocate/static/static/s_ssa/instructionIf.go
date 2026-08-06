// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIf.go
// Brief: If Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionIf struct {
	InstructionBase

	case_if   int
	case_else int
}

func newIf(f *Function, inst *ssa.If, i int, data *Data) *InstructionIf {
	// Note: During the creation for the SSA, if true:bool goto 8 else 7, does not need to mean, that
	// 8 corresponds to the if and 7 to the select. If statements like 'if !x {', exists, it can be flipped.
	// But it seams theat the order of blocks in the SSA is equal to the order in the code. (TODO: only from observation, check if this is always true)
	// We can therefore determine the Succ block with lower id to be the if case

	succ0 := inst.Block().Succs[0]
	succ1 := inst.Block().Succs[1]

	if succ0.Index > succ1.Index {
		succ0, succ1 = succ1, succ0
	}

	return &InstructionIf{
		InstructionBase: newInstructionBase(f, Ic_if, inst, i),
		case_if:         succ0.Index,
		case_else:       succ1.Index,
	}
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
