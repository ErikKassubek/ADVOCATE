// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: Alias analysis
//
// Author: Erik Kassubek
// Created: 2026-06-30
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/utils/types"
)

func (self *Data) MayAlias(v1, v2 *Instruction) bool {
	alloc1 := self.alloc[v1]
	alloc2 := self.alloc[v2]

	return types.HasCommonElement(alloc1, alloc2)
}
