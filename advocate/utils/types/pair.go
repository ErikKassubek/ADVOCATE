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

// Pair is a type to implement a pair of values
// The types of the values do not have to be the same
type Pair[K comparable, V comparable] struct {
	X K
	Y V
}

// NewPair returns a new pair
//
// Parameter:
//   - x comparable: first value
//   - y comparable: second value
func NewPair[K comparable, V comparable](x K, y V) Pair[K, V] {
	return Pair[K, V]{X: x, Y: y}
}
