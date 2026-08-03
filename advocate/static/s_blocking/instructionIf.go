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

func ParseIf(inst *s_ssa.InstructionIf, rout int, elem trace.Element) (i s_ssa.Instruction, info *instructionWithInfo) {
	e := elem.(*trace.ElementControllFlow)

	switch elem.Type(true) {
	case trace.ControllIf:
		i = followIfChain(inst, rout, e.ChosenCase())
	case trace.ControllSwitch:
		i = followSwitchChain(inst, rout, e.ChosenCase())
	}
	return i, nil
}

func followIfChain(inst *s_ssa.InstructionIf, rout, chosen int) s_ssa.Instruction {
	switch chosen {
	case 0:
		return inst.FirstInBlock(inst.If())
	case 1:
		return inst.FirstInBlock(inst.Else())
	}

	inst, ok := inst.FirstInBlock(inst.Else()).(*s_ssa.InstructionIf)
	if !ok {
		return inst
	}
	return followIfChain(inst, rout, chosen-1)
}

func followSwitchChain(inst *s_ssa.InstructionIf, rout, chosen int) s_ssa.Instruction {
	return followSwitchRec(inst, rout, chosen)
}

func followSwitchRec(inst s_ssa.Instruction, rout, chosen int) s_ssa.Instruction {
	if _, ok := inst.(*s_ssa.InstructionBinOp); ok {
		next := inst.Next()
		return followSwitchRec(next, rout, chosen)
	}

	instrIf := inst.(*s_ssa.InstructionIf)

	switch chosen {
	case 0:
		return inst.FirstInBlock(instrIf.If())
	case 1:
		return inst.FirstInBlock(instrIf.Else())
	}

	inst = inst.FirstInBlock(instrIf.Else())

	return followSwitchRec(inst, rout, chosen-1)
}
