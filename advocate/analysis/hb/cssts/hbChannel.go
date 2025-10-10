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
	"advocate/utils/flags"
	"advocate/utils/log"
)

var (
	chanBuffer     = make(map[int]([]data.BufferedVC))
	chanBufferSize = make(map[int]int)
)

// UpdateHBChannel updates the csst to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateHBChannel(ch *trace.ElementChannel) {
	if ch.GetTPost() == 0 {
		return
	}

	opC := ch.GetType(true)
	cl := ch.GetClosed()

	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			Send(ch)
		case trace.ChannelRecv:
			if cl { // recv on closed channel
				RecvC(ch, true)
			} else {
				Recv(ch)
			}
		case trace.ChannelClose:
		default:
			err := "Unknown operation: " + ch.ToString()
			log.Error(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.ChannelSend:
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(ch, partner)
				// increase index for recv is done in analysis/elements/channel.go
			}

		case trace.ChannelRecv: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(partner, ch)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if cl { // recv on closed channel
					RecvC(ch, false)
				}
			}
		case trace.ChannelClose:
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

	if !noChannel {
		chosenCase := se.GetChosenCase()
		chosenCase.SetVc(se.GetVC())

		UpdateHBChannel(chosenCase)
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
	if sender.GetTPost() != 0 && recv.GetTPost() != 0 {
		AddEdge(sender, recv, false)
	}
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func Send(ch *trace.ElementChannel) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetID()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()
	qCount := ch.GetQCount()

	if !flags.IgnoreFifo {
		r := data.MostRecentSend[routine][id]
		if r.Elem != nil {
			AddEdge(r.Elem, ch, false)
		}
	}

	// direct communication without using the buffer
	if qCount == 0 {
		return
	}

	newBuffer(id, qSize)

	count := qCount - 1

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(chanBuffer[id]) >= count && len(chanBuffer[id]) < chanBufferSize[id] {
		chanBuffer[id] = append(chanBuffer[id], data.BufferedVC{
			Occupied: false,
			Send:     nil})
	}

	// if count > qSize || chanBuffer[id][count].Occupied {
	// 	log.Error("Write to occupied buffer position or to big count")
	// }

	s := chanBuffer[id][count].Send
	if s != nil {
		AddEdge(s, ch, false)
	}

	if !flags.IgnoreFifo {
		r := data.MostRecentSend[routine][id]
		if r.Elem != nil {
			AddEdge(r.Elem, ch, false)
		}
	}

	chanBuffer[id][count] = data.BufferedVC{
		Occupied: true,
		Send:     ch,
	}
}

// Recv updates and calculates the vector clocks given a receive on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func Recv(ch *trace.ElementChannel) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetID()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()

	newBuffer(id, qSize)

	s := chanBuffer[id][0].Send

	if s != nil {
		AddEdge(s, ch, false)
	}

	if !flags.IgnoreFifo {
		r := data.MostRecentReceive[routine][id]
		if r.Elem != nil {
			AddEdge(r.Elem, ch, false)
		}
	}

	chanBuffer[id] = append(chanBuffer[id][1:], data.BufferedVC{
		Occupied: false,
		Send:     nil,
	})
}

// RecvC updates and calculates the vector clocks given a receive on a closed channel.
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

// Create a new map of buffered vector clocks for a channel if not already in
// data.bufferedVCs.
//
// Parameter:
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
func newBuffer(id int, qSize int) {
	if _, ok := chanBuffer[id]; !ok {
		chanBuffer[id] = make([]data.BufferedVC, 1)
		chanBufferSize[id] = qSize
		chanBuffer[id][0] = data.BufferedVC{
			Occupied: false,
			Send:     nil,
		}
	}
}
