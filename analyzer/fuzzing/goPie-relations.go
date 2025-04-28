// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-relations.go
// Brief: Calculate the relation for goPie
//
// Author: Erik Kassubek
// Created: 2025-03-24
//
// License: BSD-3-Clause

package fuzzing

import (
	"analyzer/trace"
	"analyzer/utils"
)

// We define <c, c'> in CPOP1, if c and c' are operations in the same routine.
// We define <c, c'> in CPOP2, if c and c' are operations in different routines
// but on the same primitive.
// From this we define the relations Rel1 and Rel2 with the following rules:
// Rule 1: exists c, c', <c, c'> in CPOP1 -> c' in Rel1(c)  (same routine, element before and after)
// Rule 2: exists c, c', <c, c'> in CPOP2 -> c' in Rel2(c)  (different routine, same primitive)
// Rule 3: exists c, c', c'', c' in Rel1(c), c'' in Rel2(c') -> c'' in Rel2(c)
// Rule 4: exists c, c', c'', c' in Rel2(c), c'' in Rel2(c') -> c'' in Rel2(c)

var (
	counterCPOP1 = 0
	counterCPOP2 = 0

	rel1 = make(map[trace.TraceElement]map[trace.TraceElement]struct{})
	rel2 = make(map[trace.TraceElement]map[trace.TraceElement]struct{})
)

// For each element in a routine trace, store the rule 1 information
//
// Parameter:
//   - routineTrace []analysis.TraceElement: the list of elems in the same trace
func calculateRelRule1(routineTrace []trace.TraceElement) {
	for i := 0; i < len(routineTrace)-1; i++ {
		elem1 := routineTrace[i]
		if !isGoPieElem(elem1) {
			continue
		}
		for j := i + 1; j < len(routineTrace); j++ {
			elem2 := routineTrace[j]
			if !isGoPieElem(elem2) {
				continue
			}
			if _, ok := rel1[elem1]; !ok {
				rel1[elem1] = make(map[trace.TraceElement]struct{})
			}
			if _, ok := rel1[elem2]; !ok {
				rel1[elem2] = make(map[trace.TraceElement]struct{})
			}
			rel1[elem1][elem2] = struct{}{}
			rel1[elem2][elem1] = struct{}{}
		}
	}
}

// For each element in a routine trace, add it to the map from id to operation
//
// Parameter:
//   - elem analysis.TraceElement: Element to add
func calculateRelRule2AddElem(elem trace.TraceElement) {
	if !isGoPieElem(elem) {
		return
	}

	id := elem.GetID()
	if _, ok := elemsByID[id]; !ok {
		elemsByID[id] = make([]trace.TraceElement, 0)
	}
	elemsByID[id] = append(elemsByID[id], elem)
	counterCPOP2++
}

// For all elements apply rule 2
func calculateRelRule2() {
	for _, elems := range elemsByID {
		for i := 0; i < len(elems)-1; i++ {
			for j := i + 1; i < len(elems); i++ {
				elem1 := elems[i]
				elem2 := elems[j]
				if elem1.GetRoutine() != elem2.GetRoutine() {
					if _, ok := rel2[elem1]; !ok {
						rel2[elem1] = make(map[trace.TraceElement]struct{})
					}
					if _, ok := rel2[elem2]; !ok {
						rel2[elem2] = make(map[trace.TraceElement]struct{})
					}

					rel2[elem1][elem2] = struct{}{}
					rel2[elem2][elem1] = struct{}{}
				}
			}
		}
	}
}

// For all elements apply rules 3 and 4
func calculateRelRule3And4() {
	hasChanged := true

	for hasChanged {
		hasChanged = false

		// Rule3
		for c, rel := range rel1 {
			for c1 := range rel {
				for c2 := range rel2[c1] {
					if c.GetTraceID() == c2.GetTraceID() {
						continue
					}
					if _, ok := rel2[c]; !ok {
						rel2[c] = make(map[trace.TraceElement]struct{})
					}
					if _, ok := rel2[c][c2]; !ok {
						hasChanged = true
					}
					rel2[c][c2] = struct{}{}
				}
			}
		}

		// rule4
		for c, rel := range rel2 {
			for c1 := range rel {
				for c2 := range rel2[c1] {
					if c.GetTraceID() == c2.GetTraceID() {
						continue
					}
					if _, ok := rel2[c]; !ok {
						rel2[c] = make(map[trace.TraceElement]struct{})
					}
					if _, ok := rel2[c][c2]; !ok {
						hasChanged = true
					}
					rel2[c][c2] = struct{}{}
				}
			}
		}
	}
}

// GoPie only looks at fork, mutex, rwmutex and channel (and select)
// GoPieHB uses all repayable elements
//
// Parameter:
//   - elem analysis.TraceElement: the element to check
//
// Returns:
//   - bool: true if elem should be used in chains, false if not
func isGoPieElem(elem trace.TraceElement) bool {
	elemTypeShort := elem.GetObjType(false)

	if fuzzingMode == GoPie {
		validTypes := []string{
			trace.ObjectTypeMutex, trace.ObjectTypeChannel,
			trace.ObjectTypeSelect}
		return utils.Contains(validTypes, elemTypeShort)
	}

	invalidTypes := []string{trace.ObjectTypeNew,
		trace.ObjectTypeReplay, trace.ObjectTypeRoutineEnd}
	return !utils.Contains(invalidTypes, elemTypeShort)
}
