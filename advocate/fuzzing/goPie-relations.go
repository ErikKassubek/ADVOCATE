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
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/memory"
	"sort"
)

// We define <c, c'> in CPOP1, if c and c' are operations in the same routine.
// We define <c, c'> in CPOP2, if c and c' are operations in different routines
// but on the same primitive.
// From this we define the relations Rel1 and Rel2 with the following rules:
// Rule 1: exists c, c', <c, c'> in CPOP1 -> c' in Rel1(c)  (same routine, element before and after)
// Rule 2: exists c, c', <c, c'> in CPOP2 -> c' in Rel2(c)  (different routine, same primitive)
// Rule 3: exists c, c', c'', c' in Rel1(c), c'' in Rel2(c') -> c'' in Rel2(c)
// Rule 4: exists c, c', c'', c' in Rel2(c), c'' in Rel2(c') -> c'' in Rel2(c)

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
			counterCPOP1++
			break
		}
		if memory.WasCanceled() {
			return
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

// For all elements apply rule 2 and directly rule 4
// This calculates rel2 as the tuples, of elements on the same primitive
// but in different routines
func calculateRelRule2And4() {
	for _, elems := range elemsByID {
		sort.Slice(elems, func(i, j int) bool {
			return elems[i].GetTSort() < elems[j].GetTSort()
		})

		for i := 0; i < len(elems)-1; i++ {
			elem1 := elems[i]
			for j := i + 1; j < len(elems)-1; j++ {
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
					counterCPOP2++
				}
			}
			if memory.WasCanceled() {
				return
			}
		}
	}
}

// For all elements apply rule 3
func calculateRelRule3() {
	changed := true
	for changed {
		changed = false

		// Rule 3
		for c, rel1Elems := range rel1 {
			for cPrime := range rel1Elems {
				if rel2Elems, ok := rel2[cPrime]; ok {
					if _, exists := rel2[c]; !exists {
						rel2[c] = make(map[trace.TraceElement]struct{})
					}
					for cDoublePrime := range rel2Elems {
						if _, exists := rel2[c][cDoublePrime]; !exists {
							rel2[c][cDoublePrime] = struct{}{}
							changed = true
						}
					}
				}
			}
		}

		// Rule 4
		// for c, rel2Elems := range rel2 {
		// 	newTargets := make(map[trace.TraceElement]struct{})
		// 	for cPrime := range rel2Elems {
		// 		if nestedElems, ok := rel2[cPrime]; ok {
		// 			for cDoublePrime := range nestedElems {
		// 				if _, exists := rel2Elems[cDoublePrime]; !exists {
		// 					newTargets[cDoublePrime] = struct{}{}
		// 					changed = true
		// 				}
		// 			}
		// 		}
		// 	}

		// 	for k := range newTargets {
		// 		rel2[c][k] = struct{}{}
		// 	}
		// }
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
		return helper.Contains(validTypes, elemTypeShort)
	}

	invalidTypes := []string{trace.ObjectTypeNew,
		trace.ObjectTypeReplay, trace.ObjectTypeRoutineEnd}
	return !helper.Contains(invalidTypes, elemTypeShort)
}
