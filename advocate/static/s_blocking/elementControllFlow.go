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

func followIfChain(pos s_ssa.SsaPos, inst s_ssa.Instruction, chosen int) s_ssa.SsaPos {
	instrIf, ok := inst.(*s_ssa.InstructionIf)
	if !ok {
		return pos
	}

	if chosen == 0 {
		pos.NewBlock(pos.Blocks()[instrIf.If()])
		return pos
	}

	pos.NewBlock(pos.Blocks()[instrIf.Else()])
	return followIfChain(pos, pos.I, chosen-1)
}

func followSwitchChain(pos s_ssa.SsaPos, inst s_ssa.Instruction, chosen int) s_ssa.SsaPos {
	if _, ok := inst.(*s_ssa.InstructionBinOp); ok {
		next := pos.Next()
		return followSwitchChain(next, next.I, chosen)
	}

	instrIf, ok := inst.(*s_ssa.InstructionIf)
	if !ok {
		return pos
	}

	if chosen == 0 {
		pos.NewBlock(pos.Blocks()[instrIf.If()])
		return pos
	}

	pos.NewBlock(pos.Blocks()[instrIf.Else()])
	return followSwitchChain(pos, pos.I, chosen-1)
}
