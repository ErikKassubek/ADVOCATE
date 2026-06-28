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
	self.addFuncIfNotExists(fdecl)

	// prevent function from calling itself, if it is not recursive
	if self.getPos(call) == self.getPos(fdecl) {
		return
	}

	funcDecl := self.getFuncDecl(call)
	info := self.funcInfo[fdecl]
	info.funcCalls[fdecl] = funcCall{call, funcDecl, self.getName(call), self.getCallType(call, funcDecl)}
	self.funcInfo[fdecl] = info
}

func (self *staticData) recordOperation(f *ast.FuncDecl, expr ast.Expr, name funcName) {
	self.addFuncIfNotExists(f)

	info := self.funcInfo[f]

	if info.ops == nil {
		info.ops = make(map[ast.Expr]map[funcName]struct{})
	}

	if _, ok := info.ops[expr]; !ok {
		info.ops[expr] = make(map[funcName]struct{})
	}

	info.ops[expr][name] = struct{}{}
	self.funcInfo[f] = info
}

// TODO: go mu.Lock() and similar does not work yet
func (self *staticData) recordGoStatement(fdecl *ast.FuncDecl, call *ast.GoStmt) {
	self.addFuncIfNotExists(fdecl)

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

	decl := &ast.FuncDecl{
		Name: &ast.Ident{
			NamePos: lit.Pos(),
			Name:    fmt.Sprintf("<func-lit-%d>", self.nextID),
		},
		Type: lit.Type,
		Body: lit.Body,
	}
	self.nextID++

	self.funcLitDecl[lit] = decl

	self.funcInfo[fdecl].goCalls[goStmt] = decl

	self.addFuncIfNotExists(decl)

	// Analyze the function literal just like any other function.
	self.getOpsInFunc(decl)

	return decl
}
