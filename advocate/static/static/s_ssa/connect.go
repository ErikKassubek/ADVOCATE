// Copyright (c) 2026 Erik Kassubek
//
// File: connect.go
// Brief: Find elements from the trace or code in the SSA and vice versa
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"
	"strings"
)

func (this *Data) TraceToSSA(elem trace.Element) (*Function, Instruction) {
	// we are only interested in blocking types and alloc
	switch elem.(type) {
	case *trace.ElementAlloc, *trace.ElementChannel, *trace.ElementMutex, *trace.ElementCond, *trace.ElementWait, *trace.ElementSelect, *trace.ElementFunc:
	default:
		return nil, nil
	}

	f := elem.Function()

	ssaF := this.TraceFuncToSSAFunc(f)

	return ssaF, this.FindSSAInstrFromSSAFunc(ssaF, elem)
}

func (this *Data) TraceFuncToSSAFunc(f *trace.ElementFunc) *Function {
	if f == nil {
		return this.mainFunc
	}

	name := this.elemNameToSSAName(f)

	for _, fu := range this.funcs {
		if name == fu.name {
			return fu
		}
	}

	return nil
}

func (this *Data) elemNameToSSAName(elem *trace.ElementFunc) string {
	name := elem.Name()

	if strings.HasPrefix(name, "main") {
		name = strings.TrimPrefix(name, "main")
		projName := strings.Split(this.mainFunc.name, ".")
		name = projName[0] + name
	}

	return name
}

func (this *Data) FindSSAInstrFromSSAFunc(f *Function, elem trace.Element) Instruction {
	for _, b := range f.blocks {
		for _, i := range b.insts {
			if this.correspondsSSAInstrTraceElem(i, elem) {
				return i
			}
		}
	}

	return nil
}

func (this *Data) correspondsSSAInstrTraceElem(i Instruction, elem trace.Element) bool {
	ssaClass := i.Class()
	file, line := this.getInstructionPos(i.Inst())

	if file != elem.File() || line != elem.Line() {
		return false
	}

	switch e := elem.(type) {
	case *trace.ElementAlloc:
		if ssaClass == Ic_alloc {
			if elem.Type(true) == trace.NewMutex && i.HasMutex() {
				return true
			} else if elem.Type(true) == trace.NewCond && i.HasCond() {
				return true
			} else if elem.Type(true) == trace.NewWait && i.HasWG() {
				return true
			}
		}
		if ssaClass == Ic_makeChan && elem.Type(true) == trace.NewChannel && i.HasChannel() {
			return true
		}
		if ssaClass == Ic_call {
			return true
		}
	case *trace.ElementChannel:
		if i.HasChannel() && (ssaClass == Ic_send || ssaClass == Ic_unOp || ssaClass == Ic_call) {
			return true
		}
	case *trace.ElementMutex:
		if i.HasMutex() && ssaClass == Ic_call {
			return true
		}
	case *trace.ElementCond:
		if i.HasCond() && ssaClass == Ic_call {
			return true
		}
	case *trace.ElementWait:
		if i.HasWG() && ssaClass == Ic_call {
			return true
		}
	case *trace.ElementSelect:
		if i.HasChannel() && ssaClass == Ic_select {
			return true
		}
	case *trace.ElementFunc:
		if ssaClass == Ic_call && i.Variable() == e.Name() {
			return true
		}
	case *trace.ElementFork:
		if ssaClass == Ic_go {
			return true
		}
	}

	return false
}
