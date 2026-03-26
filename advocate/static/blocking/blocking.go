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
	"go/ast"
	"go/token"
)

func blocking() {
	dir := ""                     // TODO: determine program directory
	vars := make([]*ast.Ident, 0) // TODO: determine vars

	fset := token.NewFileSet()

	pkgs, err := loadPackagesAndFset(dir)
	if err != nil {
		log.Error(err.Error())
	}

	npm := buildNodePackageMap(pkgs)

	_ = parseFiles(fset, npm, dir, vars) // information on which function contains which variables from vars
}
