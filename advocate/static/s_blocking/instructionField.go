// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionField.go
// Brief: Field Instruction
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

func instInfoField(inst *s_ssa.InstructionField, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionField NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseField(inst *s_ssa.InstructionField, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoField(inst, rout, elem)
	return inst.Next(), info
}
