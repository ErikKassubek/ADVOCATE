// advocate/analysis/analysis/elements/channel.go

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
	"advocate/analysis/baseA"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/hbcalc"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
	"fmt"
)

// UpdateChannel updates the vector clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateChannel(ch *trace.ElementChannel) {
	id := ch.GetID()
	opC := ch.GetType(true)
	oID := ch.GetOID()
	cl := ch.GetClosed()

	// run hold back recv if the send has been processed
	for _, elem := range baseA.WaitingReceive {
		if elem.GetOID() <= baseA.MaxOpID[id] {
			if len(baseA.WaitingReceive) != 0 {
				baseA.WaitingReceive = baseA.WaitingReceive[1:]
			}
			UpdateChannel(elem)
		}
	}

	// hold back receive operations, until the send operation is processed
	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			baseA.MaxOpID[id] = oID
		case trace.ChannelRecv:
			if oID > baseA.MaxOpID[id] && !cl {
				baseA.WaitingReceive = append(baseA.WaitingReceive, ch)
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
		case trace.ChannelSend:
			Send(ch, vc.CurrentVC, vc.CurrentWVC)
		case trace.ChannelRecv:
			if cl { // recv on closed channel
				RecvC(ch, vc.CurrentVC, vc.CurrentWVC, true)
			} else {
				Recv(ch, vc.CurrentVC, vc.CurrentWVC)
			}
		case trace.ChannelClose:
			Close(ch)
		default:
			err := "Unknown operation: " + ch.ToString()
			log.Error(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.ChannelSend:
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				Unbuffered(ch, partner)
				// advance index of receive routine, send routine is already advanced
				baseA.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					SendC(ch)
				}
			}

		case trace.ChannelRecv: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				Unbuffered(partner, ch)
				// advance index of receive routine, send routine is already advanced
				baseA.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					RecvC(ch, vc.CurrentVC, vc.CurrentWVC, false)
				}
			}
		case trace.ChannelClose:
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

	if baseA.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerSelect(se, vc.CurrentVC[routine])
	}

	hbcalc.UpdateHBSelect(se)

	cases := se.GetCases()

	for _, c := range cases {
		opC := c.GetType(true)
		switch opC {
		case trace.ChannelSend:
			setChannelAsLastSend(&c)
		case trace.ChannelRecv:
			setChannelAsLastReceive(&c)
		}
	}

	if baseA.AnalysisCasesMap[flags.SendOnClosed] {
		chosenIndex := se.GetChosenIndex()
		for i, c := range cases {
			if i == chosenIndex {
				continue
			}

			opC := c.GetType(true)

			if _, ok := baseA.CloseData[c.GetID()]; ok {
				switch opC {
				case trace.ChannelSend:
					scenarios.FoundSendOnClosedChannel(&c, false)
				case trace.ChannelRecv:
					scenarios.FoundReceiveOnClosedChannel(&c, false)
				}
			}
		}
	}

	if baseA.AnalysisCasesMap[flags.Leak] {
		for _, c := range cases {
			scenarios.CheckForLeakChannelRun(routine, c.GetRoutine(),
				baseA.ElemWithVc{
					Vc:   se.GetVC().Copy(),
					Elem: se},
				c.GetType(true), c.IsBuffered())
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
	if baseA.AnalysisCasesMap[flags.ConcurrentRecv] || baseA.AnalysisFuzzingFlow { // or fuzzing
		switch r := recv.(type) {
		case *trace.ElementChannel:
			scenarios.CheckForConcurrentRecv(r, vc.CurrentVC)
		case *trace.ElementSelect:
			scenarios.CheckForConcurrentRecv(r.GetChosenCase(), vc.CurrentVC)
		}
	}

	if baseA.AnalysisFuzzingFlow {
		switch s := sender.(type) {
		case *trace.ElementChannel:
			scenarios.GetConcurrentSendForFuzzing(s)
		case *trace.ElementSelect:
			scenarios.GetConcurrentSendForFuzzing(s.GetChosenCase())
		}
	}

	if sender.GetTPost() != 0 && recv.GetTPost() != 0 {
		if baseA.MostRecentReceive[recv.GetRoutine()] == nil {
			baseA.MostRecentReceive[recv.GetRoutine()] = make(map[int]baseA.ElemWithVcVal)
		}
		if baseA.MostRecentSend[sender.GetRoutine()] == nil {
			baseA.MostRecentSend[sender.GetRoutine()] = make(map[int]baseA.ElemWithVcVal)
		}

		// for detection of send on closed
		baseA.HasSend[sender.GetID()] = true
		baseA.MostRecentSend[sender.GetRoutine()][sender.GetID()] = baseA.ElemWithVcVal{
			Elem: sender,
			Vc:   baseA.MostRecentSend[sender.GetRoutine()][sender.GetID()].Vc.Sync(vc.CurrentVC[sender.GetRoutine()]).Copy(),
			Val:  sender.GetID()}

		// for detection of receive on closed
		baseA.HasReceived[sender.GetID()] = true
		baseA.MostRecentReceive[recv.GetRoutine()][sender.GetID()] = baseA.ElemWithVcVal{Elem: recv,
			Vc:  baseA.MostRecentReceive[recv.GetRoutine()][sender.GetID()].Vc.Sync(vc.CurrentVC[recv.GetRoutine()]).Copy(),
			Val: sender.GetID(),
		}
	}

	if baseA.AnalysisCasesMap[flags.SendOnClosed] {
		if _, ok := baseA.CloseData[sender.GetID()]; ok {
			scenarios.FoundSendOnClosedChannel(sender, true)
		}
	}

	log.Info(fmt.Sprintf("[Channel] MixedDeadlock flag = %v", baseA.AnalysisCasesMap[flags.MixedDeadlock]))

	if baseA.AnalysisCasesMap[flags.MixedDeadlock] {
		scenarios.CheckForMixedDeadlock(sender.GetRoutine(), recv.GetRoutine())
	}

	if baseA.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(sender, vc.CurrentVC[sender.GetRoutine()], true, false)
		scenarios.CheckForSelectCaseWithPartnerChannel(recv, vc.CurrentVC[recv.GetRoutine()], false, false)
	}

	if baseA.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(sender.GetRoutine(), sender.GetID(), baseA.ElemWithVc{Vc: vc.CurrentVC[sender.GetRoutine()].Copy(), Elem: sender}, trace.ChannelSend, false)
		scenarios.CheckForLeakChannelRun(recv.GetRoutine(), sender.GetID(), baseA.ElemWithVc{Vc: vc.CurrentVC[recv.GetRoutine()].Copy(), Elem: recv}, trace.ChannelRecv, false)
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

	if baseA.MostRecentSend[routine] == nil {
		baseA.MostRecentSend[routine] = make(map[int]baseA.ElemWithVcVal)
	}

	// for detection of send on closed
	baseA.HasSend[id] = true
	baseA.MostRecentSend[routine][id] = baseA.ElemWithVcVal{
		Elem: ch,
		Vc:   baseA.MostRecentSend[routine][id].Vc.Sync(vc[routine]).Copy(),
		Val:  id,
	}

	if baseA.AnalysisCasesMap[flags.SendOnClosed] {
		if _, ok := baseA.CloseData[id]; ok {
			scenarios.FoundSendOnClosedChannel(ch, true)
		}
	}

	if baseA.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if baseA.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, trace.ChannelSend, true)
	}

	for i, hold := range baseA.HoldRecv {
		if hold.Ch.GetID() == id {
			Recv(hold.Ch, hold.Vc, hold.WVc)
			baseA.HoldRecv = append(baseA.HoldRecv[:i], baseA.HoldRecv[i+1:]...)
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

	if baseA.AnalysisCasesMap[flags.ConcurrentRecv] || baseA.AnalysisFuzzingFlow {
		scenarios.CheckForConcurrentRecv(ch, vc)
	}

	if ch.GetTPost() == 0 {
		return
	}

	if baseA.MostRecentReceive[routine] == nil {
		baseA.MostRecentReceive[routine] = make(map[int]baseA.ElemWithVcVal)
	}

	// for detection of receive on closed
	baseA.HasReceived[id] = true
	baseA.MostRecentReceive[routine][id] = baseA.ElemWithVcVal{
		Elem: ch,
		Vc:   baseA.MostRecentReceive[routine][id].Vc.Sync(vc[routine]),
		Val:  id,
	}

	if baseA.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if baseA.AnalysisCasesMap[flags.MixedDeadlock] {
		routSend := ch.GetPartner().GetRoutine()
		scenarios.CheckForMixedDeadlock(routSend, routine)
	}
	if baseA.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, trace.ChannelRecv, true)
	}

	for i, hold := range baseA.HoldSend {
		if hold.Ch.GetID() == id {
			Send(hold.Ch, hold.Vc, hold.WVc)
			baseA.HoldSend = append(baseA.HoldSend[:i], baseA.HoldSend[i+1:]...)
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

	if baseA.AnalysisCasesMap[flags.CloseOnClosed] {
		scenarios.CheckForClosedOnClosed(ch) // must be called before closePos is updated
	}

	baseA.CloseData[id] = ch

	if baseA.AnalysisCasesMap[flags.SendOnClosed] || baseA.AnalysisCasesMap[flags.ReceiveOnClosed] {
		scenarios.CheckForCommunicationOnClosedChannel(ch)
	}

	if baseA.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerClose(ch, vc.CurrentVC[routine])
	}

	if baseA.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
			Vc:   vc.CurrentVC[routine].Copy(),
			Elem: ch,
		}, trace.ChannelClose, true)
	}
}

