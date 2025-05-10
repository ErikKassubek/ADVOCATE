// Copyright (c) 2024 Erik Kassubek
//
// File: vcChannel.go
// Brief: Update functions for vector clocks from channel operations
//        Some of the update function also start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-27
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/clock"
	"advocate/timer"
	"advocate/trace"
	"advocate/utils"
	"strconv"
)

// elements for buffered channel internal vector clock
type bufferedVC struct {
	occupied    bool
	oID         int
	vc          *clock.VectorClock
	routineSend int
	tID         string
}

// UpdateVCChannel updates the vecto clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateVCChannel(ch *trace.TraceElementChannel) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := ch.GetID()
	routine := ch.GetRoutine()
	oID := ch.GetOID()
	opC := ch.GetOpC()
	cl := ch.GetClosed()

	ch.SetVc(currentVC[routine])
	ch.SetWVc(currentWVC[routine])

	if ch.GetTPost() == 0 {
		return
	}

	if ch.GetPartner() == nil {
		findPartner(ch)
	}

	// hold back receive operations, until the send operation is processed
	for _, elem := range waitingReceive {
		if elem.GetOID() <= maxOpID[id] {
			if len(waitingReceive) != 0 {
				waitingReceive = waitingReceive[1:]
			}
			UpdateVCChannel(elem)
		}
	}

	if ch.IsBuffered() {
		if opC == trace.SendOp {
			maxOpID[id] = oID
		} else if opC == trace.RecvOp {
			if oID > maxOpID[id] && !cl {
				waitingReceive = append(waitingReceive, ch)
				return
			}
		}

		switch opC {
		case trace.SendOp:
			Send(ch, currentVC, currentWVC, fifo)
		case trace.RecvOp:
			if cl { // recv on closed channel
				RecvC(ch, currentVC, currentWVC, true)
			} else {
				Recv(ch, currentVC, currentWVC, fifo)
			}
		case trace.CloseOp:
			Close(ch, currentVC, currentWVC)
		default:
			err := "Unknown operation: " + ch.ToString()
			utils.LogError(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.SendOp:
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				partner.SetVc(currentVC[partnerRout])
				sel := partner.GetSelect()
				if sel != nil {
					sel.SetVc(currentVC[partnerRout])
				}
				Unbuffered(ch, partner)
				// advance index of receive routine, send routine is already advanced
				MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					SendC(ch)
				} else {
					StuckChan(routine, currentVC, currentWVC)
				}
			}

		case trace.RecvOp: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				partner.SetVc(currentVC[partnerRout])
				Unbuffered(partner, ch)
				// advance index of receive routine, send routine is already advanced
				MainTraceIter.IncreaseIndex(partnerRout)
			} else {
				if cl { // recv on closed channel
					RecvC(ch, currentVC, currentWVC, false)
				} else {
					StuckChan(routine, currentVC, currentWVC)
				}
			}
		case trace.CloseOp:
			Close(ch, currentVC, currentWVC)
		default:
			err := "Unknown operation: " + ch.ToString()
			utils.LogError(err)
		}
	}
}

// UpdateVCSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateVCSelect(se *trace.TraceElementSelect) {
	noChannel := se.GetChosenDefault() || se.GetTPost() == 0

	routine := se.GetRoutine()

	se.SetVc(currentVC[routine])
	se.SetWVc(currentVC[routine])

	if noChannel {
		currentVC[routine].Inc(routine)
		currentWVC[routine].Inc(routine)
	} else {
		chosenCase := se.GetChosenCase()
		chosenCase.SetVc(se.GetVC())

		findPartner(chosenCase)
		UpdateVCChannel(chosenCase)
	}

	if modeIsFuzzing {
		CheckForSelectCaseWithPartnerSelect(se, currentVC[routine])
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

	if analysisCases["sendOnClosed"] {
		chosenIndex := se.GetChosenIndex()
		for i, c := range cases {
			if i == chosenIndex {
				continue
			}

			opC := c.GetOpC()

			if _, ok := closeData[c.GetID()]; ok {
				if opC == trace.SendOp {
					foundSendOnClosedChannel(&c, false)
				} else if opC == trace.RecvOp {
					foundReceiveOnClosedChannel(&c, false)
				}
			}
		}
	}

	if analysisCases["leak"] {
		for _, c := range cases {
			CheckForLeakChannelRun(routine, c.GetRoutine(),
				elemWithVc{
					vc:   se.GetVC().Copy(),
					elem: se},
				int(c.GetOpC()), c.IsBuffered())
		}
	}
}

