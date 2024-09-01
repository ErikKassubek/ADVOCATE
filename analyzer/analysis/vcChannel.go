// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcChannel.go
// Brief: Update functions for vector clocks from channel operations
//        Some of the update function also start analysis functions
// 
// Author: Erik Kassubek <kassubek.erik@gmail.com>
// Created: 2023-07-27
// LastChange: 2024-09-01
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/logging"
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
 * 	routSend (int): the route of the sender
 * 	routRecv (int): the route of the receiver
 * 	id (int): the id of the channel
 * 	tID_send (string): the position of the send in the program
 * 	tID_recv (string): the position of the receive in the program
 * 	vc (map[int]VectorClock): the current vector clocks
 *  tPost (int): the timestamp at the end of the event
 */
func Unbuffered(routSend int, routRecv int, id int, tIDSend string,
	tIDRecv string, vc map[int]clock.VectorClock, tPost int) {
	if analysisCases["concurrentRecv"] {
		checkForConcurrentRecv(routRecv, id, tIDRecv, vc, tPost)
	}

	if tPost != 0 {

		if mostRecentReceive[routRecv] == nil {
			mostRecentReceive[routRecv] = make(map[int]VectorClockTID3)
		}
		if mostRecentSend[routSend] == nil {
			mostRecentSend[routSend] = make(map[int]VectorClockTID3)
		}

		vc[routRecv] = vc[routRecv].Sync(vc[routSend])
		vc[routSend] = vc[routRecv].Copy()
		vc[routSend] = vc[routSend].Inc(routSend)
		vc[routRecv] = vc[routRecv].Inc(routRecv)

		// for detection of send on closed
		hasSend[id] = true
		mostRecentSend[routSend][id] = VectorClockTID3{routSend, tIDSend, mostRecentSend[routSend][id].Vc.Sync(vc[routSend]).Copy(), id}

		// for detection of receive on closed
		hasReceived[id] = true
		mostRecentReceive[routRecv][id] = VectorClockTID3{routRecv, tIDRecv, mostRecentReceive[routRecv][id].Vc.Sync(vc[routRecv]).Copy(), id}

		logging.Debug("Set most recent send of "+strconv.Itoa(id)+" to "+mostRecentSend[routSend][id].Vc.ToString(), logging.DEBUG)
		logging.Debug("Set most recent recv of "+strconv.Itoa(id)+" to "+mostRecentReceive[routRecv][id].Vc.ToString(), logging.DEBUG)

	} else {
		vc[routSend] = vc[routSend].Inc(routSend)
	}

	if analysisCases["sendOnClosed"] {
		if _, ok := closeData[id]; ok {
			foundSendOnClosedChannel(routSend, id, tIDSend)
		}
	}

	if analysisCases["mixedDeadlock"] {
		CheckForSelectCaseWithoutPartnerChannel(id, vc[routSend], tIDSend, true, false)
		CheckForSelectCaseWithoutPartnerChannel(id, vc[routRecv], tIDRecv, false, false)
		checkForMixedDeadlock(routSend, routRecv, tIDSend, tIDRecv)
	}

	if analysisCases["selectWithoutPartner"] {
		CheckForSelectCaseWithoutPartnerChannel(id, vc[routSend], tIDSend, true, false)
		CheckForSelectCaseWithoutPartnerChannel(id, vc[routRecv], tIDRecv, false, false)
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(routSend, id, VectorClockTID{vc[routSend].Copy(), tIDSend, routSend}, 0, false)
		CheckForLeakChannelRun(routRecv, id, VectorClockTID{vc[routRecv].Copy(), tIDRecv, routRecv}, 1, false)
	}

}

type holdObj struct {
	rout  int
	id    int
	oID   int
	size  int
	tID   string
	vc    map[int]clock.VectorClock
	fifo  bool
	tPost int
}

var holdSend = make([]holdObj, 0)
var holdRecv = make([]holdObj, 0)

/*
 * Update and calculate the vector clocks given a send on a buffered channel.
 * Args:
 * 	rout (int): the route of the sender
 * 	id (int): the id of the sender
 * 	oID (int): the id of the communication
 * 	size (int): buffer size
 *  tId (string): the position of the send in the program
 * 	vc (map[int]VectorClock): the current vector clocks
 *  fifo (bool): true if the channel buffer is assumed to be fifo
 *  tPost (int): the timestamp at the end of the event
 */
