// Copyrigth (c) 2024 Erik Kassubek
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
	timemeasurement "analyzer/timeMeasurement"
	"log"
	"strconv"
)

// elements for buffered channel internal vector clock
type bufferedVC struct {
	occupied    bool
	oID         int
	vc          clock.VectorClock
	routineSend int
	tID         string
}

/*
 * Update and calculate the vector clocks given a send/receive pair on a unbuffered
 * channel.
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 * 	routSend (int): the route of the sender
 * 	routRecv (int): the route of the receiver
 * 	tID_send (string): the position of the send in the program
 * 	tID_recv (string): the position of the receive in the program
 * 	vc (map[int]VectorClock): the current vector clocks
 */
func Unbuffered(sender TraceElement, recv TraceElement, vc map[int]clock.VectorClock) {
	if analysisCases["concurrentRecv"] {
		timemeasurement.Start("other")
		switch r := recv.(type) {
		case *TraceElementChannel:
			checkForConcurrentRecv(r, vc)
		case *TraceElementSelect:
			checkForConcurrentRecv(&r.chosenCase, vc)
		}
		timemeasurement.Start("End")
	}

	if sender.GetTPost() != 0 && recv.GetTPost() != 0 {

		if mostRecentReceive[recv.GetRoutine()] == nil {
			mostRecentReceive[recv.GetRoutine()] = make(map[int]VectorClockTID3)
		}
		if mostRecentSend[sender.GetRoutine()] == nil {
			mostRecentSend[sender.GetRoutine()] = make(map[int]VectorClockTID3)
		}

		// for detection of send on closed
		hasSend[sender.GetID()] = true
		mostRecentSend[sender.GetRoutine()][sender.GetID()] = VectorClockTID3{sender, mostRecentSend[sender.GetRoutine()][sender.GetID()].Vc.Sync(vc[sender.GetRoutine()]).Copy(), sender.GetID()}

		// for detection of receive on closed
		hasReceived[sender.GetID()] = true
		mostRecentReceive[recv.GetRoutine()][sender.GetID()] = VectorClockTID3{recv, mostRecentReceive[recv.GetRoutine()][sender.GetID()].Vc.Sync(vc[recv.GetRoutine()]).Copy(), sender.GetID()}

		vc[recv.GetRoutine()] = vc[recv.GetRoutine()].Sync(vc[sender.GetRoutine()])
		vc[sender.GetRoutine()] = vc[recv.GetRoutine()].Copy()
		vc[sender.GetRoutine()] = vc[sender.GetRoutine()].Inc(sender.GetRoutine())
		vc[recv.GetRoutine()] = vc[recv.GetRoutine()].Inc(recv.GetRoutine())

	} else {
		vc[sender.GetRoutine()] = vc[sender.GetRoutine()].Inc(sender.GetRoutine())
	}

	if analysisCases["sendOnClosed"] {
		timemeasurement.Start("panic")
		if _, ok := closeData[sender.GetID()]; ok {
			foundSendOnClosedChannel(sender.GetRoutine(), sender.GetID(), sender.GetTID(), true)
		}
		timemeasurement.End("panic")
	}

	timemeasurement.Start("other")
	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(sender.GetRoutine(), recv.GetRoutine(), sender.GetTID(), recv.GetTID())
	}

	if analysisCases["selectWithoutPartner"] || runFuzzing {
		CheckForSelectCaseWithoutPartnerChannel(sender, vc[sender.GetRoutine()], true, false)
		CheckForSelectCaseWithoutPartnerChannel(recv, vc[recv.GetRoutine()], false, false)
	}
	timemeasurement.End("other")

	if analysisCases["leak"] {
		timemeasurement.Start("leak")
		CheckForLeakChannelRun(sender.GetRoutine(), sender.GetID(), VectorClockTID{vc[sender.GetRoutine()].Copy(), sender.GetTID(), sender.GetRoutine()}, 0, false)
		CheckForLeakChannelRun(recv.GetRoutine(), sender.GetID(), VectorClockTID{vc[recv.GetRoutine()].Copy(), recv.GetTID(), recv.GetRoutine()}, 1, false)
		timemeasurement.End("leak")
	}
}

type holdObj struct {
	ch   *TraceElementChannel
	vc   map[int]clock.VectorClock
	fifo bool
}

var holdSend = make([]holdObj, 0)
var holdRecv = make([]holdObj, 0)

/*
 * Update and calculate the vector clocks given a send on a buffered channel.
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 * 	vc (map[int]VectorClock): the current vector clocks
 *  fifo (bool): true if the channel buffer is assumed to be fifo
 */
