// Copyright (c) 2026 Erik Kassubek
//
// File: maths.go
// Brief: MAths functions
//
// Author: Erik Kassubek
// Created: 2026-05-30
//
// License: BSD-3-Clause

package math

func Clamp[T ~int | ~int64 | ~float64](v, low, high T) T {
	return min(max(v, low), high)
}
