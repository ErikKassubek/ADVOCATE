// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data required for por
//
// Author: Erik Kassubek
// Created: 2026-03-16
//
// License: BSD-3-Clause

package por

import "advocate/fuzzing/baseF"

var alreadyRunROC = make([]baseF.Constraint, 0)

func Reset() {
	alreadyRunROC = make([]baseF.Constraint, 0)
}