func Send(ch *TraceElementChannel, vc map[int]clock.VectorClock, fifo bool) {

	if ch.tPost == 0 {
		vc[ch.routine] = vc[ch.routine].Inc(ch.routine)
		return
	}

	if mostRecentSend[ch.routine] == nil {
		mostRecentSend[ch.routine] = make(map[int]VectorClockTID3)
	}

	newBufferedVCs(ch.id, ch.qSize, vc[ch.routine].GetSize())

	count := bufferedVCsCount[ch.id]

	if bufferedVCsSize[ch.id] <= count {
		holdSend = append(holdSend, holdObj{ch, vc, fifo})
		return
		// panic("BufferedVCsCount is bigger than the buffer qSize for chan " + strconv.Itoa(id) + " with count " + strconv.Itoa(count) + " and qSize " + strconv.Itoa(qSize) + "\n\tand tID " + tID)
	}

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(bufferedVCs[ch.id]) >= count && len(bufferedVCs[ch.id]) < bufferedVCsSize[ch.id] {
		bufferedVCs[ch.id] = append(bufferedVCs[ch.id], bufferedVC{false, 0, clock.NewVectorClock(vc[ch.routine].GetSize()), 0, ""})
	}

	if count > ch.qSize || bufferedVCs[ch.id][count].occupied {
		log.Print("Write to occupied buffer position or to big count")
	}

	v := bufferedVCs[ch.id][count].vc
	vc[ch.routine] = vc[ch.routine].Sync(v)

	if fifo {
		vc[ch.routine] = vc[ch.routine].Sync(mostRecentSend[ch.routine][ch.id].Vc)
	}

	// for detection of send on closed
	hasSend[ch.id] = true
	mostRecentSend[ch.routine][ch.id] = VectorClockTID3{ch, mostRecentSend[ch.routine][ch.id].Vc.Sync(vc[ch.routine]), ch.id}

	vc[ch.routine] = vc[ch.routine].Inc(ch.routine)

	bufferedVCs[ch.id][count] = bufferedVC{true, ch.oID, vc[ch.routine].Copy(), ch.routine, ch.GetTID()}

	bufferedVCsCount[ch.id]++

	if analysisCases["sendOnClosed"] {
		timemeasurement.Start("panic")
		if _, ok := closeData[ch.id]; ok {
			foundSendOnClosedChannel(ch.routine, ch.id, ch.GetTID(), true)
		}
		timemeasurement.End("panic")
	}

	timemeasurement.Start("other")
	if analysisCases["selectWithoutPartner"] || runFuzzing {
		CheckForSelectCaseWithoutPartnerChannel(ch, vc[ch.routine], true, true)
	}
	timemeasurement.Start("other")

	if analysisCases["leak"] {
		timemeasurement.Start("leak")
		CheckForLeakChannelRun(ch.routine, ch.id, VectorClockTID{vc[ch.routine].Copy(), ch.GetTID(), ch.routine}, 0, true)
		timemeasurement.End("leak")
	}

	for i, hold := range holdRecv {
		if hold.ch.id == ch.id {
			Recv(hold.ch, hold.vc, hold.fifo)
			holdRecv = append(holdRecv[:i], holdRecv[i+1:]...)
			break
		}
	}

}

/*
 * Update and calculate the vector clocks given a receive on a buffered channel.
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 * 	vc (map[int]VectorClock): the current vector clocks
 *  fifo (bool): true if the channel buffer is assumed to be fifo
 */
