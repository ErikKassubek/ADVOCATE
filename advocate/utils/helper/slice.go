// Copyright (c) 2024 Erik Kassubek
//
// File: slice.go
// Brief: Utility functions on slices
//
// Author: Erik Kassubek
// Created: 2024-04-06
//
// License: BSD-3-Clause

package helper

import (
	"strings"
)

// Contains checks if a slice of strings contains an element
//
// Parameter:
//   - s []T comparable) slice to check
//   - e T comparable: element to check
//
// Returns:
//   - bool: true is e in s, false otherwise
func Contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// SplitAtLast splits the string into two parts at the last occurrence of the separator
//
// Parameter:
//   - str string: string to split
//   - sep string: separator to split at
//
// Returns:
//   - []string: If sep in string: list with two elements split at the sep,
//   - if not then list containing str
func SplitAtLast(str string, sep string) []string {
	if sep == "" {
		return []string{str}
	}

	i := strings.LastIndex(str, sep)
	if i == -1 {
		return []string{str}
	}
	return []string{str[:i], str[i+1:]}
}

// AddIfNotContains adds an element to a list, if it does not contain the element
//
// Parameter:
//   - l []T comparable: The list
//   - e T comparable: The element
func AddIfNotContains[T comparable](l []T, e T) []T {
	if !Contains(l, e) {
		l = append(l, e)
	}
	return l
}

// MergeLists takes two lists and returns a list containing all the elements from both
// lists. The resulting list does not contain duplicated.
//
// Parameter:
//   - l1 []T comparable: list 1
//   - l2 []T comparable: list 2
func MergeLists[T comparable](l1, l2 []T) []T {
	uniqueMap := make(map[T]bool)
	res := []T{}

	for _, val := range l1 {
		if !uniqueMap[val] {
			uniqueMap[val] = true
			res = append(res, val)
		}
	}

	for _, val := range l2 {
		if !uniqueMap[val] {
			uniqueMap[val] = true
			res = append(res, val)
		}
	}

	return res
}
