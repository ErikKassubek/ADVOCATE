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
	"go/ast"

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

	data.printInfo()

	var fFunc *ast.FuncDecl
	var mainFunc *ast.FuncDecl

	for p, funcDecl := range data.funcDeclMap {
		if data.getPosFromPos(p) == "[/main.go:7]" {
			fFunc = funcDecl
		} else if data.getPosFromPos(p) == "[/main.go:30]" {
			mainFunc = funcDecl
		}
	}

	// TODO: only for debug. Remove
	res, path := data.isReachableFuncFromFunc(mainFunc, fFunc, true)
	if res {
		fmt.Println(path)
	} else {
		fmt.Println("No Path Found")
	}
	//

	data.runAliasAnalysis()
	return nil
}

func (self *staticData) printInfo() {
	for p, c := range self.funcsInfo {
		fmt.Println(self.getName(p.Name), self.getPos(p))

		fmt.Println("  Funcs: ")
		for _, call := range self.funcsInfo[p].funcCalls {
			fmt.Println("    ", call.name, self.getPos(call.call), self.getPos(call.decl))
		}

		fmt.Println("  Go: ")
		for ch, call := range self.funcsInfo[p].goCalls {
			if call == nil {
				fmt.Println("    ", self.getPos(ch), "FL")
			} else {
				fmt.Println("    ", self.getPos(ch), self.getPos(call))
			}
		}

		fmt.Println("  Ops: ")
		for op, ch := range c.ops {
			for f := range ch {
				fmt.Println("    ", f, self.getPos(op))
			}
		}
	}

	fmt.Println("")

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