// SendC record an actual send on closed
func SendC(ch *trace.ElementChannel) {
	if baseA.AnalysisCasesMap[flags.SendOnClosed] {
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

	if baseA.AnalysisCasesMap[flags.ReceiveOnClosed] {
		scenarios.FoundReceiveOnClosedChannel(ch, true)
	}

	if baseA.ModeIsFuzzing {
		scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], false, buffered)
	}
	if baseA.MostRecentReceive[routine] == nil {
		baseA.MostRecentReceive[routine] = make(map[int]baseA.ElemWithVcVal)
	}

	// for detection of receive on closed
	baseA.HasReceived[id] = true
	baseA.MostRecentReceive[routine][id] = baseA.ElemWithVcVal{
		Elem: ch,
		Vc:   baseA.MostRecentReceive[routine][id].Vc.Sync(vc[routine]),
		Val:  id,
	}

	if baseA.AnalysisCasesMap[flags.MixedDeadlock] {
		scenarios.CheckForMixedDeadlock(baseA.CloseData[id].GetRoutine(), routine)
	}

	if baseA.AnalysisCasesMap[flags.Leak] {
		scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, trace.ChannelRecv, buffered)
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

	if baseA.MostRecentSend[routine] == nil {
		baseA.MostRecentSend[routine] = make(map[int]baseA.ElemWithVcVal)
	}
	baseA.MostRecentSend[routine][id] = baseA.ElemWithVcVal{
		Elem: c,
		Vc:   c.GetVC(),
		Val:  id,
	}
	baseA.HasSend[routine] = true
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

	if baseA.MostRecentReceive[routine] == nil {
		baseA.MostRecentReceive[routine] = make(map[int]baseA.ElemWithVcVal)
	}
	baseA.MostRecentReceive[routine][id] = baseA.ElemWithVcVal{
		Elem: c,
		Vc:   c.GetVC(),
		Val:  id,
	}
	baseA.HasReceived[id] = true
}
