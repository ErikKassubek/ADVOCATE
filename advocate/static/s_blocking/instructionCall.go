// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionCall.go
// Brief: Call Instruciton
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

func ParseCall(inst *s_ssa.InstructionCall, rout int, _ trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoCall(inst, rout, nil)

	f := inst.GetFunc(data.Ssa())
	if f != nil {
		blocking.JumpBackPos[rout].Push(inst.Next())
		return s_ssa.NewSsaPosFunc(f), info
	}

	return inst.Next(), info

}

func instInfoCall(inst *s_ssa.InstructionCall, rout int, elem trace.Element) *instructionWithInfo {
	log.Todo("InstructionCall NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}
