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
	"analyzer/clock"
	"analyzer/timer"
	"analyzer/utils"
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

// Update and calculate the vector clocks given a send/receive pair on a unbuffered
// channel.
//
// Parameter:
//   - ch (*TraceElementChannel): The trace element
//   - routSend (int): the route of the sender
//   - routRecv (int): the route of the receiver
//   - tID_send (string): the position of the send in the program
//   - tID_recv (string): the position of the receive in the program
func Unbuffered(sender TraceElement, recv TraceElement) {
	if analysisCases["concurrentRecv"] || analysisFuzzing { // or fuzzing
		switch r := recv.(type) {
		case *TraceElementChannel:
			checkForConcurrentRecv(r, currentVC)
		case *TraceElementSelect:
			checkForConcurrentRecv(&r.chosenCase, currentVC)
		}
	}

	if analysisFuzzing {
		switch s := sender.(type) {
		case *TraceElementChannel:
			getConcurrentSendForFuzzing(s)
		case *TraceElementSelect:
			getConcurrentSendForFuzzing(&s.chosenCase)
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
		checkForMixedDeadlock(sender.GetRoutine(), recv.GetRoutine(), sender.GetTID(), recv.GetTID())
	}

	if analysisCases["selectWithoutPartner"] || modeIsFuzzing {
		CheckForSelectCaseWithoutPartnerChannel(sender, currentVC[sender.GetRoutine()], true, false)
		CheckForSelectCaseWithoutPartnerChannel(recv, currentVC[recv.GetRoutine()], false, false)
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(sender.GetRoutine(), sender.GetID(), elemWithVc{currentVC[sender.GetRoutine()].Copy(), sender}, 0, false)
		CheckForLeakChannelRun(recv.GetRoutine(), sender.GetID(), elemWithVc{currentVC[recv.GetRoutine()].Copy(), recv}, 1, false)
	}
}

type holdObj struct {
	ch   *TraceElementChannel
	vc   map[int]*clock.VectorClock
	wvc  map[int]*clock.VectorClock
	fifo bool
}

// Update and calculate the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch (*TraceElementChannel): The trace element
//   - vc (map[int]*VectorClock): the current vector clocks
//   - wVc (map[int]*VectorClock): the current weak vector clocks
//   - fifo (bool): true if the channel buffer is assumed to be fifo
func Send(ch *TraceElementChannel, vc, wVc map[int]*clock.VectorClock, fifo bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if ch.tPost == 0 {
		vc[ch.routine].Inc(ch.routine)
		wVc[ch.routine].Inc(ch.routine)
		return
	}

	if mostRecentSend[ch.routine] == nil {
		mostRecentSend[ch.routine] = make(map[int]ElemWithVcVal)
	}

	newBufferedVCs(ch.id, ch.qSize, vc[ch.routine].GetSize())

	count := bufferedVCsCount[ch.id]

	if bufferedVCsSize[ch.id] <= count {
		holdSend = append(holdSend, holdObj{ch, vc, wVc, fifo})
		return
	}

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(bufferedVCs[ch.id]) >= count && len(bufferedVCs[ch.id]) < bufferedVCsSize[ch.id] {
		bufferedVCs[ch.id] = append(bufferedVCs[ch.id], bufferedVC{false, 0, clock.NewVectorClock(vc[ch.routine].GetSize()), 0, ""})
	}

	if count > ch.qSize || bufferedVCs[ch.id][count].occupied {
		utils.LogError("Write to occupied buffer position or to big count")
	}

	v := bufferedVCs[ch.id][count].vc
	vc[ch.routine].Sync(v)

	if fifo {
		vc[ch.routine].Sync(mostRecentSend[ch.routine][ch.id].Vc)
	}

	// for detection of send on closed
	hasSend[ch.id] = true
	mostRecentSend[ch.routine][ch.id] = ElemWithVcVal{ch, mostRecentSend[ch.routine][ch.id].Vc.Sync(vc[ch.routine]).Copy(), ch.id}

	vc[ch.routine].Inc(ch.routine)
	wVc[ch.routine].Inc(ch.routine)

	bufferedVCs[ch.id][count] = bufferedVC{true, ch.oID, vc[ch.routine].Copy(), ch.routine, ch.GetTID()}

	bufferedVCsCount[ch.id]++

	timer.Stop(timer.AnaHb)

	if analysisCases["sendOnClosed"] {
		if _, ok := closeData[ch.id]; ok {
			foundSendOnClosedChannel(ch, true)
		}
	}

	if analysisCases["selectWithoutPartner"] || modeIsFuzzing {
		CheckForSelectCaseWithoutPartnerChannel(ch, vc[ch.routine], true, true)
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(ch.routine, ch.id, elemWithVc{vc[ch.routine].Copy(), ch}, 0, true)
	}

	for i, hold := range holdRecv {
		if hold.ch.id == ch.id {
			Recv(hold.ch, hold.vc, hold.wvc, hold.fifo)
			holdRecv = append(holdRecv[:i], holdRecv[i+1:]...)
			break
		}
	}

}

// Update and calculate the vector clocks given a receive on a buffered channel.
//
// Parameter:
//   - ch (*TraceElementChannel): The trace element
//   - vc (map[int]*VectorClock): the current vector clocks
//   - wVc (map[int]*VectorClock): the current weak vector clocks
//   - fifo (bool): true if the channel buffer is assumed to be fifo
func Recv(ch *TraceElementChannel, vc, wVc map[int]*clock.VectorClock, fifo bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if analysisCases["concurrentRecv"] || analysisFuzzing {
		checkForConcurrentRecv(ch, vc)
	}

	if ch.tPost == 0 {
		vc[ch.routine].Inc(ch.routine)
		wVc[ch.routine].Inc(ch.routine)
		return
	}

	if mostRecentReceive[ch.routine] == nil {
		mostRecentReceive[ch.routine] = make(map[int]ElemWithVcVal)
	}

	newBufferedVCs(ch.id, ch.qSize, vc[ch.routine].GetSize())

	if bufferedVCsCount[ch.id] == 0 {
		holdRecv = append(holdRecv, holdObj{ch, vc, wVc, fifo})
		return
		// results.Debug("Read operation on empty buffer position", results.ERROR)
	}
	bufferedVCsCount[ch.id]--

	if bufferedVCs[ch.id][0].oID != ch.oID {
		found := false
		for i := 1; i < len(bufferedVCs[ch.id]); i++ {
			if bufferedVCs[ch.id][i].oID == ch.oID {
				found = true
				bufferedVCs[ch.id][0] = bufferedVCs[ch.id][i]
				bufferedVCs[ch.id][i] = bufferedVC{false, 0, vc[ch.routine].Copy(), 0, ""}
				break
			}
		}
		if !found {
			err := "Read operation on wrong buffer position - ID: " + strconv.Itoa(ch.id) + ", OID: " + strconv.Itoa(ch.oID) + ", SIZE: " + strconv.Itoa(ch.qSize)
			utils.LogError(err)
		}
	}
	v := bufferedVCs[ch.id][0].vc
	routSend := bufferedVCs[ch.id][0].routineSend
	tIDSend := bufferedVCs[ch.id][0].tID

	vc[ch.routine] = vc[ch.routine].Sync(v)

	if fifo {
		vc[ch.routine] = vc[ch.routine].Sync(mostRecentReceive[ch.routine][ch.id].Vc)
	}

	bufferedVCs[ch.id] = append(bufferedVCs[ch.id][1:], bufferedVC{false, 0, vc[ch.routine].Copy(), 0, ""})

	// for detection of receive on closed
	hasReceived[ch.id] = true
	mostRecentReceive[ch.routine][ch.id] = ElemWithVcVal{ch, mostRecentReceive[ch.routine][ch.id].Vc.Sync(vc[ch.routine]), ch.id}

	vc[ch.routine].Inc(ch.routine)
	wVc[ch.routine].Inc(ch.routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["selectWithoutPartner"] || modeIsFuzzing {
		CheckForSelectCaseWithoutPartnerChannel(ch, vc[ch.routine], true, true)
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(routSend, ch.routine, tIDSend, ch.GetTID())
	}
	if analysisCases["leak"] {
		CheckForLeakChannelRun(ch.routine, ch.id, elemWithVc{vc[ch.routine].Copy(), ch}, 1, true)
	}

	for i, hold := range holdSend {
		if hold.ch.id == ch.id {
			Send(hold.ch, hold.vc, hold.wvc, hold.fifo)
			holdSend = append(holdSend[:i], holdSend[i+1:]...)
			break
		}
	}
}

// Update and calculate the vector clocks for a stuck channel element
//
// Parameter:
//   - routint (int): the route of the operation
//   - vc (map[int]*VectorClock): the current vector clocks
//   - wVc (map[int]VectorClock): the current weak vector clocks
func StuckChan(routine int, vc, wVc map[int]*clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)
}

// Update and calculate the vector clocks given a close on a channel.
//
// Parameter:
//   - ch (*TraceElementChannel): The trace element
//   - vc (map[int]VectorClock): the current vector clocks
//   - wVc (map[int]VectorClock): the current weakvector clocks
func Close(ch *TraceElementChannel, vc, wVc map[int]*clock.VectorClock) {
	if ch.tPost == 0 {
		return
	}

	ch.cl = true

	if analysisCases["closeOnClosed"] {
		checkForClosedOnClosed(ch) // must be called before closePos is updated
	}

	timer.Start(timer.AnaHb)

	vc[ch.routine].Inc(ch.routine)
	wVc[ch.routine].Inc(ch.routine)

	closeData[ch.id] = ch

	timer.Stop(timer.AnaHb)

	if analysisCases["sendOnClosed"] || analysisCases["receiveOnClosed"] {
		checkForCommunicationOnClosedChannel(ch)
	}

	if analysisCases["selectWithoutPartner"] || modeIsFuzzing {
		CheckForSelectCaseWithoutPartnerClose(ch, vc[ch.routine])
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(ch.routine, ch.id, elemWithVc{vc[ch.routine].Copy(), ch}, 2, true)
	}
}

func SendC(ch *TraceElementChannel) {
	if analysisCases["sendOnClosed"] {
		foundSendOnClosedChannel(ch, true)
	}
}

// Update and calculate the vector clocks given a receive on a closed channel.
//
// Parameter:
//   - ch (*TraceElementChannel): The trace element
//   - vc (map[int]VectorClock): the current vector clocks
//   - wVc (map[int]VectorClock): the current weakvector clocks
//   - buffered (bool): true if the channel is buffered
func RecvC(ch *TraceElementChannel, vc, wVc map[int]*clock.VectorClock, buffered bool) {
	if ch.tPost == 0 {
		return
	}

	if analysisCases["receiveOnClosed"] {
		foundReceiveOnClosedChannel(ch, true)
	}

	timer.Start(timer.AnaHb)
	if _, ok := closeData[ch.id]; ok {
		vc[ch.routine] = vc[ch.routine].Sync(closeData[ch.id].vc)
	}

	vc[ch.routine].Inc(ch.routine)
	wVc[ch.routine].Inc(ch.routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["selectWithoutPartner"] || modeIsFuzzing {
		CheckForSelectCaseWithoutPartnerChannel(ch, vc[ch.routine], false, buffered)
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(closeData[ch.id].routine, ch.routine, closeData[ch.id].GetTID(), ch.GetTID())
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(ch.routine, ch.id, elemWithVc{vc[ch.routine].Copy(), ch}, 1, buffered)
	}
}

// Create a new map of buffered vector clocks for a channel if not already in
// bufferedVCs.
//
// Parameter:
//   - id (int): the id of the channel
//   - qSize (int): the buffer qSize of the channel
//   - numRout (int): the number of routines
func newBufferedVCs(id int, qSize int, numRout int) {
	if _, ok := bufferedVCs[id]; !ok {
		bufferedVCs[id] = make([]bufferedVC, 1)
		bufferedVCsCount[id] = 0
		bufferedVCsSize[id] = qSize
		bufferedVCs[id][0] = bufferedVC{false, 0, clock.NewVectorClock(numRout), 0, ""}
	}
}

// Set the channel as the last send operation.
// Used for not executed select send
//
// Parameter:
//   - id (int): the id of the channel
//   - routine (int): the route of the operation
//   - vc (VectorClock): the vector clock of the operation
//   - tID (string): the position of the send in the program
func SetChannelAsLastSend(c TraceElement) {
	if mostRecentSend[c.GetRoutine()] == nil {
		mostRecentSend[c.GetRoutine()] = make(map[int]ElemWithVcVal)
	}
	mostRecentSend[c.GetRoutine()][c.GetID()] = ElemWithVcVal{c, c.GetVC(), c.GetID()}
	hasSend[c.GetID()] = true
}

// Set the channel as the last recv operation.
// Used for not executed select recv
//
// Parameter:
//   - id (int): the id of the channel
//   - rout (int): the route of the operation
//   - vc (VectorClock): the vector clock of the operation
//   - tID (string): the position of the recv in the program
func SetChannelAsLastReceive(c TraceElement) {
	if mostRecentReceive[c.GetRoutine()] == nil {
		mostRecentReceive[c.GetRoutine()] = make(map[int]ElemWithVcVal)
	}
	mostRecentReceive[c.GetRoutine()][c.GetID()] = ElemWithVcVal{c, c.GetVC(), c.GetID()}
	hasReceived[c.GetID()] = true
}
