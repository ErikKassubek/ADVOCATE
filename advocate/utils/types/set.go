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

// Set represents an unordered set with unique values
type Set[T comparable] struct {
	data map[T]struct{}
}

// NewSet creates a new, empty set
//
// Returns:
//   - Set[T]: the new set
func NewSet[T comparable]() Set[T] {
	return Set[T]{data: make(map[T]struct{})}
}

// Add adds an element to the set
//
// Parameter:
//   - e T: the element to add
func (s *Set[T]) Add(e T) {
	s.data[e] = struct{}{}
}

// Remove removes an element from the set
// If the set does not contain the element, Remove is a no-op
//
// Parameter:
//   - e T: the element to remove
func (s *Set[T]) Remove(e T) {
	delete(s.data, e)
}

// Contains returns if the set contains a given element
//
// Parameter:
//   - e T: the element to check
//
// Returns:
//   - bool: true if the e is in the set, false otherwise
func (s *Set[T]) Contains(e T) bool {
	_, ok := s.data[e]
	return ok
}
