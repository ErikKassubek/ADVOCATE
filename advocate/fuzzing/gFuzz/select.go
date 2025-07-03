// Copyright (c) 2024 Erik Kassubek
//
// File: select.go
// Brief: File for the selects for fuzzing
//
// Author: Erik Kassubek
// Created: 2024-12-04
//
// License: BSD-3-Clause

package gFuzz

import (
	fuzzingdata "advocate/fuzzing/data"
	"advocate/trace"
	"sort"
)

// Add a select to selectInfoTrace
//
// Parameter:
//   - e *trace.TraceElementSelect: the select trace element to add
func AddFuzzingSelect(e *trace.TraceElementSelect) {
	fs := fuzzingdata.FuzzingSelect{
		Id:              e.GetReplayID(),
		T:               e.GetTPost(),
		ChosenCase:      e.GetChosenIndex(),
		NumberCases:     len(e.GetCases()),
		ContainsDefault: e.GetContainsDefault(),
		CasiWithPos:     e.GetCasiWithPosPartner(),
	}

	SelectInfoTrace[fs.Id] = append(SelectInfoTrace[fs.Id], fs)
	NumberSelects++
}

// Sort the list of occurrences of each select by the time value
func SortSelects() {
	for key := range SelectInfoTrace {
		sort.Slice(SelectInfoTrace[key], func(i, j int) bool {
			return SelectInfoTrace[key][i].T < SelectInfoTrace[key][j].T
		})
	}
}
