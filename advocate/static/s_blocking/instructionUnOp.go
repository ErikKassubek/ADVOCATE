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
	"go/token"
)

func instInfoPointerDereference(inst *s_ssa.InstructionUnOp, rout int) *instructionWithInfo {
	term := inst.Term()
	ssaVar := getDecOfSSAVar(rout, term)
	return addPathInstr(rout, inst, ssaVar.Resource)
}

func instInfoReceive(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element, forward bool) *instructionWithInfo {
	var receivedValue *instructionWithInfo
	if elem != nil {
		receivedValue = blocking.chanBuffer[elem.ResourceID()].Pop()
	}

	if forward {
		// TODO: implement
		// log.Debug("FORWARD RECV")
		// for _, res := range sendForward {
		// 	l, err := code.GetLineContent(pos)
		// 	if err != nil {
		// 		log.Error(err)
		// 	} else {
		// 		log.Debug2("Pos: ", l)
		// 	}
		// }
	}

	if receivedValue == nil {
		return addPathInstr(rout, inst, nil)
	}
	return addPathInstr(rout, inst, receivedValue.Resource)
}

func instInfoUnOp(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element, forward bool) *instructionWithInfo {
	switch inst.Instruction().Op {
	case token.MUL: // pointer dereference
		return instInfoPointerDereference(inst, rout)
	case token.ARROW: // channel receive
		return instInfoReceive(inst, rout, elem, forward)
	}

	return nil
}

func ParseUnOp(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element, forward bool) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoUnOp(inst, rout, elem, forward)

	return inst.Next(), info
}
