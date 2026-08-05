// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionUnOp.go
// Brief: Unary instruction Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"strings"
)

func instInfoUnOp(inst *s_ssa.InstructionUnOp, rout int, _ trace.Element) *instructionWithInfo {
	term := inst.Term()
	if strings.HasPrefix(term, "*") && !strings.Contains(term, " ") {
		if rout == 3 {
			log.Debug(inst.StringInfo())
		}
		ssaVar := findDecOfSSAVar(rout, term)
		return addPathInstr(rout, inst, ssaVar.Resource)
	}

	return nil
}

func ParseUnOp(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoUnOp(inst, rout, elem)

	return inst.Next(), info
}
