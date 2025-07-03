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
	anadata "advocate/analysis/data"
	"advocate/fuzzing/data"
	"advocate/fuzzing/gfuzz"
	"advocate/fuzzing/gopie"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/memory"
)

var currentTrace *trace.Trace

// ParseTrace parses the trace and record all relevant data
//
// Parameter:
//   - tr *trace *analysis.Trace: The trace to parse
func ParseTrace(tr *trace.Trace) {
	currentTrace = tr

	// clear current order for gFuzz
	gfuzz.SelectInfoTrace = make(map[string][]data.FuzzingSelect)

	// clear chains for goPie
	gopie.SchedulingChains = make([]gopie.Chain, 0)
	gopie.CurrentChain = gopie.NewChain()
	gopie.LastRoutine = -1

	for _, routine := range tr.GetTraces() {
		if data.FuzzingModeGoPie {
			gopie.CalculateRelRule1(routine)
		}

		if memory.WasCanceled() {
			return
		}

		for _, elem := range routine {
			if ignoreFuzzing(elem, false) {
				continue
			}

			if data.FuzzingModeGoPie && canBeAddedToChain(elem) {
				gopie.CalculateRelRule2AddElem(elem)
			}

			if elem.GetTPost() == 0 {
				continue
			}

			switch e := elem.(type) {
			case *trace.ElementNew:
				parseNew(e)
			case *trace.ElementChannel:
				parseChannelOp(e, -2) // -2: not part of select
			case *trace.ElementSelect:
				parseSelectOp(e)
			}

			if memory.WasCanceled() {
				return
			}

		}
	}

	if data.FuzzingModeGoPie && gopie.CurrentChain.Len() != 0 {
		gopie.SchedulingChains = append(gopie.SchedulingChains, gopie.CurrentChain)
		gopie.CurrentChain = gopie.NewChain()
	}

	if data.FuzzingModeGoPie {
		gopie.CalculateRelRule2And4()
		gopie.CalculateRelRule3()
	}

	if memory.WasCanceled() {
		return
	}

	gfuzz.SortSelects()

	gfuzz.NumberSelectCasesWithPartner = anadata.NumberSelectCasesWithPartner
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
func canBeAddedToChain(elem trace.Element) bool {
	if data.FuzzingMode == data.GoPie {
		// for standard GoPie, only mutex, channel and select operations are considered
		t := elem.GetObjType(false)
		return t == trace.ObjectTypeMutex || t == trace.ObjectTypeChannel || t == trace.ObjectTypeSelect
	}

	return !ignoreFuzzing(elem, true)
}

// For the creation of mutations we ignore all elements that do not directly
// correspond to relevant operations. Those are , replay, routineEnd
//
// Parameter:
//   - elem *trace.TraceElementFork: The element to check
//   - ignoreNew bool: if true, new elem is ignored elem, otherwise not
//
// Returns:
//   - True if the element is of one of those types, false otherwise
func ignoreFuzzing(elem trace.Element, ignoreNew bool) bool {
	t := elem.GetObjType(false)
	return (ignoreNew && t == trace.ObjectTypeNew) || t == trace.ObjectTypeReplay || t == trace.ObjectTypeRoutineEnd
}

// Parse a new elem element.
// For now only channels are considered
// Add the corresponding info into FuzzingChannel
func parseNew(elem *trace.ElementNew) {
	// only process channels
	if elem.GetObjType(true) != "NC" {
		log.Important(elem.GetObjType(true))
		return
	}

	if data.FuzzingModeGFuzz {
		fuzzingElem := gfuzz.FuzzingChannel{
			GlobalID:  elem.GetPos(),
			LocalID:   elem.GetID(),
			CloseInfo: gfuzz.Never,
			QSize:     elem.GetNum(),
			MaxQCount: 0,
		}

		gfuzz.ChannelInfoTrace[fuzzingElem.LocalID] = fuzzingElem
	}
}

// Parse a channel operations.
// If the operation is a close, update the data in channelInfoTrace
// If it is an send, add it to pairInfoTrace
// If it is an recv, it is either tPost = 0 (ignore) or will be handled by the send
// selID is the case id if it is a select case, -2 otherwise
func parseChannelOp(elem *trace.ElementChannel, selID int) {

	if data.FuzzingModeGFuzz {
		op := elem.GetObjType(true)

		// close -> update channelInfoTrace
		if op == "CC" {
			e := gfuzz.ChannelInfoTrace[elem.GetID()]
			e.CloseInfo = gfuzz.Always // before is always unknown
			gfuzz.ChannelInfoTrace[elem.GetID()] = e
			gfuzz.NumberClose++
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

				if e, ok := gfuzz.PairInfoTrace[key]; ok {
					e.Com++
					gfuzz.PairInfoTrace[key] = e
				} else {
					fp := gfuzz.FuzzingPair{
						ChanID:  chanID,
						Com:     1,
						SendSel: selID,
						RecvSel: selIDRecv,
					}
					gfuzz.PairInfoTrace[key] = fp
				}
			}

			channelNew := gfuzz.ChannelInfoTrace[chanID]
			channelNew.MaxQCount = max(channelNew.MaxQCount, elem.GetQCount())
		}
	}
}

// Parse a select operation in the trace for fuzzing
//
// Parameter:
//   - elem *analysis.TraceElementSelect: the select element
func parseSelectOp(elem *trace.ElementSelect) {
	if data.FuzzingModeGFuzz {
		gfuzz.AddFuzzingSelect(elem)

		if elem.GetChosenDefault() {
			return
		}
		parseChannelOp(elem.GetChosenCase(), elem.GetChosenIndex())
	}
}
