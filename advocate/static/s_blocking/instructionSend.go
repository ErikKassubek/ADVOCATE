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
	"advocate/utils/log"
	"advocate/utils/types"
)

func instInfoSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element) *instructionWithInfo {
	log.Debug("SEND: ", elem.ObjID())
	if _, ok := blocking.chanBuffer[elem.ObjID()]; !ok {
		blocking.chanBuffer[elem.ObjID()] = types.NewStack[*instructionWithInfo]()
	}

	d := getDecOfSSAVar(rout, inst.Instruction().X.Name())
	blocking.chanBuffer[elem.ObjID()].Push(d)

	return addPathInstr(rout, inst, nil)
}

func ParseSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoSend(inst, rout, elem)

	return inst.Next(), info
}
