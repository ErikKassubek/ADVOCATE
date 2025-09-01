// Copyright (c) 2024 Erik Kassubek
//
// File: hbChannel.go
// Brief: Update functions for happens before info for channel operations
//        Some of the update function also start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-27
//
// License: BSD-3-Clause

package elements

import (
	"advocate/analysis/analysis/scenarios"
	"advocate/analysis/data"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/hbcalc"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
)

// UpdateChannel updates the vector clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateChannel(ch *trace.ElementChannel) {
	id := ch.GetID()
	opC := ch.GetOpC()
	oID := ch.GetOID()
	cl := ch.GetClosed()

	// run hold back recv if the send has been processed
	for _, elem := range data.WaitingReceive {
		if elem.GetOID() <= data.MaxOpID[id] {
			if len(data.WaitingReceive) != 0 {
				data.WaitingReceive = data.WaitingReceive[1:]
			}
			UpdateChannel(elem)
		}
	}

	// hold back receive operations, until the send operation is processed
	if ch.IsBuffered() {
		switch opC {
		case trace.SendOp:
			data.MaxOpID[id] = oID
		case trace.RecvOp:
			if oID > data.MaxOpID[id] && !cl {
				data.WaitingReceive = append(data.WaitingReceive, ch)
				return
			}
		}
	}

	hbcalc.UpdateHBChannel(ch)

	if ch.GetTPost() == 0 {
		return
	}

	if ch.IsBuffered() {
		switch opC {
		case trace.SendOp:
			Send(ch, vc.CurrentVC, vc.CurrentWVC)
		case trace.RecvOp:
			if cl { // recv on closed channel
				RecvC(ch, vc.CurrentVC, vc.CurrentWVC, true)
			} else {
				Recv(ch, vc.CurrentVC, vc.CurrentWVC)
			}
		case trace.CloseOp:
			Close(ch)
		default:
			err := "Unknown operation: " + ch.ToString()
			log.Error(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.SendOp:
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				Unbuffered(ch, partner)
				// advance index of receive routine, send routine is already advanced
				data.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					SendC(ch)
				}
			}

		case trace.RecvOp: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				Unbuffered(partner, ch)
				// advance index of receive routine, send routine is already advanced
				data.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					RecvC(ch, vc.CurrentVC, vc.CurrentWVC, false)
				}
			}
		case trace.CloseOp:
			Close(ch)
		default:
			err := "Unknown operation: " + ch.ToString()
			log.Error(err)
		}
	}
}

// UpdateSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateSelect(se *trace.ElementSelect) {
	routine := se.GetRoutine()

	if data.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerSelect(se, vc.CurrentVC[routine])
	}

	hbcalc.UpdateHBSelect(se)

	cases := se.GetCases()

	for _, c := range cases {
		opC := c.GetOpC()
		switch opC {
		case trace.SendOp:
			setChannelAsLastSend(&c)
		case trace.RecvOp:
			setChannelAsLastReceive(&c)
		}
	}

	if data.AnalysisCasesMap[flags.SendOnClosed] {
		chosenIndex := se.GetChosenIndex()
		for i, c := range cases {
			if i == chosenIndex {
				continue
			}

			opC := c.GetOpC()

			if _, ok := data.CloseData[c.GetID()]; ok {
				switch opC {
				case trace.SendOp:
					scenarios.FoundSendOnClosedChannel(&c, false)
				case trace.RecvOp:
					scenarios.FoundReceiveOnClosedChannel(&c, false)
				}
			}
		}
	}

	if data.AnalysisCasesMap[flags.Leak] {
		for _, c := range cases {
			scenarios.CheckForLeakChannelRun(routine, c.GetRoutine(),
				data.ElemWithVc{
					Vc:   se.GetVC().Copy(),
					Elem: se},
				int(c.GetOpC()), c.IsBuffered())
		}
	}
}

