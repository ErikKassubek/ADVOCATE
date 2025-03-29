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

import "analyzer/analysis"

// TODO: is it correct, that in R1, only the element before and after the element in the same routine is used

/*
We define <c, c'> in CPOP1, if c and c' are operations in the same routine.
We define <c, c'> in CPOP2, if c and c' are operations in different routines
but on the same primitive.
From this we define the relations Rel1 and Rel2 with the following rules:
Rule 1: exists c, c', <c, c'> in CPOP1 -> c' in Rel1(c)  (same routine, element before and after)
Rule 2: exists c, c', <c, c'> in CPOP2 -> c' in Rel2(c)  (different routine, same primitive)
Rule 3: exists c, c', c'', c' in Rel1(c), c'' in Rel2(c') -> c'' in Rel2(c)
Rule 4: exists c, c', c'', c' in Rel2(c), c'' in Rel2(c') -> c'' in Rel2(c)
*/

const (
	Before = 0
	After  = 1
)

var (
	counterCPOP1 = 0
	counterCPOP2 = 0
)

func calculateRelRule1(routineTrace []analysis.TraceElement) {
	var prevValid analysis.TraceElement

	for i := range routineTrace {
		if isGoPieElem(routineTrace[i]) {
			if prevValid != nil {
				prevValid.AddRel1(routineTrace[i], After)
				routineTrace[i].AddRel1(prevValid, Before)
				counterCPOP1++
			}
			prevValid = routineTrace[i]
		}
	}
}

var (
	elemsByID = make(map[int][]analysis.TraceElement) // id -> chan/sel/mutex elem
)

func calculateRelRule2AddElem(elem analysis.TraceElement) {
	if !isGoPieElem(elem) {
		return
	}

	id := elem.GetID()
	if _, ok := elemsByID[id]; ok {
		elemsByID[id] = make([]analysis.TraceElement, 0)
	}
	elemsByID[id] = append(elemsByID[id], elem)
	counterCPOP2++
}

func calculateRelRule2() {
	for _, elems := range elemsByID {
		for i := 0; i < len(elems)-1; i++ {
			for j := i + 1; i < len(elems)-1; i++ {
				elem1 := elems[i]
				elem2 := elems[j]
				if elem1.GetRoutine() != elem2.GetRoutine() {
					elem1.AddRel2(elem2)
					elem2.AddRel2(elem1)
				}
			}
		}
	}
}

func calculateRelRule3And4() {
	hasChanged := true

	for hasChanged {
		hasChanged = false

		for _, elems := range elemsByID {
			for _, elem := range elems {
				// Rule 3
				c1 := elem.GetRel1()
				for _, c1Elem := range c1 {
					c2 := c1Elem.GetRel2()
					for _, c2Elem := range c2 {
						elem.AddRel2(c2Elem)
						hasChanged = true
					}
				}

				// Rule 4
				c1 = elem.GetRel2()
				for _, c1Elem := range c1 {
					c2 := c1Elem.GetRel2()
					for _, c2Elem := range c2 {
						elem.AddRel2(c2Elem)
						hasChanged = true
					}
				}
			}
		}
	}

}

/*
 * GoPie only looks at fork, mutex, rwmutex and channel (and select)
 * Return if the elem is one of them
 */
func isGoPieElem(elem analysis.TraceElement) bool {
	validTypes := []string{"R", "M", "C", "S"}
	elemTypeShort := elem.GetObjType(false)
	for _, t := range validTypes {
		if elemTypeShort == elemTypeShort {
			if t == "R" && elem.GetObjType(false) == "RE" { // only fork, not end
				return false
			}
			return true
		}
	}
	return false
}
