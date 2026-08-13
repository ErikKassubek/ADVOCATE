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
	"advocate/utils/types"
)

// TODO: handle case where elem is nil/all case

func instInfoSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element) *instructionWithInfo {
	iwi := getDecOfSSAVar(rout, inst.Instruction().Chan.Name())

	if elem != nil {
		if _, ok := blocking.chanBuffer[elem.ResourceID()]; !ok {
			blocking.chanBuffer[elem.ResourceID()] = types.NewStack[*instructionWithInfo]()
		}
		blocking.chanBuffer[elem.ResourceID()].Push(iwi)
	}

	if iwi != nil {
		return addPathInstr(rout, inst, iwi.Resource)
	}
	return addPathInstr(rout, inst, nil)
}

func ParseSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoSend(inst, rout, elem)

	if elem != nil && !elem.Committed() {
		return nil, info
	}

	return inst.Next(), info
}
