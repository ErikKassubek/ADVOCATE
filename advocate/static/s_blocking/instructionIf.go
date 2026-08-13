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

func ParseIfAll(inst *s_ssa.InstructionIf) []s_ssa.Instruction {
	res := make([]s_ssa.Instruction, 0)

	res = append(res, inst.FirstInBlock(inst.If()))

	elseBlock := inst.FirstInBlock(inst.Else())
	if inst, ok := elseBlock.(*s_ssa.InstructionIf); ok {
		res = append(res, ParseIfAll(inst)...)
	} else {
		if _, ok := elseBlock.(*s_ssa.InstructionBinOp); ok { // sometimes needed for switch
			next := elseBlock.Next()
			if nb, ok := next.(*s_ssa.InstructionIf); ok {
				res = append(res, ParseIfAll(nb)...)
			} else {
				res = append(res, elseBlock)
			}
		} else {
			res = append(res, elseBlock)
		}
	}

	return res
}
