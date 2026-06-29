// Copyright (c) 2026 Erik Kassubek
//
// File: blocking.go
// Brief: Entry point for static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package staticBlocking

import (
	"advocate/static/static"
	"advocate/utils/flags"
)

// Main function for testing static analysis.
// Todo: remove when static analysis is fully implemented
func Test() {
	RunStaticBlockingAnalysis(flags.ProgPath)
}

// init to static blocking analysis
func RunStaticBlockingAnalysis(dir string) error {
	// vars := make([]*ast.Ident, 0) // TODO: determine vars

	_, err := static.BuildStaticData(dir)
	if err != nil {
		return err
	}

	// data.PrintInfo()
	// data.PrintSSA(true)

	// data.TestReachable()

	return nil
}
