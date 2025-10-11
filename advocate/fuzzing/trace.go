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
	"advocate/utils/control"
	"advocate/utils/log"
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

		if control.CheckCanceled() {
			return
		}

		if data.FuzzingModeGoPie {
			gopie.CalculateRelRule1(routine)
		}

		for _, elem := range routine {

			if control.CheckCanceled() {
				return
			}

			if data.IgnoreFuzzing(elem, false) {
				continue
			}

			if data.FuzzingModeGoPie && !data.UseHBInfoFuzzing && data.CanBeAddedToChain(elem) {
				gopie.CalculateRelRule2AddElem(elem)
			}

			if elem.GetTPost() == 0 {
				continue
			}

			switch e := elem.(type) {
			case *trace.ElementNew:
				if data.FuzzingModeGFuzz {
					parseNew(e)
				}
			case *trace.ElementChannel:
				if data.FuzzingModeGFuzz {
					parseChannelOp(e, -2) // -2: not part of select
				}
			case *trace.ElementSelect:
				if data.FuzzingModeGFuzz {
					parseSelectOp(e)
				}
			}
		}
	}

	if data.FuzzingModeGoPie && gopie.CurrentChain.Len() != 0 {
		gopie.SchedulingChains = append(gopie.SchedulingChains, gopie.CurrentChain)
		gopie.CurrentChain = gopie.NewChain()
	}

	if data.FuzzingModeGoPie && !data.UseHBInfoFuzzing {
		gopie.CalculateRelRule2And4()
		if control.CheckCanceled() {
			return
		}
		gopie.CalculateRelRule3()
	}

	if control.CheckCanceled() {
		return
	}

	if data.FuzzingModeGFuzz {
		gfuzz.SortSelects()

		gfuzz.NumberSelectCasesWithPartner = anadata.NumberSelectCasesWithPartner
	}
}

// Parse a new elem element.
// For now only channels are considered
// Add the corresponding info into FuzzingChannel
func parseNew(elem *trace.ElementNew) {
	// only process channels
	if elem.GetType(true) != trace.NewChannel {
		log.Important("Unexpected new on: ", elem.GetType(true))
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
		op := elem.GetType(true)

		// close -> update channelInfoTrace
		switch op {
		case trace.ChannelClose:
			e := gfuzz.ChannelInfoTrace[elem.GetID()]
			e.CloseInfo = gfuzz.Always // before is always unknown
			gfuzz.ChannelInfoTrace[elem.GetID()] = e
			gfuzz.NumberClose++
		case trace.ChannelSend:
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
