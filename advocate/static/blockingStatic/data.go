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
	"advocate/utils/flags"
	"advocate/utils/log"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

type funcCall struct {
	call     *ast.CallExpr
	decl     *ast.FuncDecl
	name     string // TODO: multi package
	callType funcName
}

type funcInfo struct {
	decl *ast.FuncDecl

	// functions called in each function,
	// func -> call
	// ast.Expr.(type) -> *ast.Ident: direct function (foo())
	// ast.Expr.(type) -> *ast.SelectorExpr: methodCall (obj.Method())
	// ast.Expr.(type) -> *ast.FuncLit: function literal (func() {...}())
	funcCalls map[*ast.FuncDecl]funcCall

	ops map[ast.Expr]map[funcName]struct{}

	// routine spawns from functions
	// *ast.GoStmt.Call.(type) -> *ast.Ident: direct function (go foo())
	// *ast.GoStmt.Call.(type) -> *ast.SelectorExpr: methodCall (go obj.Method())
	// *ast.GoStmt.Call.(type) -> *ast.FuncLit: function literal (go func() { ... }())
	goCalls map[*ast.GoStmt]*ast.FuncDecl
}

type staticData struct { // always use buildStaticData, never staticData{}
	dir string // path to analyzed program

	pkgs []*packages.Package

	fset *token.FileSet

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
	funcsInfo map[*ast.FuncDecl]funcInfo
	routFunc  map[*ast.GoStmt]*ast.FuncDecl
}

func buildStaticData(dir string) (*staticData, error) {
	data := &staticData{
		dir:  dir,
		fset: token.NewFileSet(),

		astMap: make(map[string][]*ast.File),
		ast:    make([]*ast.File, 0),
		npm:    make(map[ast.Node]*packages.Package),

		funcDeclMap: make(map[token.Pos]*ast.FuncDecl),
		funcsInfo:   make(map[*ast.FuncDecl]funcInfo),
		routFunc:    make(map[*ast.GoStmt]*ast.FuncDecl),
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
	data.printSSA(true)
	// data.runPointerAnalysis()

	return data, nil
}

func (self *staticData) getCallType(call *ast.CallExpr, decl *ast.FuncDecl) funcName {
	if decl == nil {
		return self.getConcFuncName(call)
	}

	return unknownFunc
}

func (self *staticData) addFuncIfNotExists(fdecl *ast.FuncDecl) {
	if _, ok := self.funcsInfo[fdecl]; !ok {
		self.funcsInfo[fdecl] = funcInfo{
			decl:      fdecl,
			funcCalls: make(map[*ast.FuncDecl]funcCall, 0),
			ops:       make(map[ast.Expr]map[funcName]struct{}),
			goCalls:   make(map[*ast.GoStmt]*ast.FuncDecl),
		}
	}
}

// (fdecl *ast.FuncDecl, f *ast.CallExpr)
func (self *staticData) recordFunctionCall(fdecl *ast.FuncDecl, call *ast.CallExpr) {
	self.addFuncIfNotExists(fdecl)

	// prevent function from calling itself, if it is not recursive
	if self.getPos(call) == self.getPos(fdecl) {
		return
	}

	funcDecl := self.getFuncDecl(call)
	info := self.funcsInfo[fdecl]
	info.funcCalls[fdecl] = funcCall{call, funcDecl, self.getName(call), self.getCallType(call, funcDecl)}
	self.funcsInfo[fdecl] = info
}

func (self *staticData) recordOperation(f *ast.FuncDecl, expr ast.Expr, name funcName) {
	self.addFuncIfNotExists(f)

	info := self.funcsInfo[f]

	if info.ops == nil {
		info.ops = make(map[ast.Expr]map[funcName]struct{})
	}

	if _, ok := info.ops[expr]; !ok {
		info.ops[expr] = make(map[funcName]struct{})
	}

	info.ops[expr][name] = struct{}{}
	self.funcsInfo[f] = info
}

func (self *staticData) recordGoStatement(fdecl *ast.FuncDecl, call *ast.GoStmt) {
	self.addFuncIfNotExists(fdecl)

	info := self.funcsInfo[fdecl]

	funcDecl := self.resolveGoFunc(call)

	info.goCalls[call] = funcDecl
	self.funcsInfo[fdecl] = info

	self.routFunc[call] = fdecl
}

func (d *staticData) resolveGoFunc(goStmt *ast.GoStmt) *ast.FuncDecl {
	ident, ok := goStmt.Call.Fun.(*ast.Ident)
	if !ok {
		return nil
	}

	// 1. find the package that owns this AST node
	pkg := d.npm[goStmt]
	if pkg == nil {
		return nil
	}

	info := d.pkgInfo[pkg]
	if info == nil {
		return nil
	}

	// 2. resolve identifier → object
	obj := info.ObjectOf(ident)
	if obj == nil {
		return nil
	}

	// 3. ensure it's a function
	fn, ok := obj.(*types.Func)
	if !ok {
		return nil
	}

	// 4. map position → FuncDecl
	pos := fn.Pos()
	if !pos.IsValid() {
		return nil
	}

	return d.funcDeclMap[pos]
}

// TODO: not working
func (self *staticData) getPos(p ast.Node) string {
	pos := p.Pos()
	return self.getPosFromPos(pos)
}

func (self *staticData) getPosFromPos(pos token.Pos) string {
	if !pos.IsValid() {
		return "<invalid position>"
	}

	loc := self.fset.Position(pos)

	if strings.Contains(loc.Filename, ".cache/go-build/") {
		return "[internal]"
	}

	file := strings.TrimPrefix(loc.Filename, flags.ProgPath)

	return fmt.Sprintf("[%s:%d]", file, loc.Line)
}

func (self *staticData) isEqual(p, q ast.Node) bool {
	return self.getPos(p) == self.getPos(q)
}
