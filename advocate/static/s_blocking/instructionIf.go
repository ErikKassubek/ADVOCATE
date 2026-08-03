// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIf.go
// Brief: If Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

func ParseIf(inst *s_ssa.InstructionIf, _ int, elem trace.Element) (i s_ssa.Instruction, info *instructionWithInfo) {
	e := elem.(*trace.ElementControllFlow)

	switch elem.Type(true) {
	case trace.ControllIf:
		i = followIfChain(inst, e.ChosenCase())
	case trace.ControllSwitch:
		i = followSwitchChain(inst, e.ChosenCase())
	}
	return i, nil
}

func followIfChain(inst *s_ssa.InstructionIf, chosen int) s_ssa.Instruction {
	if chosen == 0 {
		return inst.FirstInBlock(inst.If())
	} else if chosen == 1 {
		return inst.FirstInBlock(inst.Else())
	}

	inst, ok := inst.FirstInBlock(inst.Else()).(*s_ssa.InstructionIf)
	if !ok {
		return inst
	}
	return followIfChain(inst, chosen-1)
}

func followSwitchChain(inst *s_ssa.InstructionIf, chosen int) s_ssa.Instruction {
	return followSwitchRec(inst, chosen)
}

func followSwitchRec(inst s_ssa.Instruction, chosen int) s_ssa.Instruction {
	if _, ok := inst.(*s_ssa.InstructionBinOp); ok {
		next := inst.Next()
		return followSwitchRec(next, chosen)
	}

	instrIf, ok := inst.(*s_ssa.InstructionIf)
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
