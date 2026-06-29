// Copyright (c) 2026 Erik Kassubek
//
// File: record.go
// Brief: Record data
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package static

import (
	"fmt"
	"go/ast"
)

// TODO: call is not recorded if in funcLit
func (self *staticData) recordFunctionCall(fdecl *ast.FuncDecl, call *ast.CallExpr) {
	// prevent function from calling itself, if it is not recursive
	if self.getPos(call) == self.getPos(fdecl) {
		return
	}

	name := self.getName(call)
	if name == "FuncLit" {
		return
	}

	funcDecl := self.getFuncDecl(call)
	info := self.funcInfo[fdecl]
	info.funcCalls[funcCall{call, funcDecl, name, self.getCallType(call, funcDecl)}] = struct{}{}
	self.funcInfo[fdecl] = info
}

func (self *staticData) recordOperation(f *ast.FuncDecl, expr ast.Expr, name funcName) {
	info := self.funcInfo[f]

	if info.ops == nil {
		info.ops = make(map[operation]map[ast.Expr]struct{})
	}

	// TODO: object ids
	op := operation{
		1,
		name,
	}

	if _, ok := info.ops[op]; !ok {
		info.ops[op] = make(map[ast.Expr]struct{})
	}

	info.ops[op][expr] = struct{}{}

	self.funcInfo[f] = info
}

// TODO: go mu.Lock() and similar does not work yet
func (self *staticData) recordGoStatement(fdecl *ast.FuncDecl, call *ast.GoStmt) {
	info := self.funcInfo[fdecl]

	funcDecl := self.resolveGoFunc(call)

	// handle func lit
	if funcDecl == nil {
		funcDecl = self.recordFuncLitGo(fdecl, call)
		if funcDecl == nil {
			return
		}
	}

	info.goCalls[call] = funcDecl
	self.funcInfo[fdecl] = info

	self.routFunc[call] = fdecl
}

func (self *staticData) recordFuncLitGo(
	fdecl *ast.FuncDecl,
	goStmt *ast.GoStmt,
) *ast.FuncDecl {

	if _, ok := self.funcInfo[fdecl].goCalls[goStmt]; ok {
		return nil
	}

	call := goStmt.Call

	lit, ok := call.Fun.(*ast.FuncLit)
	if !ok {
		return nil
	}

	// create dummy declaration
	decl := &ast.FuncDecl{
		Name: &ast.Ident{
			NamePos: lit.Pos(),
			Name:    fmt.Sprintf("<func-lit-%d>", self.nextFuncLitId),
		},
		Type: lit.Type,
		Body: lit.Body,
	}
	self.npm[decl.Name] = self.npm[fdecl]

	self.nextFuncLitId++

	self.funcLitDecl[lit] = decl

	self.funcInfo[fdecl].goCalls[goStmt] = decl

	self.addFuncIfNotExists(decl)

	// Analyze the function literal just like any other function.
	self.detOpsInFunc(decl)

	return decl
}
