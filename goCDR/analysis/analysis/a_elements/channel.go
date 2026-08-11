// gocdr/analysis/analysis/elements/channel.go

// Copyright (c) 2024 Erik Kassubek
//
// File: hbChannel.go
// Brief: Update functions for happens before info for channel operations
//        Some of the update function also start analysis functions
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"gocdr/analysis/a_base"
	"gocdr/analysis/analysis/a_scenarios"
	"gocdr/analysis/hb/a_clock"
	"gocdr/analysis/hb/a_hbcalc"
	"gocdr/analysis/hb/a_vc"
	"gocdr/trace"
	"gocdr/utils/flags"
	"gocdr/utils/log"
	"gocdr/utils/results/results"
)

// UpdateChannel updates the vector clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateChannel(ch *trace.ElementChannel) {
	id := ch.ObjID()
	opC := ch.Type(true)
	oID := ch.GetOID()
	cl := ch.GetClosed()

	// run hold back recv if the send has been processed
	for _, elem := range a_base.WaitingReceive {
		if elem.GetOID() <= a_base.MaxOpID[id] {
			if len(a_base.WaitingReceive) != 0 {
				a_base.WaitingReceive = a_base.WaitingReceive[1:]
			}
			UpdateChannel(elem)
		}
	}

	// hold back receive operations, until the send operation is processed
	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			a_base.MaxOpID[id] = oID
		case trace.ChannelRecv:
			if oID > a_base.MaxOpID[id] && !cl {
				a_base.WaitingReceive = append(a_base.WaitingReceive, ch)
				return
			}
		}
	}

	a_hbcalc.UpdateHBChannel(ch)

	if !ch.Committed() {
		return
	}

	results.AddContext(ch.File(), ch.Line(), ch.ObjID())

	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			Send(ch, a_vc.CurrentVC, a_vc.CurrentWVC)
		case trace.ChannelRecv:
			if cl { // recv on closed channel
				RecvC(ch, a_vc.CurrentVC, a_vc.CurrentWVC, true)
			} else {
				Recv(ch, a_vc.CurrentVC, a_vc.CurrentWVC)
			}
		case trace.ChannelClose:
			Close(ch)
		default:
			err := "Unknown operation: " + ch.String()
			log.Error(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.ChannelSend:
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.Routine()
				Unbuffered(ch, partner)
				// advance index of receive routine, send routine is already advanced
				a_base.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					SendC(ch)
				}
			}

		case trace.ChannelRecv: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.Routine()
				Unbuffered(partner, ch)
				// advance index of receive routine, send routine is already advanced
				a_base.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					RecvC(ch, a_vc.CurrentVC, a_vc.CurrentWVC, false)
				}
			}
		case trace.ChannelClose:
			Close(ch)
		default:
			err := "Unknown operation: " + ch.String()
			log.Error(err)
		}
	}
}

// UpdateSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateSelect(se *trace.ElementSelect) {
	routine := se.Routine()

	if a_base.ModeIsFuzzing {
		a_scenarios.CheckForSelectCaseWithPartnerSelect(se, a_vc.CurrentVC[routine])
	}

	a_hbcalc.UpdateHBSelect(se)

	cases := se.GetCases()

	for _, c := range cases {
		opC := c.Type(true)
		switch opC {
		case trace.ChannelSend:
			setChannelAsLastSend(c)
		case trace.ChannelRecv:
			setChannelAsLastReceive(c)
		}
	}

	if a_base.AnalysisCasesMap[flags.SendOnClosed] {
		chosenIndex := se.GetChosenIndex()
		for i, c := range cases {
			if i == chosenIndex {
				continue
			}

			opC := c.Type(true)

			if _, ok := a_base.CloseData[c.ObjID()]; ok {
				switch opC {
				case trace.ChannelSend:
					a_scenarios.FoundSendOnClosedChannel(c, false)
					// case trace.ChannelRecv:
					// scenarios.FoundReceiveOnClosedChannel(&c, false)
				}
			}
		}
	}

	// if baseA.AnalysisCasesMap[flags.Leak] {
	// 	for _, c := range cases {
	// 		scenarios.CheckForLeakChannelRun(routine, c.GetRoutine(),
	// 			baseA.ElemWithVc{
	// 				Vc:   se.GetVC(a_clock.Strong).Copy(),
	// 				Elem: se},
	// 			c.GetType(true), c.IsBuffered())
	// 	}
	// }
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
	if a_base.AnalysisCasesMap[flags.ConcurrentRecv] || a_base.AnalysisFuzzingFlow { // or fuzzing
		switch r := recv.(type) {
		case *trace.ElementChannel:
			a_scenarios.CheckForConcurrentRecv(r, a_vc.CurrentVC)
		case *trace.ElementSelect:
			a_scenarios.CheckForConcurrentRecv(r.GetChosenCase(), a_vc.CurrentVC)
		}
	}

	if a_base.AnalysisFuzzingFlow {
		switch s := sender.(type) {
		case *trace.ElementChannel:
			a_scenarios.GetConcurrentSendForFuzzing(s)
		case *trace.ElementSelect:
			a_scenarios.GetConcurrentSendForFuzzing(s.GetChosenCase())
		}
	}

	if sender.Committed() && recv.Committed() {
		if a_base.MostRecentReceive[recv.Routine()] == nil {
			a_base.MostRecentReceive[recv.Routine()] = make(map[int]a_base.ElemWithVcVal)
		}
		if a_base.MostRecentSend[sender.Routine()] == nil {
			a_base.MostRecentSend[sender.Routine()] = make(map[int]a_base.ElemWithVcVal)
		}

		// for detection of send on closed
		a_base.HasSend[sender.ObjID()] = true
		a_base.MostRecentSend[sender.Routine()][sender.ObjID()] = a_base.ElemWithVcVal{
			Elem: sender,
			Vc:   a_base.MostRecentSend[sender.Routine()][sender.ObjID()].Vc.Sync(a_vc.CurrentVC[sender.Routine()]).Copy(),
			Val:  sender.ObjID()}

		// for detection of receive on closed
		a_base.HasReceived[sender.ObjID()] = true
		a_base.MostRecentReceive[recv.Routine()][sender.ObjID()] = a_base.ElemWithVcVal{Elem: recv,
			Vc:  a_base.MostRecentReceive[recv.Routine()][sender.ObjID()].Vc.Sync(a_vc.CurrentVC[recv.Routine()]).Copy(),
			Val: sender.ObjID(),
		}
	}

	if a_base.AnalysisCasesMap[flags.SendOnClosed] {
		if _, ok := a_base.CloseData[sender.ObjID()]; ok {
			a_scenarios.FoundSendOnClosedChannel(sender, true)
		}
	}

	if a_base.ModeIsFuzzing {
		a_scenarios.CheckForSelectCaseWithPartnerChannel(sender, a_vc.CurrentVC[sender.Routine()], true, false)
		a_scenarios.CheckForSelectCaseWithPartnerChannel(recv, a_vc.CurrentVC[recv.Routine()], false, false)
	}

	// if baseA.AnalysisCasesMap[flags.Leak] {
	// 	scenarios.CheckForLeakChannelRun(sender.GetRoutine(), sender.GetObjId(), baseA.ElemWithVc{Vc: vc.CurrentVC[sender.GetRoutine()].Copy(), Elem: sender}, trace.ChannelSend, false)
	// 	scenarios.CheckForLeakChannelRun(recv.GetRoutine(), sender.GetObjId(), baseA.ElemWithVc{Vc: vc.CurrentVC[recv.GetRoutine()].Copy(), Elem: recv}, trace.ChannelRecv, false)
	// }
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
func Send(ch *trace.ElementChannel, vc, wVc map[int]*a_clock.VectorClock) {
	id := ch.ObjID()
	routine := ch.Routine()

	if !ch.Committed() {
		return
	}

	if a_base.MostRecentSend[routine] == nil {
		a_base.MostRecentSend[routine] = make(map[int]a_base.ElemWithVcVal)
	}

	// for detection of send on closed
	a_base.HasSend[id] = true
	a_base.MostRecentSend[routine][id] = a_base.ElemWithVcVal{
		Elem: ch,
		Vc:   a_base.MostRecentSend[routine][id].Vc.Sync(vc[routine]).Copy(),
		Val:  id,
	}

	if a_base.AnalysisCasesMap[flags.SendOnClosed] {
		if _, ok := a_base.CloseData[id]; ok {
			a_scenarios.FoundSendOnClosedChannel(ch, true)
		}
	}

	if a_base.ModeIsFuzzing {
		a_scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	// if baseA.AnalysisCasesMap[flags.Leak] {
	// 	scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
	// 		Vc:   vc[routine].Copy(),
	// 		Elem: ch,
	// 	}, trace.ChannelSend, true)
	// }

	for i, hold := range a_base.HoldRecv {
		if hold.Ch.ObjID() == id {
			Recv(hold.Ch, hold.Vc, hold.WVc)
			a_base.HoldRecv = append(a_base.HoldRecv[:i], a_base.HoldRecv[i+1:]...)
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
func Recv(ch *trace.ElementChannel, vc, wVc map[int]*a_clock.VectorClock) {
	id := ch.ObjID()
	routine := ch.Routine()

	if a_base.AnalysisCasesMap[flags.ConcurrentRecv] || a_base.AnalysisFuzzingFlow {
		a_scenarios.CheckForConcurrentRecv(ch, vc)
	}

	if !ch.Committed() {
		return
	}

	if a_base.MostRecentReceive[routine] == nil {
		a_base.MostRecentReceive[routine] = make(map[int]a_base.ElemWithVcVal)
	}

	// for detection of receive on closed
	a_base.HasReceived[id] = true
	a_base.MostRecentReceive[routine][id] = a_base.ElemWithVcVal{
		Elem: ch,
		Vc:   a_base.MostRecentReceive[routine][id].Vc.Sync(vc[routine]),
		Val:  id,
	}

	if a_base.ModeIsFuzzing {
		a_scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	// if baseA.AnalysisCasesMap[flags.Leak] {
	// 	scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
	// 		Vc:   vc[routine].Copy(),
	// 		Elem: ch,
	// 	}, trace.ChannelRecv, true)
	// }

	for i, hold := range a_base.HoldSend {
		if hold.Ch.ObjID() == id {
			Send(hold.Ch, hold.Vc, hold.WVc)
			a_base.HoldSend = append(a_base.HoldSend[:i], a_base.HoldSend[i+1:]...)
			break
		}
	}
}

// Close updates and calculates the vector clocks given a close on a channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func Close(ch *trace.ElementChannel) {
	if !ch.Committed() {
		return
	}

	routine := ch.Routine()
	id := ch.ObjID()

	ch.SetClosed(true)

	if a_base.AnalysisCasesMap[flags.CloseOnClosed] {
		a_scenarios.CheckForClosedOnClosed(ch) // must be called before closePos is updated
	}

	a_base.CloseData[id] = ch

	if a_base.AnalysisCasesMap[flags.SendOnClosed] { // || baseA.AnalysisCasesMap[flags.ReceiveOnClosed] {
		a_scenarios.CheckForCommunicationOnClosedChannel(ch)
	}

	if a_base.ModeIsFuzzing {
		a_scenarios.CheckForSelectCaseWithPartnerClose(ch, a_vc.CurrentVC[routine])
	}

	// if baseA.AnalysisCasesMap[flags.Leak] {
	// 	scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
	// 		Vc:   vc.CurrentVC[routine].Copy(),
	// 		Elem: ch,
	// 	}, trace.ChannelClose, true)
	// }
}

// SendC record an actual send on closed
func SendC(ch *trace.ElementChannel) {
	if a_base.AnalysisCasesMap[flags.SendOnClosed] {
		a_scenarios.FoundSendOnClosedChannel(ch, true)
	}
}

// RecvC updates and calculates the vector clocks given a receive on a closed channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
//   - buffered bool: true if the channel is buffered
func RecvC(ch *trace.ElementChannel, vc, wVc map[int]*a_clock.VectorClock, buffered bool) {
	if !ch.Committed() {
		return
	}

	id := ch.ObjID()
	routine := ch.Routine()

	if a_base.MostRecentReceive[routine] == nil {
		a_base.MostRecentReceive[routine] = make(map[int]a_base.ElemWithVcVal)
	}

	// for detection of receive on closed
	a_base.HasReceived[id] = true
	a_base.MostRecentReceive[routine][id] = a_base.ElemWithVcVal{
		Elem: ch,
		Vc:   a_base.MostRecentReceive[routine][id].Vc.Sync(vc[routine]),
		Val:  id,
	}

	// if baseA.AnalysisCasesMap[flags.ReceiveOnClosed] {
	// 	scenarios.FoundReceiveOnClosedChannel(ch, true)
	// }

	if a_base.ModeIsFuzzing {
		a_scenarios.CheckForSelectCaseWithPartnerChannel(ch, vc[routine], false, buffered)
	}

	// if baseA.AnalysisCasesMap[flags.Leak] {
	// 	scenarios.CheckForLeakChannelRun(routine, id, baseA.ElemWithVc{
	// 		Vc:   vc[routine].Copy(),
	// 		Elem: ch,
	// 	}, trace.ChannelRecv, buffered)
	// }
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
	id := c.ObjID()
	routine := c.Routine()

	if a_base.MostRecentSend[routine] == nil {
		a_base.MostRecentSend[routine] = make(map[int]a_base.ElemWithVcVal)
	}
	a_base.MostRecentSend[routine][id] = a_base.ElemWithVcVal{
		Elem: c,
		Vc:   c.GetVC(a_clock.Strong),
		Val:  id,
	}
	a_base.HasSend[routine] = true
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
	id := c.ObjID()
	routine := c.Routine()

	if a_base.MostRecentReceive[routine] == nil {
		a_base.MostRecentReceive[routine] = make(map[int]a_base.ElemWithVcVal)
	}
	a_base.MostRecentReceive[routine][id] = a_base.ElemWithVcVal{
		Elem: c,
		Vc:   c.GetVC(a_clock.Strong),
		Val:  id,
	}
	a_base.HasReceived[id] = true
}
