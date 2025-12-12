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

func (this *TraceEq) BuildCanonicalSignature() {
	this.calcVc()
	this.signature = this.calcSignature()
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
}

func (this *TraceEq) calcSignature() string {
	n := len(this.trace)
	if n == 0 {
		return ""
	}

	k := this.trace[0].GetVC().GetSize() // number of routines

	// use stable sort to sort events according to vector clocks

	type eventWithVC struct {
		id int
		vc []int
	}

	events := make([]eventWithVC, n)
	for i := 0; i < n; i++ {
		events[i] = eventWithVC{
			id: this.trace[i].GetID(),
			vc: this.trace[i].GetVC().AsSlice(),
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		hbRel := clock.GetHappensBeforeSlice(&events[i].vc, &events[j].vc)
		switch hbRel {
		case hb.Before:
			return true
		case hb.After:
			return false
		default:
			return events[i].id < events[j].id
		}
	})

	normalizedVCs := make([][]int, n)
	for i := 0; i < n; i++ {
		normalizedVCs[i] = make([]int, k)
	}

	for r := 0; r < k; r++ {
		// Collect all values of routine r in sorted order
		values := make([]int, n)
		for i := 0; i < n; i++ {
			values[i] = events[i].vc[r]
		}

		// Map unique values to ranks
		valueSet := make(map[int]bool)
		for _, v := range values {
			valueSet[v] = true
		}
		uniqueVals := make([]int, 0, len(valueSet))
		for v := range valueSet {
			uniqueVals = append(uniqueVals, v)
		}
		sort.Ints(uniqueVals)

		valToRank := make(map[int]int)
		for idx, v := range uniqueVals {
			valToRank[v] = idx + 1
		}

		// Assign normalized rank
		for i := 0; i < n; i++ {
			normalizedVCs[i][r] = valToRank[events[i].vc[r]]
		}
	}

	// generate canonical signature
	sigParts := make([]string, n)
	for i := 0; i < n; i++ {
		compStrings := make([]string, k)
		for r := 0; r < k; r++ {
			compStrings[r] = fmt.Sprintf("%d", normalizedVCs[i][r])
		}
		sigParts[i] = strings.Join(compStrings, "-")
	}

	return strings.Join(sigParts, "|")
}
