// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionPhi.go
// Brief: Phi Instruction
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

func instInfoPhi(inst *s_ssa.InstructionPhi, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionPhi NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParsePhi(inst *s_ssa.InstructionPhi, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoPhi(inst, rout, elem)
	return inst.Next(), info
}
