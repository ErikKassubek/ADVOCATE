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
	"analyzer/trace"
)

var currentTrace *trace.Trace

// ParseTrace parses the trace and record all relevant data
//
// Parameter:
//   - tr *trace *analysis.Trace: The trace to parse
func ParseTrace(tr *trace.Trace) {
	currentTrace = tr

	// clear current order for gFuzz
	selectInfoTrace = make(map[string][]fuzzingSelect)

	// clear chains for goPie
	schedulingChains = make([]chain, 0)
	currentChain = newChain()
	lastRoutine = -1

	for _, routine := range tr.GetTraces() {
		if fuzzingModeGoPie {
			calculateRelRule1(routine)
		}

		for _, elem := range routine {
			if ignoreFuzzing(elem) {
				continue
			}

			if fuzzingModeGoPie && canBeAddedToChain(elem) {
				calculateRelRule2AddElem(elem)
			}

			if elem.GetTPost() == 0 {
				continue
			}

			switch e := elem.(type) {
			case *trace.TraceElementNew:
				parseNew(e)
			case *trace.TraceElementChannel:
				parseChannelOp(e, -2) // -2: not part of select
			case *trace.TraceElementSelect:
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

// Decides if an element can be added to a scheduling chain
// For GoPie without improvements (!useHBInfoFuzzing) those are only mutex and channel (incl. select)
// With improvements those are all not ignored fuzzing elements
//
// Parameter:
//   - elem analysis.TraceElement: Element to check
//
// Returns:
//   - true if it can be added to a scheduling chain, false otherwise
func canBeAddedToChain(elem trace.TraceElement) bool {
	if fuzzingMode == GoPie {
		// for standard GoPie, only mutex, channel and select operations are considered
		t := elem.GetObjType(false)
		return t == trace.ObjectTypeMutex || t == trace.ObjectTypeChannel || t == trace.ObjectTypeSelect
	}

	return !ignoreFuzzing(elem)
}

// For the creation of mutations we ignore all elements that do not directly
// correspond to relevant operations. Those are , replay, routineEnd
//
// Parameter:
//   - elem *trace.TraceElementFork: The element to check
//
// Returns:
//   - True if the element is of one of those types, false otherwise
func ignoreFuzzing(elem trace.TraceElement) bool {
	t := elem.GetObjType(false)
	return t == trace.ObjectTypeNew || t == trace.ObjectTypeReplay || t == trace.ObjectTypeRoutineEnd
}

// Parse a new elem element.
// For now only channels are considered
// Add the corresponding info into fuzzingChannel
func parseNew(elem *trace.TraceElementNew) {
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

// Parse a channel operations.
// If the operation is a close, update the data in channelInfoTrace
// If it is an send, add it to pairInfoTrace
// If it is an recv, it is either tPost = 0 (ignore) or will be handled by the send
// selID is the case id if it is a select case, -2 otherwise
func parseChannelOp(elem *trace.TraceElementChannel, selID int) {

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

// Parse a select operation in the trace for fuzzing
//
// Parameter:
//   - elem *analysis.TraceElementSelect: the select element
func parseSelectOp(elem *trace.TraceElementSelect) {
	if fuzzingModeGFuzz {
		addFuzzingSelect(elem)

		if elem.GetChosenDefault() {
			return
		}
		parseChannelOp(elem.GetChosenCase(), elem.GetChosenIndex())
	}
}