func Recv(ch *TraceElementChannel, vc map[int]clock.VectorClock, fifo bool) {

	if analysisCases["concurrentRecv"] {
		timemeasurement.Start("other")
		checkForConcurrentRecv(ch, vc)
		timemeasurement.End("other")
	}

	if ch.tPost == 0 {
		vc[ch.routine] = vc[ch.routine].Inc(ch.routine)
		return
	}

	if mostRecentReceive[ch.routine] == nil {
		mostRecentReceive[ch.routine] = make(map[int]VectorClockTID3)
	}

	newBufferedVCs(ch.id, ch.qSize, vc[ch.routine].GetSize())

	if bufferedVCsCount[ch.id] == 0 {
		holdSend = append(holdSend, holdObj{ch, vc, fifo})
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
			log.Print(err)
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
	mostRecentReceive[ch.routine][ch.id] = VectorClockTID3{ch, mostRecentReceive[ch.routine][ch.id].Vc.Sync(vc[ch.routine]), ch.id}

	vc[ch.routine] = vc[ch.routine].Inc(ch.routine)

	if analysisCases["selectWithoutPartner"] || runFuzzing {
		timemeasurement.Start("other")
		CheckForSelectCaseWithoutPartnerChannel(ch, vc[ch.routine], true, true)
		timemeasurement.End("other")
	}

	if analysisCases["mixedDeadlock"] {
		timemeasurement.Start("other")
		checkForMixedDeadlock(routSend, ch.routine, tIDSend, ch.GetTID())
		timemeasurement.End("other")
	}
	if analysisCases["leak"] {
		timemeasurement.Start("leak")
		CheckForLeakChannelRun(ch.routine, ch.id, VectorClockTID{vc[ch.routine].Copy(), ch.GetTID(), ch.routine}, 1, true)
		timemeasurement.End("leak")
	}

	for i, hold := range holdSend {
		if hold.ch.id == ch.id {
			Send(hold.ch, hold.vc, hold.fifo)
			holdSend = append(holdSend[:i], holdSend[i+1:]...)
			break
		}
	}
}

/*
 * Update and calculate the vector clocks for a stuck channel element
 * Args:
 *  routint (int): the route of the operation
 *  vc (map[int]VectorClock): the current vector clocks
 */
func StuckChan(routine int, vc map[int]clock.VectorClock) {
	vc[routine] = vc[routine].Inc(routine)
}

/*
 * Update and calculate the vector clocks given a close on a channel.
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 * 	vc (map[int]VectorClock): the current vector clocks
 */
func Close(ch *TraceElementChannel, vc map[int]clock.VectorClock) {
	if ch.tPost == 0 {
		return
	}

	ch.cl = true

	if analysisCases["closeOnClosed"] {
		timemeasurement.Start("other")
		checkForClosedOnClosed(ch) // must be called before closePos is updated
		timemeasurement.End("other")
	}

	vc[ch.routine] = vc[ch.routine].Inc(ch.routine)

	closeData[ch.id] = ch

	if analysisCases["sendOnClosed"] || analysisCases["receiveOnClosed"] {
		checkForCommunicationOnClosedChannel(ch)
	}

	if analysisCases["selectWithoutPartner"] || runFuzzing {
		timemeasurement.Start("other")
		CheckForSelectCaseWithoutPartnerClose(ch, vc[ch.routine])
		timemeasurement.Start("other")
	}

	if analysisCases["leak"] {
		timemeasurement.Start("leak")
		CheckForLeakChannelRun(ch.routine, ch.id, VectorClockTID{vc[ch.routine].Copy(), ch.GetTID(), ch.routine}, 2, true)
		timemeasurement.End("leak")
	}
}

func SendC(ch *TraceElementChannel) {
	timemeasurement.Start("other")
	if analysisCases["sendOnClosed"] {
		foundSendOnClosedChannel(ch.routine, ch.id, ch.GetTID(), true)
	}
	timemeasurement.End("other")
}

/*
 * Update and calculate the vector clocks given a receive on a closed channel.
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 * 	vc (map[int]VectorClock): the current vector clocks
 *  buffered (bool): true if the channel is buffered
 */
func RecvC(ch *TraceElementChannel, vc map[int]clock.VectorClock, buffered bool) {
	if ch.tPost == 0 {
		return
	}

	if analysisCases["receiveOnClosed"] {
		foundReceiveOnClosedChannel(ch, true)
	}

	if _, ok := closeData[ch.id]; ok {
		vc[ch.routine] = vc[ch.routine].Sync(closeData[ch.id].vc)
	}
	vc[ch.routine] = vc[ch.routine].Inc(ch.routine)

	timemeasurement.Start("other")

	if analysisCases["selectWithoutPartner"] || runFuzzing {
		CheckForSelectCaseWithoutPartnerChannel(ch, vc[ch.routine], false, buffered)
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(closeData[ch.id].routine, ch.routine, closeData[ch.id].GetTID(), ch.GetTID())
	}
	timemeasurement.End("other")

	if analysisCases["leak"] {
		timemeasurement.Start("leak")
		CheckForLeakChannelRun(ch.routine, ch.id, VectorClockTID{vc[ch.routine].Copy(), ch.GetTID(), ch.routine}, 1, buffered)
		timemeasurement.End("leak")
	}
}

/*
 * Create a new map of buffered vector clocks for a channel if not already in
 * bufferedVCs.
 * Args:
 * 	id (int): the id of the channel
 * 	qSize (int): the buffer qSize of the channel
 * 	numRout (int): the number of routines
 */
func newBufferedVCs(id int, qSize int, numRout int) {
	if _, ok := bufferedVCs[id]; !ok {
		bufferedVCs[id] = make([]bufferedVC, 1)
		bufferedVCsCount[id] = 0
		bufferedVCsSize[id] = qSize
		bufferedVCs[id][0] = bufferedVC{false, 0, clock.NewVectorClock(numRout), 0, ""}
	}
}

/*
 * Set the channel as the last send operation.
 * Used for not executed select send
 * Args:
 * 	id (int): the id of the channel
 * 	routine (int): the route of the operation
 *  vc (VectorClock): the vector clock of the operation
 *  tID (string): the position of the send in the program
 */
func SetChannelAsLastSend(c TraceElement) {
	if mostRecentSend[c.GetRoutine()] == nil {
		mostRecentSend[c.GetRoutine()] = make(map[int]VectorClockTID3)
	}
	mostRecentSend[c.GetRoutine()][c.GetID()] = VectorClockTID3{c, c.GetVC(), c.GetID()}
	hasSend[c.GetID()] = true
}

/*
 * Set the channel as the last recv operation.
 * Used for not executed select recv
 * Args:
 * 	id (int): the id of the channel
 * 	rout (int): the route of the operation
 *  vc (VectorClock): the vector clock of the operation
 *  tID (string): the position of the recv in the program
 */
func SetChannelAsLastReceive(c TraceElement) {
	if mostRecentReceive[c.GetRoutine()] == nil {
		mostRecentReceive[c.GetRoutine()] = make(map[int]VectorClockTID3)
	}
	mostRecentReceive[c.GetRoutine()][c.GetID()] = VectorClockTID3{c, c.GetVC(), c.GetID()}
	hasReceived[c.GetID()] = true
}
