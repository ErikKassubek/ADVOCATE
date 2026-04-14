// Copyright (c) 2025 Erik Kassubek
//
// File: atomic.go
// Brief: Create constraint from atomics
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package constraints

import (
	"advocate/analysis/baseA"
	"advocate/trace"
)

// AddAtomic  add the element to be the last write for an atomic write
// For an atomic read, add a constraint between the last writer and the element
//
// Parameter:
//   - elem *trace.ElementAtomic: the atomic trace element
func AddAtomic(elem *trace.ElementAtomic) {
	// all operation other than store, read from the atomic variable
	// Set a constraint with the last writer. If there is non, the variable
	// tries to read a default value, which does not create an constraint
	if elem.GetType(true) != trace.AtomicStore {
		if lw, ok := baseA.LastAtomicWriter[elem.GetObjId()]; ok {
			AddConstraint(true, lw, elem)
		}
	}

	// all operations other than load write to the atomic variable
	// set as last writer
	if elem.GetType(true) != trace.AtomicLoad {
		baseA.LastAtomicWriter[elem.GetObjId()] = elem
	}
}
