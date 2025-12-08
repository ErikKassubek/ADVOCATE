// Copyright (c) 2025 Erik Kassubek
//
// File: dyad.go
// Brief: Implement a non-comparable pair data type
//
// Author: Erik Kassubek
// Created: 2025-07-02
//
// License: BSD-3-Clause

package types

// Dyad is a type to implement a pair of values
// The types of the values do not have to be the same
type Dyad[K any, V any] struct {
	X K
	Y V
}

// NewDyad returns a new pair
//
// Parameter:
//   - x comparable: first value
//   - y comparable: second value
func NewDyad[K any, V any](x K, y V) Dyad[K, V] {
	return Dyad[K, V]{X: x, Y: y}
}
