// Copyright (c) 2026 Erik Kassubek
//
// File: connect.go
// Brief: Find elements from the trace or code in the SSA and vice versa
//
// Author: Erik Kassubek
// Created: 2026-05-07
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"
	"strings"
)

// TODO: global

func (this *Data) TraceToSSA(elem trace.Element) (*Function, *Instruction) {
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

	// TODO: globals/init

	name := this.elemNameToSSAName(f)

	for _, fu := range this.funcs {
		if name == fu.name {
			return fu
		}
	}

	return nil
}

func (this *Data) elemNameToSSAName(elem *trace.ElementFunc) string {
	name := elem.GetName()

	if strings.HasPrefix(name, "main") {
		name = strings.TrimPrefix(name, "main")
		projName := strings.Split(this.mainFunc.name, ".")
		name = projName[0] + name
	}

	return name
}

func (this *Data) FindSSAInstrFromSSAFunc(f *Function, elem trace.Element) *Instruction {
	for _, b := range f.blocks {
		for _, i := range b.insts {
			if this.correspondsSSAInstrTraceElem(&i, elem) {
				return &i
			}
		}
	}

	return nil
}

func (this *Data) correspondsSSAInstrTraceElem(i *Instruction, elem trace.Element) bool {
	ssaClass := i.class
	file, line := this.getInstructionPos(i.inst)

	if file != elem.File() || line != elem.Line() {
		return false
	}

	switch e := elem.(type) {
	case *trace.ElementAlloc:
		if ssaClass == ic_alloc {
			if elem.Type(false) == trace.Mutex && i.hasMutex() {
				return true
			} else if elem.Type(false) == trace.Cond && i.hasCond() {
				return true
			} else if elem.Type(false) == trace.Wait && i.hasWg() {
				return true
			}
		}
		if ssaClass == ic_makeChan && elem.Type(true) == trace.NewChannel && i.hasChannel() {
			return true
		}
	case *trace.ElementChannel:
		if i.hasChannel() && (ssaClass == ic_send || ssaClass == ic_unOp || ssaClass == ic_call) {
			return true
		}
	case *trace.ElementMutex:
		if i.hasMutex() && ssaClass == ic_call {
			return true
		}
	case *trace.ElementCond:
		if i.hasCond() && ssaClass == ic_call {
			return true
		}
	case *trace.ElementWait:
		if i.hasWg() && ssaClass == ic_call {
			return true
		}
	case *trace.ElementSelect:
		if i.hasChannel() && ssaClass == ic_select {
			return true
		}
	case *trace.ElementFunc:
		if ssaClass == ic_call && i.name == e.GetName() {
			return true
		}
	}

	return false
}
