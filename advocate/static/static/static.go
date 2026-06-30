// Copyright (c) 2026 Erik Kassubek
//
// File: static.go
// Brief: Top Level Package for static analysis
//
// Author: Erik Kassubek
// Created: 2026-06-30
//
// License: BSD-3-Clause

package static

import (
	"advocate/static/static/staticAST"
	"advocate/static/static/staticSSA"
)

type Data struct { // always use buildStaticData, never staticData{}
	ast *staticAST.Data
	ssa *staticSSA.Data
}

func BuildStaticData(dir string) (*Data, error) {
	ast, err := staticAST.BuildAst(dir)
	if err != nil {
		return nil, err
	}

	ssa := staticSSA.BuildSsa(ast)

	data := &Data{
		ast: ast,
		ssa: ssa,
	}

	return data, nil
}

func (self *Data) Ast() *staticAST.Data {
	return self.ast
}

func (self *Data) Ssa() *staticSSA.Data {
	return self.ssa
}
