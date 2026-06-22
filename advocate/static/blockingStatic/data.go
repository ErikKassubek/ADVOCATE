// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blockingStatic

import (
	"advocate/utils/log"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

type funcCall struct {
	call *ast.CallExpr
	decl *ast.FuncDecl
}

type staticData struct { // always use buildStaticData, never staticData{}
	dir string // path to analyzed program

	pkgs []*packages.Package

	fst *token.FileSet

	pkgInfo map[*packages.Package]*types.Info
	uses    map[*ast.Ident]types.Object
	defs    map[*ast.Ident]types.Object

	astMap map[string][]*ast.File         // pkg path -> files
	ast    []*ast.File                    // flattened list
	npm    map[ast.Node]*packages.Package // node packages map

	ssa      *ssa.Program // static single assignment (intermediate program representation where each variable is assigned exactly once)
	ssaPkgs  []*ssa.Package
	ssaMains []*ssa.Package

	funcDeclMap map[token.Pos]*ast.FuncDecl

	// operations per function: func -> vartiable id (TODO: change to point to variable and not expression) -> funcs
	opsPerFunk map[*ast.FuncDecl]map[*ast.Expr]map[funcName]struct{}

	// functions called in each function,
	// ast.Expr.(type) -> *ast.Ident: direct function (foo())
	// ast.Expr.(type) -> *ast.SelectorExpr: methodCall (obj.Method())
	// ast.Expr.(type) -> *ast.FuncLit: function literal (func() {...}())
	funcsPerFunc map[*ast.FuncDecl][]funcCall

	// routine spawns from functions
	// *ast.GoStmt.Call.(type) -> *ast.Ident: direct function (go foo())
	// *ast.GoStmt.Call.(type) -> *ast.SelectorExpr: methodCall (go obj.Method())
	// *ast.GoStmt.Call.(type) -> *ast.FuncLit: function literal (go func() { ... }())
	goStatementPerFunc map[*ast.FuncDecl][]*ast.GoStmt
}

func buildStaticData(dir string) (*staticData, error) {
	data := &staticData{
		dir: dir,
		fst: token.NewFileSet(),

		astMap: make(map[string][]*ast.File),
		ast:    make([]*ast.File, 0),
		npm:    make(map[ast.Node]*packages.Package),

		funcDeclMap:        make(map[token.Pos]*ast.FuncDecl),
		opsPerFunk:         make(map[*ast.FuncDecl]map[*ast.Expr]map[funcName]struct{}),
		funcsPerFunc:       make(map[*ast.FuncDecl][]funcCall),
		goStatementPerFunc: make(map[*ast.FuncDecl][]*ast.GoStmt),
	}
	err := data.loadPackages()
	if err != nil {
		log.Error(err.Error())
		return data, err
	}
	data.buildTypeInfo()
	data.buildAst()
	// must be called afer load packages
	data.buildSsa()
	// data.runPointerAnalysis()

	return data, nil
}

func (self staticData) callName(call funcCall) string {
	res := self.getName(call.call) + ":"
	if call.decl == nil {
		res += self.getConcFuncName(call.call)
	} else {
		res += self.getName(call.decl.Name)
	}

	return res
}
