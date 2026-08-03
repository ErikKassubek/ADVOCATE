// Copyright (c) 2026 Erik Kassubek
//
// File: static.go
// Brief: Top Level Package for static analysis
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package static

import (
	"advocate/advoc/toolchain"
	"advocate/static/static/s_ast"
	"advocate/static/static/s_ssa"
)

type Data struct { // always use buildStaticData, never staticData{}
	ast *s_ast.Data
	ssa *s_ssa.Data
}

func BuildStaticData(dir string) (*Data, error) {
	file, line := toolchain.ImportInsertStatic()

	ast, err := s_ast.BuildAst(dir)
	if err != nil {
		toolchain.ImportRemoveStatic(file, line)
		return nil, err
	}

	ssa := s_ssa.BuildSsa(ast)

	data := &Data{
		ast: ast,
		ssa: ssa,
	}

	toolchain.ImportRemoveStatic(file, line)
	return data, nil
}

func (self *Data) Ast() *s_ast.Data {
	return self.ast
}

func (self *Data) Ssa() *s_ssa.Data {
	return self.ssa
}

func (self *Data) Blocking() *s_ssa.BlockingData {
	return self.Ssa().Blocking
}
