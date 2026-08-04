// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionNext.go
// Brief: Next Instruction
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

func instInfoNext(inst *s_ssa.InstructionNext, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionNext NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseNext(inst *s_ssa.InstructionNext, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoNext(inst, rout, elem)
	return inst.Next(), info
}
