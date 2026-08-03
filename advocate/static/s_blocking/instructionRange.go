// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionRange.go
// Brief: Range Instruction
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

func instInfoRange(inst *s_ssa.InstructionRange, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionRange NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseRange(inst *s_ssa.InstructionRange, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoRange(inst, rout, elem)
	return inst.Next(), info
}
