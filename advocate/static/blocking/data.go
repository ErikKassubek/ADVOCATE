// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blocking

import (
	"advocate/utils/log"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

type staticData struct { // always use buildStaticData, never staticData{}
	dir string // path to analyzed program

	pkgs []*packages.Package

	fst *token.FileSet

	typeInfo *types.Info

	astMap map[string][]*ast.File         // pkg path -> files
	ast    []*ast.File                    // flattened list
	npm    map[ast.Node]*packages.Package // node packages map

	ssa     *ssa.Program // static single assignment (intermediate program representation where each variable is assigned exactly once)
	ssaPkgs []*ssa.Package

	callGraph *callgraph.Graph

	// TODO: determine
	opsPerFunk map[*ast.FuncDecl]map[*ast.Expr]map[funcs]struct{} // operations per function: func -> vartiable id (TODO: define) -> funcs
	opsPerRout map[int]map[*ast.Expr]map[funcs]struct{}           // operations per routine routine id (TODO: define or use ast value) -> vartiable id (TODO: define) -> funcs
}

func buildStaticData(dir string) (*staticData, error) {
	data := &staticData{
		dir: dir,
		fst: token.NewFileSet(),
	}

	err := data.loadPackages()
	if err != nil {
		log.Error(err.Error())
		return data, err
	}

	data.buildAst()

	// must be called afer load packages
	data.buildSsa()
	data.buildCallGraph()

	return data, nil
}
