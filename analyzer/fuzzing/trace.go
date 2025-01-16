// Copyright (c) 2024 Erik Kassubek
//
// File: trace.go
// Brief: Function to parse the trace and get all relevant information
//
// Author: Erik Kassubek
// Created: 2024-11-29
//
// License: BSD-3-Clause

package fuzzing

import "analyzer/analysis"

/*
 * Parse the current trace and record all relevant data
 */
func parseTrace() {
	// clear current order
	allSelects = make(map[string][]fuzzingSelect)

	for _, trace := range *analysis.GetTraces() {
		for _, elem := range trace {
			if elem.GetTPost() == 0 {
				continue
			}

			switch e := elem.(type) {
			case *analysis.TraceElementNew:
				parseNew(e)
			case *analysis.TraceElementChannel:
				parseChannelOp(e, -2) // -2: not part of select
			case *analysis.TraceElementSelect:
				parseSelectOp(e)
			}
		}
	}

	sortSelects()
}

/*
 * Parse a new elem element.
 * For now only channels are considered
 * Add the corresponding info into fuzzingChannel
 */
func parseNew(elem *analysis.TraceElementNew) {
	// only process channels
	if elem.GetObjType() != "NC" {
		return
	}

	fuzzingElem := fuzzingChannel{
		globalID:  elem.GetPos(),
		localID:   elem.GetID(),
		closeInfo: never,
		qSize:     elem.GetNum(),
		maxQCount: 0,
	}

	channelInfoTrace[fuzzingElem.localID] = fuzzingElem
}

/*
 * Parse a channel operations.
 * If the operation is a close, update the data in channelInfoTrace
 * If it is an send, add it to pairInfoTrace
 * If it is an recv, it is either tPost = 0 (ignore) or will be handled by the send
 * selID is the case id if it is a select case, -2 otherwise
 */
func parseChannelOp(elem *analysis.TraceElementChannel, selID int) {
	op := elem.GetObjType()

	// close -> update channelInfoTrace
	if op == "CC" {
		e := channelInfoTrace[elem.GetID()]
		e.closeInfo = always // before is always unknown
		channelInfoTrace[elem.GetID()] = e
		numberClose++
	} else if op == "CS" {
		if elem.GetTPost() == 0 {
			return
		}

		recv := elem.GetPartner()
		if recv == nil {
			panic("fuzzing parseChannelOp, send without partner: first run find partner")
		}

		sendPos := elem.GetPos()
		recvPos := recv.GetPos()
		chanID := elem.GetID()
		key := sendPos + "-" + recvPos

		// if receive is a select case
		selIDRecv := -2
		selRecv := recv.GetSelect()
		if selRecv != nil {
			selIDRecv = selRecv.GetChosenIndex()
		}

		if e, ok := pairInfoTrace[key]; ok {
			e.com++
			pairInfoTrace[key] = e
		} else {
			fp := fuzzingPair{
				chanID:  chanID,
				com:     1,
				sendSel: selID,
				recvSel: selIDRecv,
			}
			pairInfoTrace[key] = fp
		}

		channelNew := channelInfoTrace[chanID]
		channelNew.maxQCount = max(channelNew.maxQCount, elem.GetQCount())
	}
}

func parseSelectOp(e *analysis.TraceElementSelect) {
	addFuzzingSelect(e.GetPos(), e.GetTPost(), e.GetChosenIndex(), len(e.GetCases()), e.GetContainsDefault())

	if e.GetChosenDefault() {
		return
	}
	parseChannelOp(e.GetChosenCase(), e.GetChosenIndex())
}
