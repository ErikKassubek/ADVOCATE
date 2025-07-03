// Copyright (c) 2025 Erik Kassubek
//
// File: pair.go
// Brief: Implement a stack data type
//
// Author: Erik Kassubek
// Created: 2025-07-02
//
// License: BSD-3-Clause

package types

type Stack[T any] struct {
	items []T
}

// Push adds an item to the top
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop removes and returns the top item
func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		var zero T
		return zero // empty stack
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top
}

// Peek returns the top item without removing it
func (s *Stack[T]) Peek() T {
	if len(s.items) == 0 {
		var zero T
		return zero
	}
	return s.items[len(s.items)-1]
}

// IsEmpty returns whether the stack is empty
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of items
func (s *Stack[T]) Size() int {
	return len(s.items)
}
