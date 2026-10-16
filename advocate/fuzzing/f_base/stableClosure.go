// Copyright (c) 2026 Erik Kassubek
//
// File: stableClosure.go
// Brief: Calculate the stable closure for a scheduling constraint
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_base

import (
	"advocate/analysis/a_base"
	"advocate/trace"
)

// The stable closure is calculated based on the following three rules.
// Assume there is a constraint s and trace T.
// The GCO-Closure CL_T(S) is the smallest set of events such that
// (a) S ⊆ Cl_T
// (b) for every (k,q,l,o) in T and (k',q',l',o') in Cl_T(S) where o is a commit event, q = q' and k < k' we have that (k, q, l, o) in Cl_T, and
// (c) for every (k,q,l,go(p')) in T and (k',q',l',o') with q' = p' we have that (k,q,l,go(p')) in Cl_t(S)
//
// THe Stable closure CL is then CL_T(S) \ S

func calcuateStableClosure(s Constraint) Constraint {
	t := a_base.MainTrace
	tIter := t.AsIterator()

	cl := NewConstraint()

	maxTs := make(map[int]int)
	forks := make(map[int]bool)

	// (b)
	for _, scElem := range s.Elems {
		rout := scElem.RoutineID()
		ts := scElem.T(trace.Commit)
		maxTs[rout] = max(maxTs[rout], ts)

		if e, ok := scElem.(*trace.ElementFork); ok {
			forks[e.ResourceID()] = true
		}
	}

	// (c)
	for _, scElem := range s.Elems {
		rout := scElem.RoutineID()
		if forks[rout] {
			continue
		}
		forkElem := t.GetRoutineTrace(rout).Fork()
		cl.Add(forkElem)

		forks[rout] = true
		maxTs[rout] = max(maxTs[rout], forkElem.T(trace.Commit))
	}

	newElemAdded := true
	for newElemAdded {
		newElemAdded = false

		tIter.Reset()

		// (b)
		for elem := tIter.Next(); elem != nil; elem = tIter.Next() {
			rout := elem.RoutineID()
			ts := elem.T(trace.Commit)
			if ts < maxTs[rout] {
				cl.Add(elem)
				newElemAdded = true
			}
		}

		// (c)
		for rout := range maxTs {
			if forks[rout] {
				continue
			}
			forkElem := t.GetRoutineTrace(rout).Fork()
			cl.Add(forkElem)

			forks[rout] = true
			maxTs[rout] = max(maxTs[rout], forkElem.T(trace.Commit))
			newElemAdded = true
		}
	}

	return cl
}
