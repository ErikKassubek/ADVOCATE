// Copyright (c) 2026 Erik Kassubek
//
// File: parseFiles.go
// Brief: Parse source code file to read access information
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package blockingStatic

import (
	"advocate/utils/flags"
	"fmt"

	"golang.org/x/tools/go/packages"
)

// Main function for testing static analysis.
// Todo: remove when static analysis is fully implemented
func Test() {
	RunStaticBlockingAnalysis(flags.ProgPath)
}

// init to static blocking analysis
func RunStaticBlockingAnalysis(dir string) error {
	// vars := make([]*ast.Ident, 0) // TODO: determine vars

	data, err := buildStaticData(dir)
	if err != nil {
		return err
	}

	data.collectOperations()
	data.runAliasAnalysis()
	return nil
}

// Determine the packages and type info
//
// Parameter:
//   - dir: string: root directory of project
func (self *staticData) loadPackages() error {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.LoadAllSyntax,
		Dir: self.dir,
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
