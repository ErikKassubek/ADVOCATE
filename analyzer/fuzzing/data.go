// Copyright (c) 2024 Erik Kassubek
//
// File: data.go
// Brief: File to define and contain the fuzzing data
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package fuzzing

import (
	"analyzer/analysis"
	"analyzer/utils"
)

/*
 * Info for a channel wether it was closed in all runs,
 * never closed or in some runs closed and in others not
 */
type closeInfo string

const (
	always    closeInfo = "a"
	never     closeInfo = "n"
	sometimes closeInfo = "s"
)

var (
	numberOfPreviousRuns     = 0
	numberOfWrittenGoPieMuts = 0
	maxGFuzzScore            = 0.0
	maxGoPieScore            = 0
	// Info for the current trace
	channelInfoTrace = make(map[int]fuzzingChannel)     // localID -> fuzzingChannel
	pairInfoTrace    = make(map[string]fuzzingPair)     // posSend-posRecv -> fuzzing pair
	selectInfoTrace  = make(map[string][]fuzzingSelect) // id -> []fuzzingSelects
	numberSelects    = 0
	numberClose      = 0
	elemsByID        = make(map[int][]analysis.TraceElement) // id -> chan/sel/mutex elem
	// Info from the file/the previous runs
	channelInfoFile              = make(map[string]fuzzingChannel) // globalID -> fuzzingChannel
	pairInfoFile                 = make(map[string]fuzzingPair)    // posSend-noPrintosRecv -> fuzzing pair
	selectInfoFile               = make(map[string][]int)          // globalID -> executed casi
	numberSelectCasesWithPartner = 0

	alreadyDelayedElems = make(map[string][]int)

	useHBInfoFuzzing = true
	runFullAnalysis  = true
)

/*
 * For each channel that has ever been created, store the
 * following information:
 *   globalId: file:line of creation with new
 *   localId: id in this run
 *   qSize: buffer size of the channel
 *   maxQSize: maximum buffer fullness over all runs
 *   whether the channel has always/never/sometimes been closed
 */
type fuzzingChannel struct {
	globalID  string
	localID   int
	closeInfo closeInfo
	qSize     int
	maxQCount int
}

/*
 * For each pair of channel operations, that have communicated, store the following information:
 *    sendID: file:line:caseSend of the send
 *      caseSend: If the send is in a select, the case ID, otherwise 0
 *    recvID: file:line:Recv of the recv
 *      caseRecv: If the recv is in a select, the case ID, otherwise 0
 *    chanID: local ID of the channel
 *    sendSel: id of the select case, if not part of select: -2
 *    recvSel: id of the select case, if not part of select: -2
 *    com: number of communication in this run of avg of communications over all runs
 */
type fuzzingPair struct {
	sendID  string
	recvID  string
	chanID  int
	sendSel int
	recvSel int
	com     float64
}

/*
 * Merge the close information for a channel from a trace into the internal
 * Args:
 * 	trace (closeInfo): info from the last recorded run
 * 	file (closeInfo): stored close info
 * Returns:
 * 	closeInfo: the new close info for the channel
 */
func mergeCloseInfo(trace closeInfo, file closeInfo) closeInfo {
	if trace != file {
		return sometimes
	}
	return file
}

/*
 * For each channel merge the close info from the last run into the
 * internal close info for all ever executed channel close
 */
func mergeTraceInfoIntoFileInfo() {
	// channel info
	for _, cit := range channelInfoTrace {
		if cif, ok := channelInfoFile[cit.globalID]; !ok {
			channelInfoFile[cit.globalID] = cit
		} else {
			channelInfoFile[cit.globalID] = fuzzingChannel{cit.globalID, 0,
				mergeCloseInfo(cif.closeInfo, cit.closeInfo),
				cit.qSize, max(cif.maxQCount, cit.maxQCount)}
		}
	}

	// pair info
	for id, pit := range pairInfoTrace {
		if pif, ok := pairInfoFile[id]; !ok {
			pairInfoFile[id] = pit
		} else {
			npr := float64(numberOfPreviousRuns)
			pif.com = (npr*pif.com + pit.com) / (npr + 1)
			pairInfoFile[id] = pif
		}
	}

	// select info
	for id, sits := range selectInfoTrace {
		if _, ok := selectInfoFile[id]; !ok {
			selectInfoFile[id] = make([]int, 0)
		}

		for _, sit := range sits {
			selectInfoFile[id] = utils.AddIfNotContains(selectInfoFile[id], sit.chosenCase)
		}
	}
}

/*
 * Reset the fuzzing data that is unique for each run
 */
func clearData() {
	// clear the trace data
	channelInfoTrace = make(map[int]fuzzingChannel)
	pairInfoTrace = make(map[string]fuzzingPair)
	selectInfoTrace = make(map[string][]fuzzingSelect)
	elemsByID = make(map[int][]analysis.TraceElement)
}

/*
 * Reset the fuzzing data that is unique for each test but used for each fuzzing
 * run of a test
 */
func clearDataFull() {
	numberOfPreviousRuns = 0
	maxGFuzzScore = 0.0
	// Info for the current trace
	channelInfoTrace = make(map[int]fuzzingChannel)    // localID -> fuzzingChannel
	pairInfoTrace = make(map[string]fuzzingPair)       // posSend-posRecv -> fuzzing pair
	selectInfoTrace = make(map[string][]fuzzingSelect) // id -> []fuzzingSelects
	numberSelects = 0
	numberClose = 0
	// Info from the file/the previous runs
	channelInfoFile = make(map[string]fuzzingChannel) // globalID -> fuzzingChannel
	pairInfoFile = make(map[string]fuzzingPair)       // posSend-noPrintosRecv -> fuzzing pair
	selectInfoFile = make(map[string][]int)           // globalID -> executed casi
	numberSelectCasesWithPartner = 0

	alreadyDelayedElems = make(map[string][]int)
}
