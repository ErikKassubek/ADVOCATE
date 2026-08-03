// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionFieldAddr.go
// Brief: Field Address Instruction
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

func instInfoFieldAddr(inst *s_ssa.InstructionFieldAddr, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionFieldAddr NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseFieldAddr(inst *s_ssa.InstructionFieldAddr, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoFieldAddr(inst, rout, elem)
	return inst.Next(), info
}
