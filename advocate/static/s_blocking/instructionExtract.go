// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionExtrace.go
// Brief: Extract Instruction
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

func instInfoExtract(inst *s_ssa.InstructionExtract, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionExtract NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseExtract(inst *s_ssa.InstructionExtract, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoExtract(inst, rout, elem)
	return inst.Next(), info
}
