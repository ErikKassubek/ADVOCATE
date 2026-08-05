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
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

func ParseCall(inst *s_ssa.InstructionCall, rout int, _ trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	f := inst.GetFunc(data.Ssa())

	parseCallParameter(inst.Instruction(), rout, rout, f, inst.Variable())

	info := instInfoCall(inst, rout, nil)
	if f != nil {
		blocking.jumpBackPos[rout].Push(inst.Next())
		return s_ssa.NewSsaPosFunc(f), info
	}

	return inst.Next(), info

}

func parseCallParameter(inst ssa.CallInstruction, routCall int, routFunc int, f *s_ssa.Function, retVarName string) {
	if f == nil {
		return
	}

	paras := make(map[string]*instructionWithInfo)

	if inst != nil {
		for i, param := range f.Params() {
			if !sharesUnderlyingResource(param.Type()) {
				paras[param.Name()] = nil
			} else {
				arg := inst.Common().Args[i].String()
				d := getDecOfSSAVar(routCall, arg)
				paras[arg] = d
			}
		}
	}

	blocking.NewFuncStack(routFunc, retVarName)

	for p, i := range paras {
		addPathParam(routFunc, p, i.Resource)
	}

	if routCall != routFunc { // fork
		info := blocking.lastClosure[routCall]

		fv := f.FreeVar()

		if len(fv) == 0 {
			return
		}

		if len(info) != len(fv) {
			panic(fmt.Sprintf("Invalid length of free var at %s: %d != %d", f.Name(), len(info), len(fv)))
		}

		for i := 0; i < len(info); i++ {
			addPathParam(routFunc, fv[i].Name(), info[i].Resource)
		}
	}
}

func instInfoCall(inst *s_ssa.InstructionCall, rout int, elem trace.Element) *instructionWithInfo {
	log.Todo("InstructionCall NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func sharesUnderlyingResource(t types.Type) bool {
	switch t := t.Underlying().(type) {
	case *types.Pointer:
		return true

	case *types.Slice:
		return true

	case *types.Map:
		return true

	case *types.Chan:
		return true

	case *types.Signature:
		// Function values are descriptors to code/closure.
		return true

	case *types.Interface:
		// An interface copy shares the underlying dynamic value.
		return true

	case *types.Basic:
		return false

	case *types.Array:
		return false

	case *types.Struct:
		// A struct copy duplicates all of its fields.
		// Whether the copied struct still aliases depends on its fields.
		for i := 0; i < t.NumFields(); i++ {
			if sharesUnderlyingResource(t.Field(i).Type()) {
				return true
			}
		}
		return false

	case *types.Named:
		// Normally unreachable because of Underlying(),
		// but included for completeness.
		return sharesUnderlyingResource(t.Underlying())

	default:
		return false
	}
}