func Send(rout int, id int, oID int, size int, tID string,
	vc map[int]clock.VectorClock, fifo bool, tPost int) {

	if tPost == 0 {
		vc[rout] = vc[rout].Inc(rout)
		return
	}

	if mostRecentSend[rout] == nil {
		mostRecentSend[rout] = make(map[int]VectorClockTID3)
	}

	newBufferedVCs(id, size, vc[rout].GetSize())

	count := bufferedVCsCount[id]

	if len(bufferedVCs[id]) <= count {
		holdSend = append(holdSend, holdObj{rout, id, oID, size, tID, vc, fifo, tPost})
		return
		// panic("BufferedVCsCount is bigger than the buffer size for chan " + strconv.Itoa(id) + " with count " + strconv.Itoa(count) + " and size " + strconv.Itoa(size) + "\n\tand tID " + tID)
	}

	if count > size || bufferedVCs[id][count].occupied {
		logging.Debug("Write to occupied buffer position or to big count", logging.ERROR)
	}

	v := bufferedVCs[id][count].vc
	vc[rout] = vc[rout].Sync(v)

	if fifo {
		vc[rout] = vc[rout].Sync(mostRecentSend[rout][id].Vc)
	}

	bufferedVCs[id][count] = bufferedVC{true, oID, vc[rout].Copy(), rout, tID}

	bufferedVCsCount[id]++

	// for detection of send on closed
	hasSend[id] = true
	mostRecentSend[rout][id] = VectorClockTID3{rout, tID, mostRecentSend[rout][id].Vc.Sync(vc[rout]), id}

	vc[rout] = vc[rout].Inc(rout)

	if analysisCases["sendOnClosed"] {
		if _, ok := closeData[id]; ok {
			foundSendOnClosedChannel(rout, id, tID)
		}
	}

	if analysisCases["selectWithoutPartner"] {
		CheckForSelectCaseWithoutPartnerChannel(id, vc[rout], tID, true, true)
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(rout, id, VectorClockTID{vc[rout].Copy(), tID, rout}, 0, true)
	}

	for i, hold := range holdRecv {
		if hold.id == id {
			Recv(hold.rout, hold.id, hold.oID, hold.size, hold.tID, hold.vc, hold.fifo, hold.tPost)
			holdRecv = append(holdRecv[:i], holdRecv[i+1:]...)
			break
		}
	}

}

/*
 * Update and calculate the vector clocks given a receive on a buffered channel.
 * Args:
 * 	rout (int): the route of the sender
 * 	id (int): the id of the sender
 * 	oId (int): the id of the communication
 * 	size (int): buffer size
 *  tID (string): the position of the send in the program
 * 	vc (map[int]VectorClock): the current vector clocks
 *  fifo (bool): true if the channel buffer is assumed to be fifo
 *  tPost (int): the timestamp at the end of the event
 */
