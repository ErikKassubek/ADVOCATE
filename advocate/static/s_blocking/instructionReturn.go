// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionReturn.go
// Brief: Return Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
)

func instInfoReturn(inst *s_ssa.InstructionReturn, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionReturn NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseReturn(inst *s_ssa.InstructionReturn, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoReturn(inst, rout, elem)

	i := blocking.JumpBackPos[rout].Pop()

	return i, info
}
