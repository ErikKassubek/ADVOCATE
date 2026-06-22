// Copyright (c) 2026 Erik Kassubek
//
// File: ast.go
// Brief: Create and work on the ast
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package blockingStatic

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// var functionVar = make(map[string]map[string]map[string]struct{}{}) // function creation location -> variable -> function
// var funcInFunc = make(map[string][]string)                          // function creation location -> called created in function

// Build the ast and a map from ast node to package information
func (self *staticData) buildAst() {
	for _, pkg := range self.pkgs {
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

func (self *staticData) getType(id *ast.Ident) types.Type {
	pkg := self.npm[id]
	if pkg == nil {
		return nil
	}

	return pkg.TypesInfo.TypeOf(id)
}

func (self *staticData) getObject(id *ast.Ident) (types.Object, *packages.Package) {
	pkg := self.npm[id]
	if pkg == nil {
		return nil, nil
	}

	return pkg.TypesInfo.ObjectOf(id), pkg
}

// TODO: does not seem to work
func (self *staticData) getNamed(id *ast.Ident) (*types.Named, bool) {
	t := self.getType(id)

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	res, ok := t.(*types.Named)

	return res, ok
}

// Given a call expression, find and record the corresponding function declaration
func (self *staticData) getFuncDecl(call *ast.CallExpr) *ast.FuncDecl {

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

func (self *staticData) calledObject(call *ast.CallExpr) types.Object {

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

func (self *staticData) FuncDeclForObject(fn *types.Func) *ast.FuncDecl {

	pos := fn.Pos()

	for _, pkg := range self.pkgs {
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