func Recv(rout int, id int, oID, size int, tID string, vc map[int]clock.VectorClock,
	fifo bool, tPost int) {

	if analysisCases["concurrentRecv"] {
		checkForConcurrentRecv(rout, id, tID, vc, tPost)
	}

	if tPost == 0 {
		vc[rout] = vc[rout].Inc(rout)
		return
	}

	if mostRecentReceive[rout] == nil {
		mostRecentReceive[rout] = make(map[int]VectorClockTID3)
	}

	newBufferedVCs(id, size, vc[rout].GetSize())

	if bufferedVCsCount[id] == 0 {
		holdSend = append(holdSend, holdObj{rout, id, oID, size, tID, vc, fifo, tPost})
		return
		// logging.Debug("Read operation on empty buffer position", logging.ERROR)
	}
	bufferedVCsCount[id]--

	if bufferedVCs[id][0].oID != oID {
		found := false
		for i := 1; i < size; i++ {
			if bufferedVCs[id][i].oID == oID {
				found = true
				bufferedVCs[id][0] = bufferedVCs[id][i]
				bufferedVCs[id][i] = bufferedVC{false, 0, vc[rout].Copy(), 0, ""}
				break
			}
		}
		if !found {
			err := "Read operation on wrong buffer position - ID: " + strconv.Itoa(id) + ", OID: " + strconv.Itoa(oID) + ", SIZE: " + strconv.Itoa(size)
			logging.Debug(err, logging.INFO)
		}
	}
	v := bufferedVCs[id][0].vc
	routSend := bufferedVCs[id][0].routineSend
	tIDSend := bufferedVCs[id][0].tID

	vc[rout] = vc[rout].Sync(v)

	if fifo {
		vc[rout] = vc[rout].Sync(mostRecentReceive[rout][id].Vc)
	}

	bufferedVCs[id] = bufferedVCs[id][1:]
	bufferedVCs[id] = append(bufferedVCs[id], bufferedVC{false, 0, vc[rout].Copy(), 0, ""})

	// for detection of receive on closed
	hasReceived[id] = true
	mostRecentReceive[rout][id] = VectorClockTID3{rout, tID, mostRecentReceive[rout][id].Vc.Sync(vc[rout]), id}

	vc[rout] = vc[rout].Inc(rout)

	if analysisCases["selectWithoutPartner"] {
		CheckForSelectCaseWithoutPartnerChannel(id, vc[rout], tID, true, true)
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(routSend, rout, tIDSend, tID)
	}
	if analysisCases["leak"] {
		CheckForLeakChannelRun(rout, id, VectorClockTID{vc[rout].Copy(), tID, rout}, 1, true)
	}

	for i, hold := range holdSend {
		if hold.id == id {
			Send(hold.rout, hold.id, hold.oID, hold.size, hold.tID, hold.vc, hold.fifo, hold.tPost)
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
 * 	rout (int): the route of the operation
 * 	id (int): the id of the channel
 * 	tID (string): the position of the close in the program
 * 	vc (map[int]VectorClock): the current vector clocks
 *  tPost (int): the timestamp at the end of the event
 *  buffered (bool): true if the channel is buffered
 */
func Close(rout int, id int, tID string, vc map[int]clock.VectorClock, tPost int, buffered bool) {
	if tPost == 0 {
		return
	}

	if analysisCases["closeOnClosed"] {
		checkForClosedOnClosed(rout, id, tID) // must be called before closePos is updated
	}

	vc[rout] = vc[rout].Inc(rout)

	closeData[id] = VectorClockTID3{Routine: rout, TID: tID, Vc: vc[rout].Copy(), Val: id}

	if analysisCases["sendOnClosed"] || analysisCases["receiveOnClosed"] {
		checkForCommunicationOnClosedChannel(id, tID)
	}

	if analysisCases["selectWithoutPartner"] {
		CheckForSelectCaseWithoutPartnerClose(id, vc[rout])
	}

	if analysisCases["leak"] {
		CheckForLeakChannelRun(rout, id, VectorClockTID{vc[rout].Copy(), tID, rout}, 2, true)
	}
}

func SendC(rout int, id int, tID string) {
	if analysisCases["sendOnClosed"] {
		foundSendOnClosedChannel(rout, id, tID)
	}
}

/*
 * Update and calculate the vector clocks given a receive on a closed channel.
 * Args:
 * 	rout (int): the route of the sender
 * 	id (int): the id of the sender
 * 	tID (string): the position of the close in the program
 * 	vc (map[int]VectorClock): the current vector clocks
 *  tPost (int): the timestamp at the end of the event
 *  buffered (bool): true if the channel is buffered
 */
func RecvC(rout int, id int, tID string, vc map[int]clock.VectorClock, tPost int,
	buffered bool) {
	if tPost == 0 {
		return
	}

	if analysisCases["receiveOnClosed"] {
		foundReceiveOnClosedChannel(rout, id, tID)
	}

	vc[rout] = vc[rout].Sync(closeData[id].Vc)
	vc[rout] = vc[rout].Inc(rout)

	if analysisCases["selectWithoutPartner"] {
		CheckForSelectCaseWithoutPartnerChannel(id, vc[rout], tID, false, buffered)
	}

	if analysisCases["mixedDeadlock"] {
		checkForMixedDeadlock(closeData[id].Routine, rout, closeData[id].TID, tID)
	}
	if analysisCases["leak"] {
		CheckForLeakChannelRun(rout, id, VectorClockTID{vc[rout].Copy(), tID, rout}, 1, buffered)
	}
}

/*
 * Create a new map of buffered vector clocks for a channel if not already in
 * bufferedVCs.
 * Args:
 * 	id (int): the id of the channel
 * 	size (int): the buffer size of the channel
 * 	numRout (int): the number of routines
 */
func newBufferedVCs(id int, size int, numRout int) {
	if _, ok := bufferedVCs[id]; !ok {
		bufferedVCs[id] = make([]bufferedVC, size)
		for i := 0; i < size; i++ {
			bufferedVCsCount[id] = 0
			bufferedVCs[id][i] = bufferedVC{false, 0, clock.NewVectorClock(numRout), 0, ""}
		}
	}
}

/*
 * Set the channel as the last send operation.
 * Used for not executed select send
 * Args:
 * 	id (int): the id of the channel
 * 	rout (int): the route of the operation
 *  vc (VectorClock): the vector clock of the operation
 *  tID (string): the position of the send in the program
 */
func SetChannelAsLastSend(id int, rout int, vc clock.VectorClock, tID string) {
	if mostRecentSend[rout] == nil {
		mostRecentSend[rout] = make(map[int]VectorClockTID3)
	}
	mostRecentSend[rout][id] = VectorClockTID3{rout, tID, vc, id}
	hasSend[id] = true
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
func SetChannelAsLastReceive(id int, rout int, vc clock.VectorClock, tID string) {
	if mostRecentReceive[rout] == nil {
		mostRecentReceive[rout] = make(map[int]VectorClockTID3)
	}
	mostRecentReceive[rout][id] = VectorClockTID3{rout, tID, vc, id}
	hasReceived[id] = true
}