// Find the partner of the channel operation
//
// Parameter:
//   - ch *trace.TraceElementChannel: the trace element
//
// Returns:
//   - int: The routine id of the partner, -1 if no partner was found
func findPartner(ch *trace.TraceElementChannel) *trace.TraceElementChannel {
	id := ch.GetID()
	oID := ch.GetOID()

	// return -1 if closed by channel
	if ch.GetClosed() || ch.GetTPost() == 0 {
		return nil
	}

	// find partner has already been applied to the partner and the communication
	// was fund. An repeated search is not necessary
	if ch.GetPartner() != nil {
		return ch.GetPartner()
	}

	// check if partner has already been processed
	if partner, ok := channelWithoutPartner[id][oID]; ok {
		if ch.IsEqual(partner) {
			return nil
		}

		// partner was already processed
		ch.SetPartner(partner)
		partner.SetPartner(ch)

		delete(channelWithoutPartner[id], oID)

		return partner

	}

	if channelWithoutPartner[id] == nil {
		channelWithoutPartner[id] = make(map[int]*trace.TraceElementChannel)
	}
	channelWithoutPartner[id][oID] = ch

	return nil
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
func Unbuffered(sender trace.TraceElement, recv trace.TraceElement) {
	if analysisCases["concurrentRecv"] || analysisFuzzing { // or fuzzing
		switch r := recv.(type) {
		case *trace.TraceElementChannel:
			checkForConcurrentRecv(r, currentVC)
		case *trace.TraceElementSelect:
			checkForConcurrentRecv(r.GetChosenCase(), currentVC)
		}
	}

	if analysisFuzzing {
		switch s := sender.(type) {
		case *trace.TraceElementChannel:
			getConcurrentSendForFuzzing(s)
		case *trace.TraceElementSelect:
			getConcurrentSendForFuzzing(s.GetChosenCase())
		}
	}

	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if sender.GetTPost() != 0 && recv.GetTPost() != 0 {

		if mostRecentReceive[recv.GetRoutine()] == nil {
			mostRecentReceive[recv.GetRoutine()] = make(map[int]ElemWithVcVal)
		}
		if mostRecentSend[sender.GetRoutine()] == nil {
			mostRecentSend[sender.GetRoutine()] = make(map[int]ElemWithVcVal)
		}

		// for detection of send on closed
		hasSend[sender.GetID()] = true
		mostRecentSend[sender.GetRoutine()][sender.GetID()] = ElemWithVcVal{sender, mostRecentSend[sender.GetRoutine()][sender.GetID()].Vc.Sync(currentVC[sender.GetRoutine()]).Copy(), sender.GetID()}

		// for detection of receive on closed
		hasReceived[sender.GetID()] = true
		mostRecentReceive[recv.GetRoutine()][sender.GetID()] = ElemWithVcVal{recv, mostRecentReceive[recv.GetRoutine()][sender.GetID()].Vc.Sync(currentVC[recv.GetRoutine()]).Copy(), sender.GetID()}

		currentVC[recv.GetRoutine()].Sync(currentVC[sender.GetRoutine()])
		currentVC[sender.GetRoutine()] = currentVC[recv.GetRoutine()].Copy()
		currentVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		currentVC[recv.GetRoutine()].Inc(recv.GetRoutine())
		currentWVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		currentWVC[recv.GetRoutine()].Inc(recv.GetRoutine())

	} else {
		currentVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		currentWVC[sender.GetRoutine()].Inc(sender.GetRoutine())
	}

	timer.Stop(timer.AnaHb)

	if analysisCases["sendOnClosed"] {
		if _, ok := closeData[sender.GetID()]; ok {
			foundSendOnClosedChannel(sender, true)
		}
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(sender.GetRoutine(), recv.GetRoutine())
	}

	if modeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(sender, currentVC[sender.GetRoutine()], true, false)
		CheckForSelectCaseWithPartnerChannel(recv, currentVC[recv.GetRoutine()], false, false)
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(sender.GetRoutine(), sender.GetID(), elemWithVc{currentVC[sender.GetRoutine()].Copy(), sender}, 0, false)
		CheckForLeakChannelRun(recv.GetRoutine(), sender.GetID(), elemWithVc{currentVC[recv.GetRoutine()].Copy(), recv}, 1, false)
	}
}

