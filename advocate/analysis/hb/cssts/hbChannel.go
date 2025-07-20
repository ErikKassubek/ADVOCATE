// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the cssts for channels
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package cssts

import (
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/log"
	"strconv"
)

var (
	// storage for the buffered vector spaces
	bufferedVCs = make(map[int]([]data.BufferedVC))
	// the current buffer position
	bufferedVCsCount = make(map[int]int)
	bufferedVCsSize  = make(map[int]int)
)

// UpdateHBChannel updates the vector clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateHBChannel(ch *trace.ElementChannel) {
	id := ch.GetID()
	oID := ch.GetOID()
	opC := ch.GetOpC()
	cl := ch.GetClosed()

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
			Send(ch, data.Fifo)
		case trace.RecvOp:
			if cl { // recv on closed channel
				RecvC(ch, true)
			} else {
				Recv(ch, data.Fifo)
			}
		case trace.CloseOp:
		default:
			err := "Unknown operation: " + ch.ToString()
			log.Error(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.SendOp:
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(ch, partner)
			}

		case trace.RecvOp: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(partner, ch)
			} else {
				if cl { // recv on closed channel
					RecvC(ch, false)
				}
			}
		case trace.CloseOp:
		default:
			err := "Unknown operation: " + ch.ToString()
			log.Error(err)
		}
	}
}

// UpdateHBSelect stores and updates the cssts for the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateHBSelect(se *trace.ElementSelect) {
	noChannel := se.GetChosenDefault() || se.GetTPost() == 0

	if !noChannel {
		chosenCase := se.GetChosenCase()
		UpdateHBChannel(chosenCase)
	}
}

// Unbuffered updates the cssts given a send/receive pair on a unbuffered
// channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - routSend int: the route of the sender
//   - routRecv int: the route of the receiver
//   - tID_send string: the position of the send in the program
//   - tID_recv string: the position of the receive in the program
func Unbuffered(sender trace.Element, recv trace.Element) {
	if sender.GetTPost() != 0 && recv.GetTPost() != 0 {
		AddEdge(sender, recv, false)
	}
}

// Send updates the cssts given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - fifo bool: true if the channel buffer is assumed to be fifo
func Send(ch *trace.ElementChannel, fifo bool) {
	id := ch.GetID()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()

	if ch.GetTPost() == 0 {
		return
	}

	count := bufferedVCsCount[id]

	if bufferedVCsSize[id] <= count {
		data.HoldSend = append(data.HoldSend, data.HoldObj{
			Ch:   ch,
			Vc:   nil,
			WVc:  nil,
			Fifo: fifo,
		})
		return
	}

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(bufferedVCs[id]) >= count && len(bufferedVCs[id]) < bufferedVCsSize[id] {
		bufferedVCs[id] = append(bufferedVCs[id], data.BufferedVC{
			Occupied: false,
			Send:     nil})
	}

	if count > qSize || bufferedVCs[id][count].Occupied {
		log.Error("Write to occupied buffer position or to big count")
	}

	s := bufferedVCs[id][count].Send
	if s != nil {
		AddEdge(s, ch, false)
	}

	if fifo {
		r := data.MostRecentSend[routine][id]
		if r.Elem != nil {
			AddEdge(r.Elem, ch, false)
		}
	}

	bufferedVCs[id][count] = data.BufferedVC{
		Occupied: true,
		Send:     ch,
	}

	bufferedVCsCount[id]++

	for i, hold := range data.HoldRecv {
		if hold.Ch.GetID() == id {
			Recv(hold.Ch, hold.Fifo)
			data.HoldRecv = append(data.HoldRecv[:i], data.HoldRecv[i+1:]...)
			break
		}
	}
}

// Recv updates the cssts given a receive on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - cl map[int]*VectorClock: the current vector clocks
//   - wCl map[int]*VectorClock: the current weak vector clocks
//   - fifo bool: true if the channel buffer is assumed to be fifo
func Recv(ch *trace.ElementChannel, fifo bool) {
	id := ch.GetID()
	routine := ch.GetRoutine()
	oID := ch.GetOID()
	qSize := ch.GetQSize()

	if ch.GetTPost() == 0 {
		return
	}

	if data.MostRecentReceive[routine] == nil {
		data.MostRecentReceive[routine] = make(map[int]data.ElemWithVcVal)
	}

	if bufferedVCsCount[id] == 0 {
		data.HoldRecv = append(data.HoldRecv, data.HoldObj{
			Ch:   ch,
			Vc:   nil,
			WVc:  nil,
			Fifo: fifo,
		})
		return
		// results.Debug("Read operation on empty buffer position", results.ERROR)
	}
	bufferedVCsCount[id]--

	if bufferedVCs[id][0].Send.GetOID() != oID {
		found := false
		for i := 1; i < len(bufferedVCs[id]); i++ {
			if bufferedVCs[id][i].Send.GetOID() == oID {
				found = true
				bufferedVCs[id][0] = bufferedVCs[id][i]
				bufferedVCs[id][i] = data.BufferedVC{
					Occupied: false,
					Send:     nil,
				}
				break
			}
		}
		if !found {
			err := "Read operation on wrong buffer position - ID: " + strconv.Itoa(id) + ", OID: " + strconv.Itoa(oID) + ", SIZE: " + strconv.Itoa(qSize)
			log.Error(err)
		}
	}

	s := bufferedVCs[id][0].Send
	AddEdge(s, ch, false)

	if fifo {
		r := data.MostRecentReceive[routine][id]
		if r.Elem != nil {
			AddEdge(r.Elem, ch, false)
		}
	}

	bufferedVCs[id] = append(bufferedVCs[id][1:], data.BufferedVC{
		Occupied: false,
		Send:     nil,
	})

	for i, hold := range data.HoldSend {
		if hold.Ch.GetID() == id {
			Send(hold.Ch, hold.Fifo)
			data.HoldSend = append(data.HoldSend[:i], data.HoldSend[i+1:]...)
			break
		}
	}
}

// RecvC updates the cssts given a receive on a closed channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - buffered bool: true if the channel is buffered
func RecvC(ch *trace.ElementChannel, buffered bool) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetID()

	if _, ok := data.CloseData[id]; ok {
		c := data.CloseData[id]
		AddEdge(c, ch, false)
	}
}
