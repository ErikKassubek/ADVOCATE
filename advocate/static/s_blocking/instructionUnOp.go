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
	"go/token"
)

func instInfoUnOp(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element) *instructionWithInfo {
	switch inst.Instruction().Op {
	case token.MUL: // pointer dereference
		term := inst.Term()
		ssaVar := getDecOfSSAVar(rout, term)
		return addPathInstr(rout, inst, ssaVar.Resource)
	case token.ARROW: // channel receive
		log.Debug("RECV: ", inst.StringInfo(), elem.StringDebug())
		receivedValue := blocking.chanBuffer[elem.ObjID()].Pop()
		return addPathInstr(rout, inst, receivedValue.Resource)
	}

	return nil
}

func ParseUnOp(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoUnOp(inst, rout, elem)

	return inst.Next(), info
}
