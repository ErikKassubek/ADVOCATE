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

	// TODO: return parents

	retInfo := make([]*instructionWithInfo, len(retSSAVar))
	for i, v := range retSSAVar {
		ret := getDecOfSSAVar(rout, v.Name())
		if ret != nil {
			retInfo[i] = ret
		}
	}

	if len(retInfo) == 0 {
		return nil
	}

	retVar := blocking.ReturnStack(rout)
	iwi := newIWI2(retVar)
	for _, ret := range retInfo {
		iwi = iwi.Merge(ret)
	}
	return addPathInstr(rout, iwi)
}

func ParseReturn(inst *s_ssa.InstructionReturn, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoReturn(inst, rout, elem)

	i := blocking.jumpBackPos[rout].Pop()

	// if i != nil {
	// 	log.Debug("JUMP BACK: ", i.StringInfo())
	// }
	return i, info
}
