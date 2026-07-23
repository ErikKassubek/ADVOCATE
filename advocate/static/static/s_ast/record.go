// Copyright (c) 2026 Erik Kassubek
//
// File: record.go
// Brief: Record data
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ast

import (
	"advocate/static/static/s_base"
	"fmt"
	"go/ast"
)

// TODO: call is not recorded if in funcLit
func (self *Data) recordFunctionCall(fdecl *ast.FuncDecl, call *ast.CallExpr) {
	// prevent function from calling itself, if it is not recursive
	if self.getPos(call) == self.getPos(fdecl) {
		return
	}

	name := self.getName(call)
	if name == "FuncLit" {
		return
	}

	funcDecl := self.getFuncDecl(call)
	info := self.FuncInfo[fdecl]
	info.FuncCalls[FuncCall{call, funcDecl, name, self.getCallType(call, funcDecl)}] = struct{}{}
	self.FuncInfo[fdecl] = info
}

func (self *Data) recordOperation(f *ast.FuncDecl, expr ast.Expr, name s_base.FuncName) {
	info := self.FuncInfo[f]

	if info.Ops == nil {
		info.Ops = make(map[Operation]map[ast.Expr]struct{})
	}

	// TODO: object ids
	op := Operation{
		1,
		name,
	}

	if _, ok := info.Ops[op]; !ok {
		info.Ops[op] = make(map[ast.Expr]struct{})
	}

	info.Ops[op][expr] = struct{}{}

	self.FuncInfo[f] = info
}

// TODO: go mu.Lock() and similar does not work yet
func (self *Data) recordGoStatement(fdecl *ast.FuncDecl, call *ast.GoStmt) {
	info := self.FuncInfo[fdecl]

	funcDecl := self.resolveGoFunc(call)

	// handle func lit
	if funcDecl == nil {
		funcDecl = self.recordFuncLitGo(fdecl, call)
		if funcDecl == nil {
			return
		}
	}

	info.GoCalls[call] = funcDecl
	self.FuncInfo[fdecl] = info

	self.RoutFunc[call] = fdecl
}

func (self *Data) recordFuncLitGo(
	fdecl *ast.FuncDecl,
	goStmt *ast.GoStmt,
) *ast.FuncDecl {

	if _, ok := self.FuncInfo[fdecl].GoCalls[goStmt]; ok {
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

	self.FuncLitDecl[lit] = decl

	self.FuncInfo[fdecl].GoCalls[goStmt] = decl

	self.addFuncIfNotExists(decl)

	// Analyze the function literal just like any other function.
	self.detOpsInFunc(decl)

	return decl
}
