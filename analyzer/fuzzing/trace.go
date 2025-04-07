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
 * Parse the current trace and record all relevant data
 */
func ParseTrace(trace map[int][]analysis.TraceElement) {
	// clear current order for gFuzz
	selectInfoTrace = make(map[string][]fuzzingSelect)

	// clear chains for goPie
	schedulingChains = make([]chain, 0)
	currentChain = newChain()
	lastRoutine = -1

	for _, routine := range trace {
		if fuzzingModeGoPie {
			calculateRelRule1(routine)
		}
		for _, elem := range routine {
			if elem.GetTPost() == 0 {
				continue
			}

			switch e := elem.(type) {
			case *analysis.TraceElementFork:
				parseFork(e)
			case *analysis.TraceElementNew:
				parseNew(e)
			case *analysis.TraceElementChannel:
				parseChannelOp(e, -2) // -2: not part of select
			case *analysis.TraceElementSelect:
				parseSelectOp(e)
			case *analysis.TraceElementMutex:
				parseMutexOp(e)
			}
		}
	}

	if fuzzingModeGoPie {
		calculateRelRule2()
		calculateRelRule3And4()
	}

	sortSelects()

	numberSelectCasesWithPartner = analysis.GetNumberSelectCasesWithPartner()
}

func parseFork(elem *analysis.TraceElementFork) {
	if fuzzingModeGoPie {
		addElemToChain(elem)
	}
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
	if fuzzingModeGoPie {
		calculateRelRule2AddElem(elem)
		addElemToChain(elem)
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
	if fuzzingModeGoPie {
		calculateRelRule2AddElem(elem)
		addElemToChain(elem)
	}
}

func parseMutexOp(elem *analysis.TraceElementMutex) {
	if fuzzingModeGoPie {
		calculateRelRule2AddElem(elem)
		addElemToChain(elem)
	}
}
