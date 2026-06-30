// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the ast analysis
//
// Author: Erik Kassubek
// Created: 2026-06-30
//
// License: BSD-3-Clause

package staticAST

import (
	"advocate/static/static/staticBase"
	"advocate/utils/log"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"runtime"

	"golang.org/x/tools/go/packages"
)

type objId int

type funcCall struct {
	call     *ast.CallExpr
	decl     *ast.FuncDecl
	name     string // TODO: multi package
	callType staticBase.FuncName
}

type operation struct {
	obj objId
	op  staticBase.FuncName
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

type Data struct {
	dir string // path to analyzed program

	Pkgs []*packages.Package

	fset *token.FileSet

	pkgInfo map[*packages.Package]*types.Info
	uses    map[*ast.Ident]types.Object
	defs    map[*ast.Ident]types.Object

	astMap map[string][]*ast.File         // pkg path -> files
	ast    []*ast.File                    // flattened list
	npm    map[ast.Node]*packages.Package // node packages map

	funcDeclMap map[token.Pos]*ast.FuncDecl
	funcInfo    map[*ast.FuncDecl]funcInfo
	routFunc    map[*ast.GoStmt]*ast.FuncDecl
	funcLitDecl map[*ast.FuncLit]*ast.FuncDecl // dummy for func lit

	nextFuncLitId int
}

func BuildAst(dir string) (*Data, error) {
	data := &Data{
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

	data.CollectOperations()

	return data, nil
}

// Determine the packages and type info
//
// Parameter:
//   - dir: string: root directory of project
func (self *Data) loadPackages() error {
	cfg := &packages.Config{
		Fset: self.fset,
		Mode: packages.LoadAllSyntax,
		Dir:  self.dir,
		Env: append(os.Environ(),
			"GO111MODULE=on",
			"GOROOT="+runtime.GOROOT(),
		),
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

	self.Pkgs = pkgs
	return nil
}

func (self *Data) buildTypeInfo() {
	self.pkgInfo = make(map[*packages.Package]*types.Info)

	self.uses = make(map[*ast.Ident]types.Object)
	self.defs = make(map[*ast.Ident]types.Object)

	for _, pkg := range self.Pkgs {
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

func (self *Data) PrintInfo() {
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
