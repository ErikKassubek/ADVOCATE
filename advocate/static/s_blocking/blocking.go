// Copyright (c) 2026 Erik Kassubek
//
// File: blocking.go
// Brief: Entry point for static blocking analysis
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static"
	"advocate/utils/flags"
)

var data *static.Data

// Main function for testing static analysis.
// Todo: remove when static analysis is fully implemented
func Test() {
	RunStaticBlockingAnalysis(flags.ProgPath)
}

// init to static blocking analysis
func RunStaticBlockingAnalysis(dir string) (err error) {
	// vars := make([]*ast.Ident, 0) // TODO: determine vars

	data, err = static.BuildStaticData(dir)
	if err != nil {
		return err
	}

	// data.Ast().PrintInfo()
	data.Ssa().Print(true)

	// data.Ast().TestReachable()

	return nil
}
