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

	iwi_new := newIwiFromIwiIndex(inst, iwi, field_index)
	iwi.Parents[field_index] = append(iwi.Parents[field_index], iwi_new) // TODO: do this for all addr

	info := addPathInstr(rout, iwi_new)
	return inst.Next(), info
}
