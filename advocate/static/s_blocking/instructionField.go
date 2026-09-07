// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionField.go
// Brief: Field Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"strconv"
	"strings"
)

func ParseField(inst *s_ssa.InstructionField, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	field_name, field_index := getFieldInfo(inst)

	iwi := getDecOfSSAVar(rout, field_name)

	iwi_new := newIwiFromIwiIndex(inst, iwi, field_index)

	setReference(iwi, field_index, iwi_new, 0)

	info := addPathInstr(rout, iwi_new)
	return inst.Next(), info
}

func getFieldInfo(inst s_ssa.Instruction) (string, int) {
	term := strings.ReplaceAll(inst.Term(), "&", "")
	fields := strings.Split(term, " ")
	if len(fields) != 2 {
		log.Errorf("Invalid Field Term: %s", inst.Term())
		return "", 0
	}

	name := strings.Split(fields[0], ".")[0]

	index_str := strings.TrimSuffix(strings.TrimPrefix(fields[1], "[#"), "]")

	index, err := strconv.Atoi(index_str)
	if err != nil {
		log.Errorf("Invalid Field Index %s", index_str)
	}

	return name, index
}
