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
	"advocate/trace"

	"golang.org/x/tools/go/ssa"
)

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

func (this *InstructionIf) Parse(_ *Data, _ int, elem trace.Element) (inst Instruction, info *InstructionWithInfo) {
	e := elem.(*trace.ElementControllFlow)

	switch elem.Type(true) {
	case trace.ControllIf:
		inst = this.FollowIfChain(e.ChosenCase())
	case trace.ControllSwitch:
		inst = this.FollowSwitchChain(e.ChosenCase())
	}
	return inst, nil
}

func (this *InstructionIf) FollowIfChain(chosen int) Instruction {
	if chosen == 0 {
		return this.FirstInBlock(this.If())
	} else if chosen == 1 {
		return this.FirstInBlock(this.Else())
	}

	inst, ok := this.FirstInBlock(this.Else()).(*InstructionIf)
	if !ok {
		return inst
	}
	return inst.FollowIfChain(chosen - 1)
}

func (this *InstructionIf) FollowSwitchChain(chosen int) Instruction {
	return followSwitchRec(this, chosen)
}

func followSwitchRec(inst Instruction, chosen int) Instruction {
	if _, ok := inst.(*InstructionBinOp); ok {
		next := inst.Next()
		return followSwitchRec(next, chosen)
	}

	instrIf, ok := inst.(*InstructionIf)
	if !ok {
		return inst
	}

	if chosen == 0 {
		inst = inst.FirstInBlock(instrIf.If())
		return inst
	}

	inst = inst.FirstInBlock(instrIf.Else())
	return followSwitchRec(inst, chosen-1)
}
