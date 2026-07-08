// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data required for por
//
// Author: Erik Kassubek
// Created: 2026-03-16
//
// License: BSD-3-Clause

package f_por

import "advocate/fuzzing/f_base"

var alreadyRunROC = make([]f_base.Constraint, 0)

func Reset() {
	alreadyRunROC = make([]f_base.Constraint, 0)
}
