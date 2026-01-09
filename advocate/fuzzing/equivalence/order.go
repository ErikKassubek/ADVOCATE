// Copyright (c) 2025 Erik Kassubek
//
// File: graph.go
// Brief: Create the hb graph for a mutation
//
// Author: Erik Kassubek
// Created: 2025-12-08
//
// License: BSD-3-Clause

package equivalence

import (
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"fmt"
	"sort"
	"strings"
)

func (this *TraceEq) BuildCanonicalSignature(shared map[int]bool, full bool) string {
	if !this.vcHaveBeenCalc {
		this.calcVc()
	}
	return this.calcSignature(shared, full)
}

func (this *TraceEq) calcVc() {
	// TODO: use local variables when calculating HB
	for _, elem := range this.trace {
		switch e := elem.(type) {
		case *trace.ElementFork:
			vc.UpdateHBFork(e)
		case *trace.ElementAtomic:
			vc.UpdateHBAtomic(e)
		case *trace.ElementChannel:
			vc.UpdateHBChannel(e)
		case *trace.ElementSelect:
			vc.UpdateHBSelect(e)
		case *trace.ElementCond:
			vc.UpdateHBCond(e)
		case *trace.ElementMutex:
			vc.UpdateHBMutex(e, false)
		case *trace.ElementOnce:
			vc.UpdateHBOnce(e)
		case *trace.ElementWait:
			vc.UpdateHBWait(e)
		}
	}
	this.vcHaveBeenCalc = true
}

func (this *TraceEq) calcSignature(shared map[int]bool, full bool) string {

	if full && this.fullSig != "" {
		return this.fullSig
	}

	events := make([]trace.Element, 0)
	for _, e := range this.trace {
		if shared[e.GetID()] {
			events = append(events, e)
		}
	}

	n := len(events)
	if n == 0 {
		return ""
	}

	k := events[0].GetVC().GetSize() // number of routines

	// use stable sort to sort events according to vector clocks

	type E struct {
		id int
		vc []int
	}

	es := make([]E, n)
	for i, e := range events {
		es[i] = E{
			id: e.GetID(),
			vc: e.GetVC().AsSlice(),
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		hbRel := clock.GetHappensBeforeSlice(&es[i].vc, &es[j].vc)
		switch hbRel {
		case hb.Before:
			return true
		case hb.After:
			return false
		default:
			return es[i].id < es[j].id
		}
	})

	norm := make([][]int, n)
	for i := 0; i < n; i++ {
		norm[i] = make([]int, k)
	}

	for r := 0; r < k; r++ {
		// Collect all values of routine r in sorted order
		vals := make(map[int]bool)
		for i := 0; i < n; i++ {
			vals[i] = true
		}

		// Map unique values to ranks
		uniq := make([]int, 0, len(vals))
		for v := range vals {
			uniq = append(uniq, v)
		}
		sort.Ints(uniq)

		rank := make(map[int]int)
		for idx, v := range uniq {
			rank[v] = idx + 1
		}

		// Assign normalized rank
		for i := 0; i < n; i++ {
			norm[i][r] = rank[es[i].vc[r]]
		}
	}

	// generate canonical signature
	var sb strings.Builder
	for i := 0; i < n; i++ {
		for r := 0; r < k; r++ {
			sb.WriteString(fmt.Sprintf("%d", norm[i][r]))
			if r+1 < k {
				sb.WriteByte('-')
			}
		}
		if i+1 < n {
			sb.WriteByte('|')
		}
	}
	res := sb.String()
	if full {
		this.fullSig = res
	}
	return res
}
