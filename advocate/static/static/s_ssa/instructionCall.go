// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionCall.go
// Brief: Call Instruciton
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionCall struct {
	InstructionBase

	funcName string
	f        *Function
}

func NewCall(f *Function, inst *ssa.Call, i int) *InstructionCall {
	name := ""
	if callee := inst.Common().StaticCallee(); callee != nil {
		name = callee.String()
	}

	return &InstructionCall{InstructionBase: newInstructionBase(f, Ic_call, inst, i),
		funcName: name}
}

func (this *InstructionCall) GetFunc(data *Data) *Function {
	if this.f != nil {
		return this.f
	}

	for _, fu := range data.funcs {
		if this.funcName == fu.name {
			this.f = fu
			return fu
		}
	}

	return nil
}

func (this *InstructionCall) Instruction() *ssa.Call {
	return this.inst.(*ssa.Call)
}

func (this *InstructionCall) setRelevant(data *Data) {
	resource := this.Conc().Resource()

	i := this.Inst().(*ssa.Call)
	fn := i.Common().StaticCallee()

	if fn == nil {
		this.relevant = resource
		this.inTrace = resource
		return
	}

	if _, ok := data.funcs[fn.String()]; !ok {
		this.relevant = resource
		this.inTrace = resource
		return
	}

	this.relevant = true
	this.inTrace = true
}
