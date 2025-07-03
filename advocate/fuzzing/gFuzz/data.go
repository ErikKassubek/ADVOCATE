// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data for gFuzz
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package gFuzz

import (
	"advocate/fuzzing/data"
	"advocate/utils/helper"
)

const (
	maxRunPerMut = 2
)

var (
	maxScore = 0.0

	// Info for the current trace
	ChannelInfoTrace = make(map[int]FuzzingChannel)          // localID -> fuzzingChannel
	PairInfoTrace    = make(map[string]FuzzingPair)          // posSend-posRecv -> fuzzing pair
	SelectInfoTrace  = make(map[string][]data.FuzzingSelect) // id -> []fuzzingSelects
	NumberSelects    = 0
	NumberClose      = 0

	ChannelInfoFile              = make(map[string]FuzzingChannel) // globalID -> fuzzingChannel
	PairInfoFile                 = make(map[string]FuzzingPair)    // posSend-noPrintosRecv -> fuzzing pair
	SelectInfoFile               = make(map[string][]int)          // globalID -> executed casi
	NumberSelectCasesWithPartner = 0
)

func ClearData() {
	maxScore = 0.0
	ChannelInfoTrace = make(map[int]FuzzingChannel)
	PairInfoTrace = make(map[string]FuzzingPair)
	SelectInfoTrace = make(map[string][]data.FuzzingSelect)

	NumberSelects = 0
	NumberClose = 0

	// Info from the file/the previous runs
	ChannelInfoFile = make(map[string]FuzzingChannel) // globalID -> fuzzingChannel
	PairInfoFile = make(map[string]FuzzingPair)       // posSend-noPrintosRecv -> fuzzing pair
	SelectInfoFile = make(map[string][]int)           // globalID -> executed casi
	NumberSelectCasesWithPartner = 0
}

// Merge the close information for a channel from a trace into the internal
//
// Parameter:
//   - trace closeInfo: info from the last recorded run
//   - file closeInfo: stored close info
//
// Returns:
//   - closeInfo: the new close info for the channel
func MergeCloseInfo(trace closeInfo, file closeInfo) closeInfo {
	if trace != file {
		return Sometimes
	}
	return file
}

// For each channel merge the close info from the last run into the
// internal close info for all ever executed channel close
func MergeTraceInfoIntoFileInfo() {
	// channel info
	for _, cit := range ChannelInfoTrace {
		if cif, ok := ChannelInfoFile[cit.GlobalID]; !ok {
			ChannelInfoFile[cit.GlobalID] = cit
		} else {
			ChannelInfoFile[cit.GlobalID] = FuzzingChannel{cit.GlobalID, 0,
				MergeCloseInfo(cif.CloseInfo, cit.CloseInfo),
				cit.QSize, max(cif.MaxQCount, cit.MaxQCount)}
		}
	}

	// pair info
	for id, pit := range PairInfoTrace {
		if pif, ok := PairInfoFile[id]; !ok {
			PairInfoFile[id] = pit
		} else {
			npr := float64(data.NumberOfPreviousRuns)
			pif.Com = (npr*pif.Com + pit.Com) / (npr + 1)
			PairInfoFile[id] = pif
		}
	}

	// select info
	for id, sits := range SelectInfoTrace {
		if _, ok := SelectInfoFile[id]; !ok {
			SelectInfoFile[id] = make([]int, 0)
		}

		for _, sit := range sits {
			SelectInfoFile[id] = helper.AddIfNotContains(SelectInfoFile[id], sit.ChosenCase)
		}
	}
}