// holdObj can temporarily hold an channel operations with additional information
// it is used in the case that for a synchronous communication, the recv is
// recorded before the send
type holdObj struct {
	ch   *trace.TraceElementChannel
	vc   map[int]*clock.VectorClock
	wvc  map[int]*clock.VectorClock
	fifo bool
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
//   - fifo bool: true if the channel buffer is assumed to be fifo
func Send(ch *trace.TraceElementChannel, vc, wVc map[int]*clock.VectorClock, fifo bool) {
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

	if mostRecentSend[routine] == nil {
		mostRecentSend[routine] = make(map[int]ElemWithVcVal)
	}

	newBufferedVCs(id, qSize, vc[routine].GetSize())

	count := bufferedVCsCount[id]

	if bufferedVCsSize[id] <= count {
		holdSend = append(holdSend, holdObj{ch, vc, wVc, fifo})
		return
	}

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(bufferedVCs[id]) >= count && len(bufferedVCs[id]) < bufferedVCsSize[id] {
		bufferedVCs[id] = append(bufferedVCs[id], bufferedVC{false, 0, clock.NewVectorClock(vc[routine].GetSize()), 0, ""})
	}

	if count > qSize || bufferedVCs[id][count].occupied {
		utils.LogError("Write to occupied buffer position or to big count")
	}

	v := bufferedVCs[id][count].vc
	vc[routine].Sync(v)

	if fifo {
		vc[routine].Sync(mostRecentSend[routine][id].Vc)
	}

	// for detection of send on closed
	hasSend[id] = true
	mostRecentSend[routine][id] = ElemWithVcVal{ch, mostRecentSend[routine][id].Vc.Sync(vc[routine]).Copy(), id}

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	bufferedVCs[id][count] = bufferedVC{true, ch.GetOID(), vc[routine].Copy(), routine, ch.GetTID()}

	bufferedVCsCount[id]++

	timer.Stop(timer.AnaHb)

	if analysisCases["sendOnClosed"] {
		if _, ok := closeData[id]; ok {
			foundSendOnClosedChannel(ch, true)
		}
	}

	if modeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, elemWithVc{vc[routine].Copy(), ch}, 0, true)
	}

	for i, hold := range holdRecv {
		if hold.ch.GetID() == id {
			Recv(hold.ch, hold.vc, hold.wvc, hold.fifo)
			holdRecv = append(holdRecv[:i], holdRecv[i+1:]...)
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
func Recv(ch *trace.TraceElementChannel, vc, wVc map[int]*clock.VectorClock, fifo bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := ch.GetID()
	routine := ch.GetRoutine()
	oID := ch.GetOID()
	qSize := ch.GetQSize()

	if analysisCases["concurrentRecv"] || analysisFuzzing {
		checkForConcurrentRecv(ch, vc)
	}

	if ch.GetTPost() == 0 {
		vc[routine].Inc(routine)
		wVc[routine].Inc(routine)
		return
	}

	if mostRecentReceive[routine] == nil {
		mostRecentReceive[routine] = make(map[int]ElemWithVcVal)
	}

	newBufferedVCs(id, qSize, vc[routine].GetSize())

	if bufferedVCsCount[id] == 0 {
		holdRecv = append(holdRecv, holdObj{ch, vc, wVc, fifo})
		return
		// results.Debug("Read operation on empty buffer position", results.ERROR)
	}
	bufferedVCsCount[id]--

	if bufferedVCs[id][0].oID != oID {
		found := false
		for i := 1; i < len(bufferedVCs[id]); i++ {
			if bufferedVCs[id][i].oID == oID {
				found = true
				bufferedVCs[id][0] = bufferedVCs[id][i]
				bufferedVCs[id][i] = bufferedVC{false, 0, vc[routine].Copy(), 0, ""}
				break
			}
		}
		if !found {
			err := "Read operation on wrong buffer position - ID: " + strconv.Itoa(id) + ", OID: " + strconv.Itoa(oID) + ", SIZE: " + strconv.Itoa(qSize)
			utils.LogError(err)
		}
	}
	v := bufferedVCs[id][0].vc
	routSend := bufferedVCs[id][0].routineSend

	vc[routine] = vc[routine].Sync(v)

	if fifo {
		vc[routine] = vc[routine].Sync(mostRecentReceive[routine][id].Vc)
	}

	bufferedVCs[id] = append(bufferedVCs[id][1:], bufferedVC{false, 0, vc[routine].Copy(), 0, ""})

	// for detection of receive on closed
	hasReceived[id] = true
	mostRecentReceive[routine][id] = ElemWithVcVal{ch, mostRecentReceive[routine][id].Vc.Sync(vc[routine]), id}

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if modeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(ch, vc[routine], true, true)
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(routSend, routine)
	}
	if analysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, elemWithVc{vc[routine].Copy(), ch}, 1, true)
	}

	for i, hold := range holdSend {
		if hold.ch.GetID() == id {
			Send(hold.ch, hold.vc, hold.wvc, hold.fifo)
			holdSend = append(holdSend[:i], holdSend[i+1:]...)
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
func Close(ch *trace.TraceElementChannel, vc, wVc map[int]*clock.VectorClock) {
	if ch.GetTPost() == 0 {
		return
	}

	routine := ch.GetRoutine()
	id := ch.GetID()

	ch.SetClosed(true)

	if analysisCases["closeOnClosed"] {
		checkForClosedOnClosed(ch) // must be called before closePos is updated
	}

	timer.Start(timer.AnaHb)

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	closeData[id] = ch

	timer.Stop(timer.AnaHb)

	if analysisCases["sendOnClosed"] || analysisCases["receiveOnClosed"] {
		checkForCommunicationOnClosedChannel(ch)
	}

	if modeIsFuzzing {
		CheckForSelectCaseWithPartnerClose(ch, vc[routine])
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, elemWithVc{vc[routine].Copy(), ch}, 2, true)
	}
}

// SendC record an actual send on closed
func SendC(ch *trace.TraceElementChannel) {
	if analysisCases["sendOnClosed"] {
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
func RecvC(ch *trace.TraceElementChannel, vc, wVc map[int]*clock.VectorClock, buffered bool) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetID()
	routine := ch.GetRoutine()

	if analysisCases["receiveOnClosed"] {
		foundReceiveOnClosedChannel(ch, true)
	}

	timer.Start(timer.AnaHb)
	if _, ok := closeData[id]; ok {
		vc[routine] = vc[routine].Sync(closeData[id].GetVC())
	}

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if modeIsFuzzing {
		CheckForSelectCaseWithPartnerChannel(ch, vc[routine], false, buffered)
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(closeData[id].GetRoutine(), routine)
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(routine, id, elemWithVc{vc[routine].Copy(), ch}, 1, buffered)
	}
}

// Create a new map of buffered vector clocks for a channel if not already in
// bufferedVCs.
//
// Parameter:
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
//   - numRout int: the number of routines
func newBufferedVCs(id int, qSize int, numRout int) {
	if _, ok := bufferedVCs[id]; !ok {
		bufferedVCs[id] = make([]bufferedVC, 1)
		bufferedVCsCount[id] = 0
		bufferedVCsSize[id] = qSize
		bufferedVCs[id][0] = bufferedVC{false, 0, clock.NewVectorClock(numRout), 0, ""}
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
func SetChannelAsLastSend(c trace.TraceElement) {
	id := c.GetID()
	routine := c.GetRoutine()

	if mostRecentSend[routine] == nil {
		mostRecentSend[routine] = make(map[int]ElemWithVcVal)
	}
	mostRecentSend[routine][id] = ElemWithVcVal{c, c.GetVC(), id}
	hasSend[routine] = true
}

// SetChannelAsLastReceive sets the channel as the last recv operation.
// Used for not executed select recv
//
// Parameter:
//   - id int: the id of the channel
//   - rout int: the route of the operation
//   - vc VectorClock: the vector clock of the operation
//   - tID string: the position of the recv in the program
func SetChannelAsLastReceive(c trace.TraceElement) {
	id := c.GetID()
	routine := c.GetRoutine()

	if mostRecentReceive[routine] == nil {
		mostRecentReceive[routine] = make(map[int]ElemWithVcVal)
	}
	mostRecentReceive[routine][id] = ElemWithVcVal{c, c.GetVC(), id}
	hasReceived[id] = true
}
