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

func (this *Data) MayAlias(v1, v2 *Instruction) bool {
	alloc1 := this.alloc[v1]
	alloc2 := this.alloc[v2]

	return types.HasCommonElement(alloc1, alloc2)
}
