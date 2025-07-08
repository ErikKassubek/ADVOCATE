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

package analysis

import (
	"advocate/analysis/concurrent/clock"
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/timer"
	"strconv"
)

// UpdateHBChannel updates the vecto clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateHBChannel(ch *trace.ElementChannel) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := ch.GetID()
	routine := ch.GetRoutine()
	oID := ch.GetOID()
	opC := ch.GetOpC()
	cl := ch.GetClosed()

	ch.SetVc(data.CurrentVC[routine])
	ch.SetWVc(data.CurrentWVC[routine])

	if ch.GetTPost() == 0 {
		return
	}

	// hold back receive operations, until the send operation is processed
	for _, elem := range data.WaitingReceive {
		if elem.GetOID() <= data.MaxOpID[id] {
			if len(data.WaitingReceive) != 0 {
				data.WaitingReceive = data.WaitingReceive[1:]
			}
			UpdateHBChannel(elem)
		}
	}

	if ch.IsBuffered() {
		if opC == trace.SendOp {
			data.MaxOpID[id] = oID
		} else if opC == trace.RecvOp {
			if oID > data.MaxOpID[id] && !cl {
				data.WaitingReceive = append(data.WaitingReceive, ch)
				return
			}
		}

		switch opC {
		case trace.SendOp:
			Send(ch, data.CurrentVC, data.CurrentWVC, data.Fifo)
		case trace.RecvOp:
			if cl { // recv on closed channel
				RecvC(ch, data.CurrentVC, data.CurrentWVC, true)
			} else {
				Recv(ch, data.CurrentVC, data.CurrentWVC, data.Fifo)
			}
		case trace.CloseOp:
			Close(ch, data.CurrentVC, data.CurrentWVC)
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
				partner.SetVc(data.CurrentVC[partnerRout])
				sel := partner.GetSelect()
				if sel != nil {
					sel.SetVc(data.CurrentVC[partnerRout])
				}
				Unbuffered(ch, partner)
				// advance index of receive routine, send routine is already advanced
				data.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					SendC(ch)
				} else {
					StuckChan(routine, data.CurrentVC, data.CurrentWVC)
				}
			}

		case trace.RecvOp: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				partner.SetVc(data.CurrentVC[partnerRout])
				Unbuffered(partner, ch)
				// advance index of receive routine, send routine is already advanced
				data.MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					RecvC(ch, data.CurrentVC, data.CurrentWVC, false)
				} else {
					StuckChan(routine, data.CurrentVC, data.CurrentWVC)
				}
			}
		case trace.CloseOp:
			Close(ch, data.CurrentVC, data.CurrentWVC)
		default:
			err := "Unknown operation: " + ch.ToString()
			log.Error(err)
		}
	}
}

// UpdateHBSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateHBSelect(se *trace.ElementSelect) {
	noChannel := se.GetChosenDefault() || se.GetTPost() == 0

	routine := se.GetRoutine()

	se.SetVc(data.CurrentVC[routine])
	se.SetWVc(data.CurrentVC[routine])

	if noChannel {
		data.CurrentVC[routine].Inc(routine)
		data.CurrentWVC[routine].Inc(routine)
	} else {
		chosenCase := se.GetChosenCase()
		chosenCase.SetVc(se.GetVC())

		UpdateHBChannel(chosenCase)
	}

	if data.ModeIsFuzzing {
		CheckForSelectCaseWithPartnerSelect(se, data.CurrentVC[routine])
	}

	cases := se.GetCases()

	for _, c := range cases {
		c.SetVc(se.GetVC())
		opC := c.GetOpC()
		if opC == trace.SendOp {
			SetChannelAsLastSend(&c)
		} else if opC == trace.RecvOp {
			SetChannelAsLastReceive(&c)
		}
	}

	if data.AnalysisCases["sendOnClosed"] {
		chosenIndex := se.GetChosenIndex()
		for i, c := range cases {
			if i == chosenIndex {
				continue
			}

			opC := c.GetOpC()

			if _, ok := data.CloseData[c.GetID()]; ok {
				if opC == trace.SendOp {
					foundSendOnClosedChannel(&c, false)
				} else if opC == trace.RecvOp {
					foundReceiveOnClosedChannel(&c, false)
				}
			}
		}
	}

	if data.AnalysisCases["leak"] {
		for _, c := range cases {
			CheckForLeakChannelRun(routine, c.GetRoutine(),
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
	if data.AnalysisCases["concurrentRecv"] || data.AnalysisFuzzing { // or fuzzing
		switch r := recv.(type) {
		case *trace.ElementChannel:
			checkForConcurrentRecv(r, data.CurrentVC)
		case *trace.ElementSelect:
			checkForConcurrentRecv(r.GetChosenCase(), data.CurrentVC)
		}
	}

	if data.AnalysisFuzzing {
		switch s := sender.(type) {
		case *trace.ElementChannel:
			getConcurrentSendForFuzzing(s)
		case *trace.ElementSelect:
			getConcurrentSendForFuzzing(s.GetChosenCase())
		}
	}

	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

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
			Vc:   data.MostRecentSend[sender.GetRoutine()][sender.GetID()].Vc.Sync(data.CurrentVC[sender.GetRoutine()]).Copy(),
			Val:  sender.GetID()}

		// for detection of receive on closed
		data.HasReceived[sender.GetID()] = true
		data.MostRecentReceive[recv.GetRoutine()][sender.GetID()] = data.ElemWithVcVal{Elem: recv,
			Vc:  data.MostRecentReceive[recv.GetRoutine()][sender.GetID()].Vc.Sync(data.CurrentVC[recv.GetRoutine()]).Copy(),
			Val: sender.GetID(),
		}

		data.CurrentVC[recv.GetRoutine()].Sync(data.CurrentVC[sender.GetRoutine()])
		data.CurrentVC[sender.GetRoutine()] = data.CurrentVC[recv.GetRoutine()].Copy()
		data.CurrentVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		data.CurrentVC[recv.GetRoutine()].Inc(recv.GetRoutine())
		data.CurrentWVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		data.CurrentWVC[recv.GetRoutine()].Inc(recv.GetRoutine())

	} else {
		data.CurrentVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		data.CurrentWVC[sender.GetRoutine()].Inc(sender.GetRoutine())
	}

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["sendOnClosed"] {
		if _, ok := data.CloseData[sender.GetID()]; ok {
			foundSendOnClosedChannel(sender, true)
		}
	}

	if data.AnalysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(sender.GetRoutine(), recv.GetRoutine())
	}

	if data.ModeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(sender, data.CurrentVC[sender.GetRoutine()], true, false)
		CheckForSelectCaseWithPartnerChannel(recv, data.CurrentVC[recv.GetRoutine()], false, false)
	}

	if data.AnalysisCases["leak"] {
		CheckForLeakChannelRun(sender.GetRoutine(), sender.GetID(), data.ElemWithVc{Vc: data.CurrentVC[sender.GetRoutine()].Copy(), Elem: sender}, 0, false)
		CheckForLeakChannelRun(recv.GetRoutine(), sender.GetID(), data.ElemWithVc{Vc: data.CurrentVC[recv.GetRoutine()].Copy(), Elem: recv}, 1, false)
	}
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
//   - fifo bool: true if the channel buffer is assumed to be fifo
func Send(ch *trace.ElementChannel, vc, wVc map[int]*clock.VectorClock, fifo bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := ch.GetID()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()

	if ch.GetTPost() == 0 {
		vc[routine].Inc(routine)
		wVc[routine].Inc(routine)
		return
	}

	if data.MostRecentSend[routine] == nil {
		data.MostRecentSend[routine] = make(map[int]data.ElemWithVcVal)
	}

	newBufferedVCs(id, qSize, vc[routine].GetSize())

	count := data.BufferedVCsCount[id]

	if data.BufferedVCsSize[id] <= count {
		data.HoldSend = append(data.HoldSend, data.HoldObj{
			Ch:   ch,
			Vc:   vc,
			WVc:  wVc,
			Fifo: fifo,
		})
		return
	}

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(data.BufferedVCs[id]) >= count && len(data.BufferedVCs[id]) < data.BufferedVCsSize[id] {
		data.BufferedVCs[id] = append(data.BufferedVCs[id], data.BufferedVC{
			Occupied:    false,
			OID:         0,
			Vc:          clock.NewVectorClock(vc[routine].GetSize()),
			RoutineSend: 0,
			TID:         ""})
	}

	if count > qSize || data.BufferedVCs[id][count].Occupied {
		log.Error("Write to occupied buffer position or to big count")
	}

	v := data.BufferedVCs[id][count].Vc
	vc[routine].Sync(v)

	if fifo {
		vc[routine].Sync(data.MostRecentSend[routine][id].Vc)
	}

	// for detection of send on closed
	data.HasSend[id] = true
	data.MostRecentSend[routine][id] = data.ElemWithVcVal{
		Elem: ch,
		Vc:   data.MostRecentSend[routine][id].Vc.Sync(vc[routine]).Copy(),
		Val:  id,
	}

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	data.BufferedVCs[id][count] = data.BufferedVC{
		Occupied:    true,
		OID:         ch.GetOID(),
		Vc:          vc[routine].Copy(),
		RoutineSend: routine,
		TID:         ch.GetTID(),
	}

	data.BufferedVCsCount[id]++

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["sendOnClosed"] {
		if _, ok := data.CloseData[id]; ok {
			foundSendOnClosedChannel(ch, true)
		}
	}

	if data.ModeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if data.AnalysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, 0, true)
	}

	for i, hold := range data.HoldRecv {
		if hold.Ch.GetID() == id {
			Recv(hold.Ch, hold.Vc, hold.WVc, hold.Fifo)
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
//   - fifo bool: true if the channel buffer is assumed to be fifo
func Recv(ch *trace.ElementChannel, vc, wVc map[int]*clock.VectorClock, fifo bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := ch.GetID()
	routine := ch.GetRoutine()
	oID := ch.GetOID()
	qSize := ch.GetQSize()

	if data.AnalysisCases["concurrentRecv"] || data.AnalysisFuzzing {
		checkForConcurrentRecv(ch, vc)
	}

	if ch.GetTPost() == 0 {
		vc[routine].Inc(routine)
		wVc[routine].Inc(routine)
		return
	}

	if data.MostRecentReceive[routine] == nil {
		data.MostRecentReceive[routine] = make(map[int]data.ElemWithVcVal)
	}

	newBufferedVCs(id, qSize, vc[routine].GetSize())

	if data.BufferedVCsCount[id] == 0 {
		data.HoldRecv = append(data.HoldRecv, data.HoldObj{
			Ch:   ch,
			Vc:   vc,
			WVc:  wVc,
			Fifo: fifo,
		})
		return
		// results.Debug("Read operation on empty buffer position", results.ERROR)
	}
	data.BufferedVCsCount[id]--

	if data.BufferedVCs[id][0].OID != oID {
		found := false
		for i := 1; i < len(data.BufferedVCs[id]); i++ {
			if data.BufferedVCs[id][i].OID == oID {
				found = true
				data.BufferedVCs[id][0] = data.BufferedVCs[id][i]
				data.BufferedVCs[id][i] = data.BufferedVC{
					Occupied: false,
					OID:      0, Vc: vc[routine].Copy(),
					RoutineSend: 0,
					TID:         "",
				}
				break
			}
		}
		if !found {
			err := "Read operation on wrong buffer position - ID: " + strconv.Itoa(id) + ", OID: " + strconv.Itoa(oID) + ", SIZE: " + strconv.Itoa(qSize)
			log.Error(err)
		}
	}
	v := data.BufferedVCs[id][0].Vc
	routSend := data.BufferedVCs[id][0].RoutineSend

	vc[routine] = vc[routine].Sync(v)

	if fifo {
		vc[routine] = vc[routine].Sync(data.MostRecentReceive[routine][id].Vc)
	}

	data.BufferedVCs[id] = append(data.BufferedVCs[id][1:], data.BufferedVC{
		Occupied:    false,
		OID:         0,
		Vc:          vc[routine].Copy(),
		RoutineSend: 0,
		TID:         "",
	})

	// for detection of receive on closed
	data.HasReceived[id] = true
	data.MostRecentReceive[routine][id] = data.ElemWithVcVal{
		Elem: ch,
		Vc:   data.MostRecentReceive[routine][id].Vc.Sync(vc[routine]),
		Val:  id,
	}

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if data.ModeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if data.AnalysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(routSend, routine)
	}
	if data.AnalysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, 1, true)
	}

	for i, hold := range data.HoldSend {
		if hold.Ch.GetID() == id {
			Send(hold.Ch, hold.Vc, hold.WVc, hold.Fifo)
			data.HoldSend = append(data.HoldSend[:i], data.HoldSend[i+1:]...)
			break
		}
	}
}

