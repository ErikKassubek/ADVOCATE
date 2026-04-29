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
	"fmt"

	"golang.org/x/tools/go/packages"
)

// init to static blocking analysis
func RunStaticBlockingAnalysis(dir string) error {
	// vars := make([]*ast.Ident, 0) // TODO: determine vars

	data, err := buildStaticData(dir)
	if err != nil {
		return err
	}

	data.collectOperations()

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
			packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, self.dir)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	self.pkgs = pkgs
	return nil
}
