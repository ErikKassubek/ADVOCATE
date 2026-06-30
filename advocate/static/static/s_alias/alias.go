// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: Entry for alias analysis
//
// Author: Erik Kassubek
// Created: 2026-06-30
//
// License: BSD-3-Clause

package s_alias

import "advocate/static/static/s_ssa"

func RunAliasAnalysis(ssa *s_ssa.Data) *Data {
	data := &Data{
		ssa: ssa,
	}

	// TODO: implement

	return data
}
