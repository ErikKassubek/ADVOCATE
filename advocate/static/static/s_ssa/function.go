// Copyright (c) 2026 Erik Kassubek
//
// File: function.go
// Brief: Function
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"fmt"

	"golang.org/x/tools/go/ssa"
)

type Function struct {
	name   string
	pkg    string
	loc    string
	blocks []*Block
	para   []*ssa.Parameter
	fv     []*ssa.FreeVar
}

func (this *Function) Name() string {
	return this.name
}

func (this *Function) string() string {
	res := fmt.Sprintf("Name: %s\n", this.name)

	res += "Parameters:\n"
	if this.para != nil {
		for i, fv := range this.para {
			res += fmt.Sprintf("\t%d:\t%s\n", i, fv.Name())
		}
	} else {
		res += "\t-\n"
	}

	res += "Free variables:\n"
	if this.fv != nil {
		for i, fv := range this.fv {
			res += fmt.Sprintf("\t%d:\t%s\n", i, fv.Name())
		}
	} else {
		res += "\t-\n"
	}

	res += "Func:\n"

	for _, block := range this.blocks {
		res += "\t" + block.String()
	}

	res += "\n --------------------------------------------------\n"

	return res
}

func (this *Data) analysisFunction(fn *ssa.Function) Function {
	f := Function{
		name:   fn.String(),
		pkg:    fn.Pkg.Pkg.Path(),
		loc:    this.ast.GetPosFromPos(fn.Pos()),
		blocks: make([]*Block, len(fn.Blocks)),
	}

	for i, block := range fn.Blocks {
		f.blocks[i] = this.analysisBlock(&f, block)
	}

	f.para = fn.Params
	f.fv = fn.FreeVars

	return f
}

func (this *Data) funcFromName(name string) *Function {
	return this.Funcs()[name]
}

func isMain(fn *ssa.Function) bool {
	return fn.Name() == "main" &&
		fn.Signature.Recv() == nil &&
		fn.Pkg != nil &&
		fn.Pkg.Pkg.Name() == "main"
}

func isInit(fn *ssa.Function) bool {
	return fn.Name() == "init" &&
		fn.Signature.Recv() == nil &&
		fn.Pkg != nil &&
		fn.Pkg.Pkg.Name() == "main"
}

func (this *Function) Blocks() []*Block {
	return this.blocks
}

func (this *Function) Params() []*ssa.Parameter {
	return this.para
}

func (this *Function) FreeVar() []*ssa.FreeVar {
	return this.fv
}
