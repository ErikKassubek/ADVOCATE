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
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

func ParseCall(inst *s_ssa.InstructionCall, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	f := inst.GetFunc(data.Ssa())

	parseCallParameter(inst.Instruction(), inst, rout, rout, f)

	if f != nil {
		blocking.jumpBackPos[rout].Push(inst.Next())
		return s_ssa.NewSsaPosFunc(f), nil
	}

	if elem != nil && !elem.Committed() {
		return nil, nil
	}

	return inst.Next(), nil

}

func parseCallParameter(call ssa.CallInstruction, inst *s_ssa.InstructionCall, routCall int, routFunc int, f *s_ssa.Function) {
	if f == nil {
		return
	}

	paras := make(map[string]*instructionWithInfo)

	if call != nil {
		for i, param := range f.Params() {
			if sharesUnderlyingResource(param.Type()) {
				arg := call.Common().Args[i]
				var argStr string
				switch arg.(type) {
				case *ssa.Alloc:
					argStr = arg.Name()
				default:
					argStr = arg.String()
				}
				d := getDecOfSSAVar(routCall, argStr)
				paras[param.Name()] = d
			} else {
				paras[param.Name()] = nil
			}
		}
	}

	blocking.NewFuncStack(routFunc, inst)

	for p, i := range paras {
		if i == nil {
			addPathParam(routFunc, p, nil)
		} else {
			addPathParam(routFunc, p, i.Resource)
		}
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
