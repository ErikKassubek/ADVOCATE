// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: Data for alias analysis
//
// Author: Erik Kassubek
// Created: 2026-06-30
//
// License: BSD-3-Clause

package s_alias

import "advocate/static/static/s_ssa"

// ================================================================
// Alias Result
// ================================================================

type aliasResult int

const (
	never aliasResult = iota
	sometimes
	always
)

func (self *aliasResult) string() string {
	switch *self {
	case never:
		return "never"
	case sometimes:
		return "sometimes"
	case always:
		return "always"
	default:
		return "unknown"
	}
}

type Data struct {
	ssa *s_ssa.Data
}
