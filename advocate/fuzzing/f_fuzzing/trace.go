// Copyright (c) 2024 Erik Kassubek
//
// File: trace.go
// Brief: Function to parse the trace and get all relevant information
//
// Author: Erik Kassubek
// Created: 2024-11-29
//
// License: BSD-3-Clause

package f_fuzzing

import (
	"advocate/analysis/a_base"
	"advocate/fuzzing/f_base"
	"advocate/fuzzing/f_gfuzz"
	"advocate/fuzzing/f_gopie"
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
	f_gfuzz.SelectInfoTrace = make(map[string][]f_base.FuzzingSelect)

	// clear chains for goPie
	f_gopie.SchedulingChains = make([]f_base.Constraint, 0)
	f_gopie.CurrentChain = f_base.NewConstraint()
	f_gopie.LastRoutine = -1

	for _, routine := range tr.GetTraces() {

		if control.WasCanceled() {
			return
		}

		if f_base.FuzzingModeGoPie {
			f_gopie.CalculateRelRule1(routine)
		}

		for _, elem := range routine {

			if control.WasCanceled() {
				return
			}

			if f_base.IgnoreFuzzing(elem, false) {
				continue
			}

			if f_base.FuzzingModeGoPie && !f_base.UseHBInfoFuzzing && f_base.CanBeAddedToConstraint(elem) {
				f_gopie.CalculateRelRule2AddElem(elem)
			}

			if !elem.Committed() {
				continue
			}

			switch e := elem.(type) {
			case *trace.ElementAlloc:
				if f_base.FuzzingModeGFuzz {
					parseNew(e)
				}
			case *trace.ElementChannel:
				if f_base.FuzzingModeGFuzz {
					parseChannelOp(e, -2) // -2: not part of select
				}
			case *trace.ElementSelect:
				if f_base.FuzzingModeGFuzz || f_base.FuzzingModeGuided {
					parseSelectOp(e)
				}
			}
		}
	}

	if f_base.FuzzingModeGoPie && f_gopie.CurrentChain.Len() != 0 {
		f_gopie.SchedulingChains = append(f_gopie.SchedulingChains, f_gopie.CurrentChain)
		f_gopie.CurrentChain = f_base.NewConstraint()
	}

	if f_base.FuzzingModeGoPie && !f_base.UseHBInfoFuzzing {
		f_gopie.CalculateRelRule2And4()
		if control.WasCanceled() {
			return
		}
		f_gopie.CalculateRelRule3()
	}

	if control.WasCanceled() {
		return
	}

	if f_base.FuzzingModeGFuzz {
		f_gfuzz.SortSelects()

		f_gfuzz.NumberSelectCasesWithPartner = a_base.NumberSelectCasesWithPartner
	}
}

// Parse a new elem element.
// For now only channels are considered
// Add the corresponding info into FuzzingChannel
func parseNew(elem *trace.ElementAlloc) {
	// only process channels
	if elem.GetType(true) != trace.NewChannel {
		log.Important("Unexpected new on: ", elem.GetType(true))
		return
	}

	if f_base.FuzzingModeGFuzz {
		fuzzingElem := f_gfuzz.FuzzingChannel{
			GlobalID:  elem.GetPos(),
			LocalID:   elem.GetObjId(),
			CloseInfo: f_gfuzz.Never,
			QSize:     elem.GetNum(),
			MaxQCount: 0,
		}

		f_gfuzz.ChannelInfoTrace[fuzzingElem.LocalID] = fuzzingElem
	}
}

// Parse a channel operations.
// If the operation is a close, update the data in channelInfoTrace
// If it is an send, add it to pairInfoTrace
// If it is an recv, it is either tPost = 0 (ignore) or will be handled by the send
// selID is the case id if it is a select case, -2 otherwise
func parseChannelOp(elem *trace.ElementChannel, selID int) {

	if f_base.FuzzingModeGFuzz {
		op := elem.GetType(true)

		// close -> update channelInfoTrace
		switch op {
		case trace.ChannelClose:
			e := f_gfuzz.ChannelInfoTrace[elem.GetObjId()]
			e.CloseInfo = f_gfuzz.Always // before is always unknown
			f_gfuzz.ChannelInfoTrace[elem.GetObjId()] = e
			f_gfuzz.NumberClose++
		case trace.ChannelSend:
			if !elem.Committed() {
				return
			}

			recv := elem.GetPartner()
			chanID := elem.GetObjId()

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

				if e, ok := f_gfuzz.PairInfoTrace[key]; ok {
					e.Com++
					f_gfuzz.PairInfoTrace[key] = e
				} else {
					fp := f_gfuzz.FuzzingPair{
						ChanID:  chanID,
						Com:     1,
						SendSel: selID,
						RecvSel: selIDRecv,
					}
					f_gfuzz.PairInfoTrace[key] = fp
				}
			}

			channelNew := f_gfuzz.ChannelInfoTrace[chanID]
			channelNew.MaxQCount = max(channelNew.MaxQCount, elem.GetQCount())
		}
	}
}

// Parse a select operation in the trace for fuzzing
//
// Parameter:
//   - elem *analysis.TraceElementSelect: the select element
func parseSelectOp(elem *trace.ElementSelect) {
	if f_base.FuzzingModeGFuzz {
		f_gfuzz.AddFuzzingSelect(elem)

		if elem.GetChosenDefault() {
			return
		}
		parseChannelOp(elem.GetChosenCase(), elem.GetChosenIndex())
	}
}
