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
	"advocate/static/static/code"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
)

var sendForward = make(map[trace.Resource]map[trace.Resource]struct{})

func instInfoSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element, forward bool) *instructionWithInfo {
	iwi := getDecOfSSAVar(rout, inst.Instruction().Chan.Name())

	if elem != nil {
		if _, ok := blocking.chanBuffer[elem.ResourceID()]; !ok {
			blocking.chanBuffer[elem.ResourceID()] = types.NewStack[*instructionWithInfo]()
		}
		blocking.chanBuffer[elem.ResourceID()].Push(iwi)
	}

	if forward {
		// TODO: make correct
		log.Debug("FORWARD SEND")
		for _, res := range iwi.Resource[0] {
			pos := res.Alloc().Pos()
			l, err := code.GetLineContent(pos)
			if err != nil {
				log.Error(err)
			} else {
				log.Debug2("Pos: ", l)
			}
		}
	}

	if iwi != nil {
		return addPathInstr(rout, inst, iwi.Resource)
	}
	return addPathInstr(rout, inst, nil)
}

func ParseSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element, forward bool) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoSend(inst, rout, elem, forward)

	if elem != nil && !elem.Committed() {
		return nil, info
	}

	return inst.Next(), info
}
