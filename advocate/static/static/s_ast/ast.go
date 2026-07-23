// Copyright (c) 2026 Erik Kassubek
//
// File: ast.go
// Brief: Create and work on the ast
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ast

import (
	"advocate/static/static/s_base"
	"advocate/utils/flags"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// var functionVar = make(map[string]map[string]map[string]struct{}{}) // function creation location -> variable -> function
// var funcInFunc = make(map[string][]string)                          // function creation location -> called created in function

// buildAst build the ast and a map from ast node to package information
func (self *Data) buildAst() {

	for _, pkg := range self.Pkgs {
		self.astMap[pkg.PkgPath] = pkg.Syntax
		self.ast = append(self.ast, pkg.Syntax...)

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if n != nil {
					self.npm[n] = pkg
				}
				return true
			})
		}
	}
}

// isEqual determines if two nodes point to the same code position
//
// Parameter:
//   - p ast.Node: first node
//   - q ast.Node: second node
//
// Returns:
//   - bool: true, if p and q point to the same code position, false otherwise
func (self *Data) isEqual(p, q ast.Node) bool {
	return self.getPos(p) == self.getPos(q)
}

// ================================================================
// Info
// ================================================================

func (self *Data) getType(id *ast.Ident) types.Type {
	pkg := self.npm[id]
	if pkg == nil {
		return nil
	}

	return pkg.TypesInfo.TypeOf(id)
}

func (self *Data) getObject(id *ast.Ident) (types.Object, *packages.Package) {
	pkg := self.npm[id]
	if pkg == nil {
		return nil, nil
	}

	return pkg.TypesInfo.ObjectOf(id), pkg
}

// TODO: does not seem to work
func (self *Data) getNamed(id *ast.Ident) (*types.Named, bool) {
	t := self.getType(id)

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	res, ok := t.(*types.Named)

	return res, ok
}

func (self *Data) isNilNode(node ast.Node) bool {
	if node == nil {
		return true
	}

	v := reflect.ValueOf(node)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}

// ================================================================
// Position
// ================================================================

func (self *Data) getPos(p ast.Node) string {
	if p == nil {
		return "<nil position>"
	}

	if self.isNilNode(p) {
		return "<nil node>"
	}

	pos := p.Pos()
	return self.GetPosFromPos(pos)
}

func (self *Data) GetPosFromPos(pos token.Pos) string {
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

// ================================================================
// Functions and Function Calls
// ================================================================

func (self *Data) addFuncIfNotExists(fdecl *ast.FuncDecl) {
	if _, ok := self.FuncInfo[fdecl]; !ok {
		self.FuncInfo[fdecl] = FuncInfo{
			decl:      fdecl,
			FuncCalls: make(map[FuncCall]struct{}),
			Ops:       make(map[Operation]map[ast.Expr]struct{}),
			GoCalls:   make(map[*ast.GoStmt]*ast.FuncDecl),
		}
	}
}

func (self *Data) getCallType(call *ast.CallExpr, decl *ast.FuncDecl) s_base.FuncName {
	if decl == nil {
		return self.getConcFuncName(call)
	}

	return s_base.UnknownFunc
}

// Given a call expression, find and record the corresponding function declaration
func (self *Data) getFuncDecl(call *ast.CallExpr) *ast.FuncDecl {

	obj := self.calledObject(call)
	if obj == nil {
		return nil
	}

	fn, ok := obj.(*types.Func)
	if !ok {
		return nil
	}

	return self.FuncDeclForObject(fn)
}

func (self *Data) calledObject(call *ast.CallExpr) types.Object {

	for _, info := range self.pkgInfo {
		switch fun := call.Fun.(type) {

		case *ast.Ident:
			if obj := info.Uses[fun]; obj != nil {
				return obj
			}

		case *ast.SelectorExpr:
			// pkg.Func or x.Method
			if obj := info.Uses[fun.Sel]; obj != nil {
				return obj
			}
		}
	}

	return nil
}

func (self *Data) FuncDeclForObject(fn *types.Func) *ast.FuncDecl {

	pos := fn.Pos()

	for _, pkg := range self.Pkgs {
		for _, file := range pkg.Syntax {

			var found *ast.FuncDecl

			ast.Inspect(file, func(n ast.Node) bool {
				fd, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				if fd.Name.Pos() == pos {
					found = fd
					return false
				}

				return true
			})

			if found != nil {
				return found
			}
		}
	}

	return nil
}

func (self *Data) FuncContainsOp(fn *ast.FuncDecl, op Operation) bool {
	info, ok := self.FuncInfo[fn]
	if !ok {
		return false
	}

	_, res := info.Ops[op]

	return res
}

// ================================================================
// Routines
// ================================================================

func (self *Data) resolveGoFunc(goStmt *ast.GoStmt) *ast.FuncDecl {
	ident, ok := goStmt.Call.Fun.(*ast.Ident)
	if !ok {
		return nil
	}

	pkg := self.npm[goStmt]
	if pkg == nil {
		return nil
	}

	info := self.pkgInfo[pkg]
	if info == nil {
		return nil
	}

	obj := info.ObjectOf(ident)
	if obj == nil {
		return nil
	}

	fn, ok := obj.(*types.Func)
	if !ok {
		return nil
	}

	pos := fn.Pos()
	if !pos.IsValid() {
		return nil
	}

	return self.FuncDeclMap[pos]
}
