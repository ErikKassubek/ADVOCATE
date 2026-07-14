// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the ssa analysis
//
// Author: Erik Kassubek
// Created: 2026-06-30
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

	funcs    []*Function
	mainFunc *Function
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

// ================================================================
// Alias
// ================================================================

type aliasResult int

const (
	never aliasResult = iota
	sometimes
	always
)

func (self *aliasResult) string() string {
	switch *self {
	case never:
		return "never"
	case sometimes:
		return "sometimes"
	case always:
		return "always"
	default:
		return "unknown"
	}
}
