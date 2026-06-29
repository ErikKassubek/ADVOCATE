// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package static

import (
	"advocate/utils/log"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

type objId int

type funcCall struct {
	call     *ast.CallExpr
	decl     *ast.FuncDecl
	name     string // TODO: multi package
	callType funcName
}

type operation struct {
	obj objId
	op  funcName
}

type funcInfo struct {
	decl *ast.FuncDecl

	// functions called in each function,
	// func -> call
	// ast.Expr.(type) -> *ast.Ident: direct function (foo())
	// ast.Expr.(type) -> *ast.SelectorExpr: methodCall (obj.Method())
	// ast.Expr.(type) -> *ast.FuncLit: function literal (func() {...}())
	funcCalls map[funcCall]struct{}

	// routine spawns from functions
	// *ast.GoStmt.Call.(type) -> *ast.Ident: direct function (go foo())
	// *ast.GoStmt.Call.(type) -> *ast.SelectorExpr: methodCall (go obj.Method())
	// *ast.GoStmt.Call.(type) -> *ast.FuncLit: function literal (go func() { ... }())
	goCalls map[*ast.GoStmt]*ast.FuncDecl

	ops map[operation]map[ast.Expr]struct{}
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
	funcInfo    map[*ast.FuncDecl]funcInfo
	routFunc    map[*ast.GoStmt]*ast.FuncDecl
	funcLitDecl map[*ast.FuncLit]*ast.FuncDecl // dummy for func lit

	nextID int
}

func BuildStaticData(dir string) (*staticData, error) {
	data := &staticData{
		dir:  dir,
		fset: token.NewFileSet(),

		astMap: make(map[string][]*ast.File),
		ast:    make([]*ast.File, 0),
		npm:    make(map[ast.Node]*packages.Package),

		funcDeclMap: make(map[token.Pos]*ast.FuncDecl),
		funcInfo:    make(map[*ast.FuncDecl]funcInfo),
		routFunc:    make(map[*ast.GoStmt]*ast.FuncDecl),

		funcLitDecl: make(map[*ast.FuncLit]*ast.FuncDecl),
	}

	err := data.loadPackages()
	if err != nil {
		log.Error(err.Error())
		return data, err
	}

	data.buildTypeInfo()

	data.buildAst()

	data.buildSsa()
	// data.printSSA(true)
	// fmt.Println("\n\n\n")
	// data.runPointerAnalysis()

	data.CollectOperations()

	return data, nil
}

// Determine the packages and type info
//
// Parameter:
//   - dir: string: root directory of project
func (self *staticData) loadPackages() error {
	cfg := &packages.Config{
		Fset: self.fset,
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports,
		Dir:   self.dir,
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, self.dir)
	if err != nil {
		return fmt.Errorf("static analysis: failed to load packages: %w", err)
	}

	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			return fmt.Errorf("static analysis: packages contain errors: %s", err.Error())
		}
	}

	self.pkgs = pkgs
	return nil
}

func (self *staticData) buildTypeInfo() {
	self.pkgInfo = make(map[*packages.Package]*types.Info)

	self.uses = make(map[*ast.Ident]types.Object)
	self.defs = make(map[*ast.Ident]types.Object)

	for _, pkg := range self.pkgs {
		if pkg.TypesInfo == nil {
			continue
		}

		self.pkgInfo[pkg] = pkg.TypesInfo

		for ident, obj := range pkg.TypesInfo.Uses {
			self.uses[ident] = obj
		}

		for ident, obj := range pkg.TypesInfo.Defs {
			self.defs[ident] = obj
		}
	}
}

func (self *staticData) PrintInfo() {
	fmt.Print("=================== Info ===================\n\n")
	for p, c := range self.funcInfo {
		pos := self.getPos(p)
		if pos == "[internal]" {
			continue
		}
		fmt.Println(self.getName(p.Name), pos)

		fmt.Println("\tFuncs: ")
		for call := range self.funcInfo[p].funcCalls {
			fmt.Println("\t\t", call.name, self.getPos(call.call), self.getPos(call.decl))
		}

		fmt.Println("\tGo: ")
		for ch, call := range self.funcInfo[p].goCalls {
			fmt.Println("\t\t", call.Name, self.getPos(ch), self.getPos(call))
		}

		fmt.Println("\tOps: ")
		for op, expr := range c.ops {
			for e := range expr {
				fmt.Println("\t\t", op.obj, op.op, self.getPos(e))
			}
		}

		fmt.Println("")
	}

}