// Unbuffered updates and calculates the vector clocks given a send/receive pair on a unbuffered
// channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - routSend int: the route of the sender
//   - routRecv int: the route of the receiver
//   - tID_send string: the position of the send in the program
//   - tID_recv string: the position of the receive in the program
func Unbuffered(sender trace.Element, recv trace.Element) {
	if data.AnalysisCasesMap[flags.ConcurrentRecv] || data.AnalysisFuzzingFlow { // or fuzzing
		switch r := recv.(type) {
		case *trace.ElementChannel:
			scenarios.CheckForConcurrentRecv(r, vc.CurrentVC)
		case *trace.ElementSelect:
			scenarios.CheckForConcurrentRecv(r.GetChosenCase(), vc.CurrentVC)
		}
	}

	if data.AnalysisFuzzingFlow {
		switch s := sender.(type) {
		case *trace.ElementChannel:
			scenarios.GetConcurrentSendForFuzzing(s)
		case *trace.ElementSelect:
			scenarios.GetConcurrentSendForFuzzing(s.GetChosenCase())
		}
	}

	if sender.GetTPost() != 0 && recv.GetTPost() != 0 {
		if data.MostRecentReceive[recv.GetRoutine()] == nil {
			data.MostRecentReceive[recv.GetRoutine()] = make(map[int]data.ElemWithVcVal)
		}
		if data.MostRecentSend[sender.GetRoutine()] == nil {
			data.MostRecentSend[sender.GetRoutine()] = make(map[int]data.ElemWithVcVal)
		}

		// for detection of send on closed
		data.HasSend[sender.GetID()] = true
		data.MostRecentSend[sender.GetRoutine()][sender.GetID()] = data.ElemWithVcVal{
			Elem: sender,
			Vc:   data.MostRecentSend[sender.GetRoutine()][sender.GetID()].Vc.Sync(vc.CurrentVC[sender.GetRoutine()]).Copy(),
			Val:  sender.GetID()}

		// for detection of receive on closed
		data.HasReceived[sender.GetID()] = true
		data.MostRecentReceive[recv.GetRoutine()][sender.GetID()] = data.ElemWithVcVal{Elem: recv,
			Vc:  data.MostRecentReceive[recv.GetRoutine()][sender.GetID()].Vc.Sync(vc.CurrentVC[recv.GetRoutine()]).Copy(),
			Val: sender.GetID(),
		}
	}

	if data.AnalysisCasesMap[flags.SendOnClosed] {
		if _, ok := data.CloseData[sender.GetID()]; ok {
			scenarios.FoundSendOnClosedChannel(sender, true)
		}
	}

	if data.AnalysisCasesMap[flags.MixedDeadlock] {
		scenarios.CheckForMixedDeadlock(sender.GetRoutine(), recv.GetRoutine())
	}

	if data.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(sender, vc.CurrentVC[sender.GetRoutine()], true, false)
		scenarios.CheckForSelectCaseWithPartnerChannel(recv, vc.CurrentVC[recv.GetRoutine()], false, false)
	}

	if data.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(sender.GetRoutine(), sender.GetID(), data.ElemWithVc{Vc: vc.CurrentVC[sender.GetRoutine()].Copy(), Elem: sender}, 0, false)
		scenarios.CheckForLeakChannelRun(recv.GetRoutine(), sender.GetID(), data.ElemWithVc{Vc: vc.CurrentVC[recv.GetRoutine()].Copy(), Elem: recv}, 1, false)
	}
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
func Send(ch *trace.ElementChannel, vc, wVc map[int]*clock.VectorClock) {
	id := ch.GetID()
	routine := ch.GetRoutine()

	if ch.GetTPost() == 0 {
		return
	}

	if data.MostRecentSend[routine] == nil {
		data.MostRecentSend[routine] = make(map[int]data.ElemWithVcVal)
	}

	// for detection of send on closed
	data.HasSend[id] = true
	data.MostRecentSend[routine][id] = data.ElemWithVcVal{
		Elem: ch,
		Vc:   data.MostRecentSend[routine][id].Vc.Sync(vc[routine]).Copy(),
		Val:  id,
	}

	if data.AnalysisCasesMap[flags.SendOnClosed] {
		if _, ok := data.CloseData[id]; ok {
			scenarios.FoundSendOnClosedChannel(ch, true)
		}
	}

	if data.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if data.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, 0, true)
	}

	for i, hold := range data.HoldRecv {
		if hold.Ch.GetID() == id {
			Recv(hold.Ch, hold.Vc, hold.WVc)
			data.HoldRecv = append(data.HoldRecv[:i], data.HoldRecv[i+1:]...)
			break
		}
	}
}

