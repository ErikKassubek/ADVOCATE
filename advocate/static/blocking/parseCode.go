// Copyright (c) 2026 Erik Kassubek
//
// File: parseFiles.go
// Brief: Parse source code file to read access information
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package blocking

import (
	"advocate/utils/log"
	tp "advocate/utils/types"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// var functionVar = make(map[string]map[string]map[string]struct{}{}) // function creation location -> variable -> function
// var funcInFunc = make(map[string][]string)                          // function creation location -> called created in function

// Determine the packages and type info
//
// Parameter:
//   - dir: string: root directory of project
func loadPackagesAndFset(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	return pkgs, nil
}

// Build a map from ast node to package information
//
// Parameter:
//   - pkgs []*packages.Package: the package information
//
// Returns:
//   - map[ast.Node]*packages.Package: map from node to package information
func buildNodePackageMap(pkgs []*packages.Package) map[ast.Node]*packages.Package {
	nodePkg := make(map[ast.Node]*packages.Package)

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if n != nil {
					nodePkg[n] = pkg
				}
				return true
			})
		}
	}

	return nodePkg
}

// parse the files the determine the type information
func parseFiles(fset *token.FileSet, npm map[ast.Node]*packages.Package, dir string, vars []*ast.Ident) map[*ast.FuncDecl]map[*ast.Ident]bool {
	fileNames := make([]string, 0) // TODO: get files

	res := make(map[*ast.FuncDecl]map[*ast.Ident]bool) // func -> var in func
	for _, fname := range fileNames {
		resFile, err := parseFile(fset, npm, fname, vars)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		tp.MergeMaps(res, resFile)
	}

	return res
}

// for now only channel and mutex
func parseFile(fset *token.FileSet, npm map[ast.Node]*packages.Package, fileName string, vars []*ast.Ident) (map[*ast.FuncDecl]map[*ast.Ident]bool, error) {
	file, err := parser.ParseFile(fset, fileName, nil, 0)
	if err != nil {
		return nil, err
	}

	variables := make(map[*ast.FuncDecl]map[*ast.Ident]bool) // func -> var
	functions := make(map[*ast.FuncDecl]map[types.Object]bool)

	// TODO: anonymous functions, go func() {}, go A(), ...
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			variablesInFunc := make(map[*ast.Ident]bool)
			funcCallsInFunc := make(map[types.Object]bool)
			ast.Inspect(fn.Body, func(n ast.Node) bool {

				// check variables
				if id, ok := n.(*ast.Ident); ok {
					for _, v := range vars {
						if sameVar(id, v, npm) {
							variablesInFunc[v] = true
						}
					}
				}

				// check function calls
				if call, ok := n.(*ast.CallExpr); ok {
					var obj types.Object

					switch fun := call.Fun.(type) {
					case *ast.Ident:
						obj, _ = getObject(fun, npm)
					case *ast.SelectorExpr:
						obj, _ = getObject(fun.Sel, npm)
					}

					if obj != nil {
						funcCallsInFunc[obj] = true
					}

				}

				return true

			})
			variables[fn] = variablesInFunc
			functions[fn] = funcCallsInFunc
		}
	}

	return variables, nil
}

func getObject(id *ast.Ident, npm map[ast.Node]*packages.Package) (types.Object, *packages.Package) {
	pkg := npm[id]
	if pkg == nil {
		return nil, nil
	}

	return pkg.TypesInfo.ObjectOf(id), pkg
}

// sameVar determines, if two variables are the same
//
// Parameter:
//   - id1 *ast.Ident: first variable
//   - id2 *ast.Ident: second variable
//   - info *types.Info
func sameVar(id1, id2 *ast.Ident, npm map[ast.Node]*packages.Package) bool {
	obj1, pkg1 := getObject(id1, npm)
	obj2, pkg2 := getObject(id2, npm)

	if obj1 == nil || obj2 == nil || pkg1 != pkg2 {
		return false
	}

	return obj1 == obj2
}
