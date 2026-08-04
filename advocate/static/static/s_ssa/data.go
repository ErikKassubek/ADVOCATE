// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the ssa analysis
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/static/static/s_ast"

	"golang.org/x/tools/go/ssa"
)

type Data struct {
	ast *s_ast.Data

	ssa      *ssa.Program // static single assignment (intermediate program representation where each variable is assigned exactly once)
	ssaPkgs  []*ssa.Package
	ssaMains []*ssa.Package

	funcs    map[string]*Function
	mainFunc *Function
	initFunc *Function
	alloc    map[*Instruction][]Instruction // instruction -> set of alloc
}

func BuildSsa(ast *s_ast.Data) *Data {
	data := &Data{
		ast: ast,
	}

	data.buildSsa(ast.Pkgs)

	data.runSSAAnalysis()

	return data
}

func (this *Data) Funcs() map[string]*Function {
	return this.funcs
}

func (this *Data) MainFunc() *Function {
	return this.mainFunc
}

func (this *Data) InitFunc() *Function {
	return this.initFunc
}