// Recv updates and calculates the vector clocks given a receive on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
func Recv(ch *trace.ElementChannel, vc, wVc map[int]*clock.VectorClock) {
	id := ch.GetID()
	routine := ch.GetRoutine()

	if data.AnalysisCasesMap[flags.ConcurrentRecv] || data.AnalysisFuzzingFlow {
		scenarios.CheckForConcurrentRecv(ch, vc)
	}

	if ch.GetTPost() == 0 {
		return
	}

	if data.MostRecentReceive[routine] == nil {
		data.MostRecentReceive[routine] = make(map[int]data.ElemWithVcVal)
	}

	// for detection of receive on closed
	data.HasReceived[id] = true
	data.MostRecentReceive[routine][id] = data.ElemWithVcVal{
		Elem: ch,
		Vc:   data.MostRecentReceive[routine][id].Vc.Sync(vc[routine]),
		Val:  id,
	}

	if data.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if data.AnalysisCasesMap[flags.MixedDeadlock] {
		routSend := ch.GetPartner().GetRoutine()
		scenarios.CheckForMixedDeadlock(routSend, routine)
	}
	if data.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, 1, true)
	}

	for i, hold := range data.HoldSend {
		if hold.Ch.GetID() == id {
			Send(hold.Ch, hold.Vc, hold.WVc)
			data.HoldSend = append(data.HoldSend[:i], data.HoldSend[i+1:]...)
			break
		}
	}
}

// Close updates and calculates the vector clocks given a close on a channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func Close(ch *trace.ElementChannel) {
	if ch.GetTPost() == 0 {
		return
	}

	routine := ch.GetRoutine()
	id := ch.GetID()

	ch.SetClosed(true)

	if data.AnalysisCasesMap[flags.CloseOnClosed] {
		scenarios.CheckForClosedOnClosed(ch) // must be called before closePos is updated
	}

	data.CloseData[id] = ch

	if data.AnalysisCasesMap[flags.SendOnClosed] || data.AnalysisCasesMap[flags.ReceiveOnClosed] {
		scenarios.CheckForCommunicationOnClosedChannel(ch)
	}

	if data.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerClose(ch, vc.CurrentVC[routine])
	}

	if data.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc.CurrentVC[routine].Copy(),
			Elem: ch,
		}, 2, true)
	}
}

// SendC record an actual send on closed
func SendC(ch *trace.ElementChannel) {
	if data.AnalysisCasesMap[flags.SendOnClosed] {
		scenarios.FoundSendOnClosedChannel(ch, true)
	}
}

// RecvC updates and calculates the vector clocks given a receive on a closed channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
//   - buffered bool: true if the channel is buffered
func RecvC(ch *trace.ElementChannel, vc, wVc map[int]*clock.VectorClock, buffered bool) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetID()
	routine := ch.GetRoutine()

	if data.AnalysisCasesMap[flags.ReceiveOnClosed] {
		scenarios.FoundReceiveOnClosedChannel(ch, true)
	}

	if data.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], false, buffered)
	}

	if data.AnalysisCasesMap[flags.MixedDeadlock] {
		scenarios.CheckForMixedDeadlock(data.CloseData[id].GetRoutine(), routine)
	}

	if data.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, 1, buffered)
	}
}

// setChannelAsLastSend sets the channel as the last send operation.
// Used for not executed select send
//
// Parameter:
//   - id int: the id of the channel
//   - routine int: the route of the operation
//   - vc VectorClock: the vector clock of the operation
//   - tID string: the position of the send in the program
func setChannelAsLastSend(c trace.Element) {
	id := c.GetID()
	routine := c.GetRoutine()

	if data.MostRecentSend[routine] == nil {
		data.MostRecentSend[routine] = make(map[int]data.ElemWithVcVal)
	}
	data.MostRecentSend[routine][id] = data.ElemWithVcVal{
		Elem: c,
		Vc:   c.GetVC(),
		Val:  id,
	}
	data.HasSend[routine] = true
}

// setChannelAsLastReceive sets the channel as the last recv operation.
// Used for not executed select recv
//
// Parameter:
//   - id int: the id of the channel
//   - rout int: the route of the operation
//   - vc VectorClock: the vector clock of the operation
//   - tID string: the position of the recv in the program
func setChannelAsLastReceive(c trace.Element) {
	id := c.GetID()
	routine := c.GetRoutine()

	if data.MostRecentReceive[routine] == nil {
		data.MostRecentReceive[routine] = make(map[int]data.ElemWithVcVal)
	}
	data.MostRecentReceive[routine][id] = data.ElemWithVcVal{
		Elem: c,
		Vc:   c.GetVC(),
		Val:  id,
	}
	data.HasReceived[id] = true
}
