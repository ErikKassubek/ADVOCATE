// Copyright (c) 2025 Erik Kassubek
//
// File: utils.go
// Brief: Utils for memory
//
// Author: Erik Kassubek
// Created: 2025-03-11
//
// License: BSD-3-Clause

package memory

import (
	"reflect"
	"unsafe"
)

// GetSizeInMB recursively estimates the memory usage of a slice, map, or nested structures.
func GetSizeInMB(data interface{}) float64 {
	visited := make(map[uintptr]bool) // Track visited pointers to prevent infinite loops
	sizeBytes := getSizeRecursive(reflect.ValueOf(data), visited)
	return float64(sizeBytes) / (1024 * 1024) // Convert to MB
}

// getSizeRecursive calculates the size of a value recursively.
func getSizeRecursive(val reflect.Value, visited map[uintptr]bool) int {
	if !val.IsValid() {
		return 0
	}

	switch val.Kind() {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return 0
		}
		// Use Elem() to get the actual value inside the pointer/interface
		return getSizeRecursive(val.Elem(), visited)

	case reflect.Slice, reflect.Array:
		totalSize := int(unsafe.Sizeof(val.Interface())) // Slice header size
		for i := 0; i < val.Len(); i++ {
			totalSize += getSizeRecursive(val.Index(i), visited)
		}
		return totalSize

	case reflect.Map:
		if val.Len() == 0 {
			return 0
		}
		// Check if map is already visited
		ptr := val.Pointer()
		if ptr != 0 && visited[ptr] {
			return 0
		}
		visited[ptr] = true

		totalSize := int(unsafe.Sizeof(val.Interface())) // Map header size
		iter := val.MapRange()
		for iter.Next() {
			totalSize += getSizeRecursive(iter.Key(), visited)
			totalSize += getSizeRecursive(iter.Value(), visited)
		}
		return totalSize

	default:
		return int(unsafe.Sizeof(val.Interface()))
	}
}
