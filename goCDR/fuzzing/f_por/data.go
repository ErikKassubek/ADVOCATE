// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data required for por
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_por

import "gocdr/fuzzing/f_base"

var alreadyRunROC = make([]f_base.Constraint, 0)

func Reset() {
	alreadyRunROC = make([]f_base.Constraint, 0)
}