// StuckChan updates and calculates the vector clocks for a stuck channel element
//
// Parameter:
//   - routine int: the route of the operation
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
func StuckChan(routine int, vc, wVc map[int]*clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)
}

// Close updates and calculates the vector clocks given a close on a channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
func Close(ch *trace.ElementChannel, vc, wVc map[int]*clock.VectorClock) {
	if ch.GetTPost() == 0 {
		return
	}

	routine := ch.GetRoutine()
	id := ch.GetID()

	ch.SetClosed(true)

	if data.AnalysisCases["closeOnClosed"] {
		checkForClosedOnClosed(ch) // must be called before closePos is updated
	}

	timer.Start(timer.AnaHb)

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	data.CloseData[id] = ch

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["sendOnClosed"] || data.AnalysisCases["receiveOnClosed"] {
		checkForCommunicationOnClosedChannel(ch)
	}

	if data.ModeIsFuzzing {
		CheckForSelectCaseWithPartnerClose(ch, vc[routine])
	}

	if data.AnalysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, 2, true)
	}
}

// SendC record an actual send on closed
func SendC(ch *trace.ElementChannel) {
	if data.AnalysisCases["sendOnClosed"] {
		foundSendOnClosedChannel(ch, true)
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

	if data.AnalysisCases["receiveOnClosed"] {
		foundReceiveOnClosedChannel(ch, true)
	}

	timer.Start(timer.AnaHb)
	if _, ok := data.CloseData[id]; ok {
		vc[routine] = vc[routine].Sync(data.CloseData[id].GetVC())
	}

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if data.ModeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(ch, vc[routine], false, buffered)
	}

	if data.AnalysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(data.CloseData[id].GetRoutine(), routine)
	}

	if data.AnalysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, data.ElemWithVc{
			Vc:   vc[routine].Copy(),
			Elem: ch,
		}, 1, buffered)
	}
}

