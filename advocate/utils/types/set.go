// Copyright (c) 2025 Erik Kassubek
//
// File: pair.go
// Brief: Implement a set data type
//
// Author: Erik Kassubek
// Created: 2025-07-02
//
// License: BSD-3-Clause

package types

type Set[T comparable] struct {
	data map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{data: make(map[T]struct{})}
}

func (s *Set[T]) Add(e T) {
	s.data[e] = struct{}{}
}

func (s *Set[T]) Remove(e T) {
	delete(s.data, e)
}
