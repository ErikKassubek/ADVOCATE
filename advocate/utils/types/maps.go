// Copyright (c) 2025 Erik Kassubek
//
// File: maps.go
// Brief: Helper functions for maps
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package types

func MergeMaps[K comparable, V any](dst, src map[K]V) {
	for k, v := range src {
		dst[k] = v
	}
}