// Create a new map of buffered vector clocks for a channel if not already in
// data.BufferedVCs.
//
// Parameter:
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
//   - numRout int: the number of routines
func newBufferedVCs(id int, qSize int, numRout int) {
	if _, ok := data.BufferedVCs[id]; !ok {
		data.BufferedVCs[id] = make([]data.BufferedVC, 1)
		data.BufferedVCsCount[id] = 0
		data.BufferedVCsSize[id] = qSize
		data.BufferedVCs[id][0] = data.BufferedVC{
			Occupied:    false,
			OID:         0,
			Vc:          clock.NewVectorClock(numRout),
			RoutineSend: 0,
			TID:         "",
		}
	}
}

// SetChannelAsLastSend sets the channel as the last send operation.
// Used for not executed select send
//
// Parameter:
//   - id int: the id of the channel
//   - routine int: the route of the operation
//   - vc VectorClock: the vector clock of the operation
//   - tID string: the position of the send in the program
func SetChannelAsLastSend(c trace.Element) {
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

// SetChannelAsLastReceive sets the channel as the last recv operation.
// Used for not executed select recv
//
// Parameter:
//   - id int: the id of the channel
//   - rout int: the route of the operation
//   - vc VectorClock: the vector clock of the operation
//   - tID string: the position of the recv in the program
func SetChannelAsLastReceive(c trace.Element) {
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
