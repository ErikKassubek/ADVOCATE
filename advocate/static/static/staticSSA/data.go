// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the ssa analysis
//
// Author: Erik Kassubek
// Created: 2026-06-30
//
// License: BSD-3-Clause

package staticSSA

import (
	"advocate/static/static/staticAST"

	"golang.org/x/tools/go/ssa"
)

type Data struct {
	ast *staticAST.Data

	ssa      *ssa.Program // static single assignment (intermediate program representation where each variable is assigned exactly once)
	ssaPkgs  []*ssa.Package
	ssaMains []*ssa.Package
}

func BuildSsa(ast *staticAST.Data) *Data {
	data := &Data{
		ast: ast,
	}

	data.buildSsa(ast.Pkgs)

	return data
}
