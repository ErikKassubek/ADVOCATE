// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSend.go
// Brief: Send Instruction
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

func instInfoSend(inst *s_ssa.InstructionSend, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionSend NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoSend(inst, rout, elem)

	return inst.Next(), info
}
