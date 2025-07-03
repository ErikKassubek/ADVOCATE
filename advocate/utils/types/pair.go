// Copyright (c) 2025 Erik Kassubek
//
// File: pair.go
// Brief: Implement a pair data type
//
// Author: Erik Kassubek
// Created: 2025-07-02
//
// License: BSD-3-Clause

package types

type Pair[K any, V any] struct {
	X K
	Y V
}

func NewPair[K any, V any](x K, y V) Pair[K, V] {
	return Pair[K, V]{X: x, Y: y}
}
