// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionFieldAddr.go
// Brief: Field Address Instruction
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

func ParseFieldAddr(inst *s_ssa.InstructionFieldAddr, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	field_name, field_index := getFieldInfo(inst)

	log.Debug2(inst)
	log.Debug2(field_name)
	iwi := getDecOfSSAVar(rout, field_name)

	log.Debug2("IWI: ", iwi)

	if len(iwi.Reference[0]) != 0 {
		log.Debug2(iwi.Reference[0][0])
	}

	iwi_new := newIwiFromIwiIndex(inst, iwi, field_index)
	setReference(iwi, field_index, iwi_new, 0)

	info := addPathInstr(rout, iwi_new)
	return inst.Next(), info
}
