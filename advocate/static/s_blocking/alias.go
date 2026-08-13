// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: Alias Info
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

type aliasInfo struct {
	inst  s_ssa.Instruction
	alias []trace.Resource
}
