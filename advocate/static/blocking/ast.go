// Copyright (c) 2026 Erik Kassubek
//
// File: ast.go
// Brief: Create and work on the ast
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package blocking

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// var functionVar = make(map[string]map[string]map[string]struct{}{}) // function creation location -> variable -> function
// var funcInFunc = make(map[string][]string)                          // function creation location -> called created in function

// Build the ast and a map from ast node to package information
func (self *staticData) buildAst() {
	self.npm = make(map[ast.Node]*packages.Package)

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

func (self *staticData) getNamed(id *ast.Ident) (*types.Named, bool) {
	t := self.getType(id)

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	res, ok := t.(*types.Named)

	return res, ok
}

// sameVar determines, if two variables are the same
// TODO: this is not correct.
// TODO: replace
//
// Parameter:
//   - id1 *ast.Ident: first variable
//   - id2 *ast.Ident: second variable
func (self *staticData) sameVar(id1, id2 *ast.Ident) bool {
	obj1, pkg1 := self.getObject(id1)
	obj2, pkg2 := self.getObject(id2)

	if obj1 == nil || obj2 == nil || pkg1 != pkg2 {
		return false
	}

	return obj1 == obj2
}
