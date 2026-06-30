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
	"advocate/static/static/s_alias"
	"advocate/static/static/s_ast"
	"advocate/static/static/s_ssa"
)

type Data struct { // always use buildStaticData, never staticData{}
	ast   *s_ast.Data
	ssa   *s_ssa.Data
	alias *s_alias.Data
}

func BuildStaticData(dir string) (*Data, error) {
	ast, err := s_ast.BuildAst(dir)
	if err != nil {
		return nil, err
	}

	ssa := s_ssa.BuildSsa(ast)

	al := s_alias.RunAliasAnalysis(ssa)

	data := &Data{
		ast:   ast,
		ssa:   ssa,
		alias: al,
	}

	return data, nil
}

func (self *Data) Ast() *s_ast.Data {
	return self.ast
}

func (self *Data) Ssa() *s_ssa.Data {
	return self.ssa
}
