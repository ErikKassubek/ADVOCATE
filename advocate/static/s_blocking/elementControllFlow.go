// Copyright (c) 2026 Erik Kassubek
//
// File: elementControllFlow.go
// Brief: Parse an if or switch element
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import "advocate/static/static/s_ssa"

func followIfChain(inst s_ssa.Instruction, chosen int) s_ssa.Instruction {
	instrIf, ok := inst.(*s_ssa.InstructionIf)
	if !ok {
		return inst
	}

	if chosen == 0 {
		inst = inst.FirstInBlock(instrIf.If())
		return inst
	}

	inst = inst.FirstInBlock(instrIf.Else())
	return followIfChain(inst, chosen-1)
}

func followSwitchChain(inst s_ssa.Instruction, chosen int) s_ssa.Instruction {
	if _, ok := inst.(*s_ssa.InstructionBinOp); ok {
		next := inst.Next()
		return followSwitchChain(next, chosen)
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
	return followSwitchChain(inst, chosen-1)
}
