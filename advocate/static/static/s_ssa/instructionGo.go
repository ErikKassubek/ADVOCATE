// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionGo.go
// Brief: Go Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
	"fmt"

	"golang.org/x/tools/go/ssa"
)

type InstructionGo struct {
	InstructionBase
}

func NewGo(f *Function, inst *ssa.Go, i int) *InstructionGo {
	return &InstructionGo{InstructionBase: newInstructionBase(f, Ic_go, inst, i)}
}

func (this *InstructionGo) Instruction() *ssa.Go {
	return this.inst.(*ssa.Go)
}

func (this *InstructionGo) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionGo) AddInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	log.Error("InstructionGo IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionGo) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.AddInstructionWithInfo(data.Blocking, rout, elem)

	v := this.Inst().(*ssa.Go)

	var fName string

	switch v := v.Call.Value.(type) {
	case *ssa.MakeClosure:
		fName = v.Fn.(*ssa.Function).String()

	case *ssa.Function:
		fName = v.String()
	}

	f := GetSSAFuncFromName(data, fName)

	data.Blocking.NextPerRout[elem.ObjID()] = NewSsaPosFunc(f)

	data.Blocking.JumpBackPos[elem.ObjID()] = types.NewStack[Instruction]()

	// we skip the func call in this case. For this case, perform it here
	parseNewFunc(data, rout, elem.ObjID(), f)

	return this.Next(), info
}

func parseNewFunc(data *Data, rout, newRout int, f *Function) {
	info := data.Blocking.LastClosure[rout]
	fv := f.FreeVar()

	if len(fv) == 0 {
		return
	}

	if len(info) != len(fv) {
		panic(fmt.Sprintf("Invalid length of free var at %s: %d != %d", f.Name(), len(info), len(fv)))
	}

	for i := 0; i < len(info); i++ {
		addPathParam(data, newRout, fv[i].Name(), info[i].Resource)
	}
}

func addPathParam(data *Data, rout int, v string, resources map[*trace.Resource]struct{}) *InstructionWithInfo {
	blocking := data.Blocking

	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	newElem := InstructionWithInfo{nil, resources, v}

	blocking.PathPerRoutine[rout] = append(blocking.PathPerRoutine[rout], &newElem)

	return &newElem
}
