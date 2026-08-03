// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeInterface.go
// Brief: Make interface Instruction
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

func instInfoMakeInterface(inst *s_ssa.InstructionMakeInterface, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionMakeInterface NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseMakeInterface(inst *s_ssa.InstructionMakeInterface, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMakeInterface(inst, rout, elem)
	return inst.Next(), info
}
