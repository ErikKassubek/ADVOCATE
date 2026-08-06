// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionReturn.go
// Brief: Return Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

func instInfoReturn(inst *s_ssa.InstructionReturn, rout int, _ trace.Element) *instructionWithInfo {
	retSSAVar := inst.Instruction().Results

	retInfo := make([]map[*trace.Resource]struct{}, len(retSSAVar))
	for i, v := range retSSAVar {
		retInfo[i] = getDecOfSSAVar(rout, v.Name()).Resource[0]
	}

	retVar := blocking.ReturnStack(rout)
	if retVar != nil {
		return addPathInstr(rout, retVar, retInfo)
	}

	return nil
}

func ParseReturn(inst *s_ssa.InstructionReturn, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoReturn(inst, rout, elem)

	i := blocking.jumpBackPos[rout].Pop()

	return i, info
}
