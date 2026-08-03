// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIndexAddr.go
// Brief: Index address Instruction
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

func instInfoIndexAddr(inst *s_ssa.InstructionIndexAddr, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionIndexAddr NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseIndexAddr(inst *s_ssa.InstructionIndexAddr, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoIndexAddr(inst, rout, elem)
	return inst.Next(), info
}
