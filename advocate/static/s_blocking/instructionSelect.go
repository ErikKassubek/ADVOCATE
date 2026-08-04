// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSelect.go
// Brief: Select Instruction
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

func instInfoSelect(inst *s_ssa.InstructionSelect, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionSelect NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseSelect(inst *s_ssa.InstructionSelect, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoSelect(inst, rout, elem)

	log.Todo("InstructionSelect NOT IMPLEMENTED YET")

	return inst.Next(), info
}
