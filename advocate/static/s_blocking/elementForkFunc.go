// Copyright (c) 2026 Erik Kassubek
//
// File: elementForkFunc.go
// Brief: Parse a fork/go and func call operation
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"fmt"

	"golang.org/x/tools/go/ssa"
)

func parseGo(rout int, elem trace.Element, inst *s_ssa.InstructionGo) *s_ssa.Function {
	v := inst.Inst().(*ssa.Go)

	var fName string

	switch v := v.Call.Value.(type) {
	case *ssa.MakeClosure:
		fName = v.Fn.(*ssa.Function).String()

	case *ssa.Function:
		fName = v.String()
	}

	f := getSSAFuncFromName(fName)

	nextPerRout[elem.ObjID()] = s_ssa.NewSsaPosFunc(f)

	return f
}

func parseNewFunc(rout int, f *s_ssa.Function) {
	info := lastClosure[rout]
	fv := f.FreeVar()
	if len(info) != len(fv) {
		panic(fmt.Sprintf("Invalid length of free var: %d != %d", len(info), len(fv)))
	}

	for i := 0; i < len(info); i++ {
		addPathParam(rout, fv[i].Name(), info[i].resource)
	}
}
