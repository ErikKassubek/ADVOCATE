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

import (
	"analyzer/analysis"
)

// TODO: maybe move this directly into the analysis

/*
 * Parse the trace and record all relevant data
 * Args:
 * 	trace (*trace *analysis.Trace): The trace to parse
 */
func ParseTrace(trace *analysis.Trace) {
	// clear current order for gFuzz
	selectInfoTrace = make(map[string][]fuzzingSelect)

	// clear chains for goPie
	schedulingChains = make([]chain, 0)
	currentChain = newChain()
	lastRoutine = -1

	for _, routine := range trace.GetTraces() {
		if fuzzingModeGoPie {
			calculateRelRule1(routine)
		}

		for _, elem := range routine {
			if ignoreFuzzing(elem) {
				continue
			}

			if fuzzingModeGoPie {
				calculateRelRule2AddElem(elem)
				addElemToChain(elem)
			}

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

	if fuzzingModeGoPie && currentChain.len() != 0 {
		schedulingChains = append(schedulingChains, currentChain)
		currentChain = newChain()
	}

	if fuzzingModeGoPie {
		calculateRelRule2()
		calculateRelRule3And4()
	}

	sortSelects()

	numberSelectCasesWithPartner = analysis.GetNumberSelectCasesWithPartner()
}

/*
 * For the creation of mutations we ignore all elements that do not directly
 * correspond to relevant operations. Those are , replay, routineEnd
 * Args:
 * 	elem (*analysis.TraceElementFork): The element to check
 * Returns:
 * 	True if the element is of one of those types, false otherwise
 */
func ignoreFuzzing(elem analysis.TraceElement) bool {
	t := elem.GetObjType(false)
	return t == analysis.ObjectTypeNew || t == analysis.ObjectTypeReplay || t == analysis.ObjectTypeRoutineEnd
}

/*
 * Parse a new elem element.
 * For now only channels are considered
 * Add the corresponding info into fuzzingChannel
 */
func parseNew(elem *analysis.TraceElementNew) {
	// only process channels
	if elem.GetObjType(true) != "NC" {
		return
	}

	if fuzzingModeGFuzz {
		fuzzingElem := fuzzingChannel{
			globalID:  elem.GetPos(),
			localID:   elem.GetID(),
			closeInfo: never,
			qSize:     elem.GetNum(),
			maxQCount: 0,
		}

		channelInfoTrace[fuzzingElem.localID] = fuzzingElem
	}
}

/*
 * Parse a channel operations.
 * If the operation is a close, update the data in channelInfoTrace
 * If it is an send, add it to pairInfoTrace
 * If it is an recv, it is either tPost = 0 (ignore) or will be handled by the send
 * selID is the case id if it is a select case, -2 otherwise
 */
func parseChannelOp(elem *analysis.TraceElementChannel, selID int) {

	if fuzzingModeGFuzz {
		op := elem.GetObjType(true)

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
			chanID := elem.GetID()

			if recv != nil {
				sendPos := elem.GetPos()
				recvPos := recv.GetPos()
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
			}

			channelNew := channelInfoTrace[chanID]
			channelNew.maxQCount = max(channelNew.maxQCount, elem.GetQCount())
		}
	}
}

func parseSelectOp(elem *analysis.TraceElementSelect) {
	if fuzzingModeGFuzz {
		addFuzzingSelect(elem)

		if elem.GetChosenDefault() {
			return
		}
		parseChannelOp(elem.GetChosenCase(), elem.GetChosenIndex())
	}
}
