// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIndex.go
// Brief: Index Instruction
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

func instInfoIndex(inst *s_ssa.InstructionIndex, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionIndex NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseIndex(inst *s_ssa.InstructionIndex, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoIndex(inst, rout, elem)
	return inst.Next(), info
}
