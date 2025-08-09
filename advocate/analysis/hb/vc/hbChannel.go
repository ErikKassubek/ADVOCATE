// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for channels
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package vc

import (
	"advocate/analysis/data"
	"advocate/analysis/hb/clock"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateHBChannel updates the vecto clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateHBChannel(ch *trace.ElementChannel) {
	routine := ch.GetRoutine()

	ch.SetVc(CurrentVC[routine])
	ch.SetWVc(CurrentWVC[routine])

	if ch.GetTPost() == 0 {
		return
	}

	opC := ch.GetOpC()
	cl := ch.GetClosed()

	if ch.IsBuffered() {
		switch opC {
		case trace.SendOp:
			Send(ch, data.Fifo)
		case trace.RecvOp:
			if cl { // recv on closed channel
				RecvC(ch, true)
			} else {
				Recv(ch, CurrentVC, CurrentWVC, data.Fifo)
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
				partner.SetVc(CurrentVC[partnerRout])
				sel := partner.GetSelect()
				if sel != nil {
					sel.SetVc(CurrentVC[partnerRout])
				}
				Unbuffered(ch, partner)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if !cl {
					StuckChan(routine)
				}
			}

		case trace.RecvOp: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.GetRoutine()
				partner.SetVc(CurrentVC[partnerRout])
				Unbuffered(partner, ch)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if cl { // recv on closed channel
					RecvC(ch, false)
				} else {
					StuckChan(routine)
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

// UpdateHBSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateHBSelect(se *trace.ElementSelect) {
	noChannel := se.GetChosenDefault() || se.GetTPost() == 0

	routine := se.GetRoutine()

	se.SetVc(CurrentVC[routine])
	se.SetWVc(CurrentVC[routine])

	if noChannel {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
	} else {
		chosenCase := se.GetChosenCase()
		chosenCase.SetVc(se.GetVC())

		UpdateHBChannel(chosenCase)
	}

	cases := se.GetCases()

	for _, c := range cases {
		c.SetVc(se.GetVC())
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
		CurrentVC[recv.GetRoutine()].Sync(CurrentVC[sender.GetRoutine()])
		CurrentVC[sender.GetRoutine()] = CurrentVC[recv.GetRoutine()].Copy()
		CurrentVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		CurrentVC[recv.GetRoutine()].Inc(recv.GetRoutine())
		CurrentWVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		CurrentWVC[recv.GetRoutine()].Inc(recv.GetRoutine())
	} else {
		CurrentVC[sender.GetRoutine()].Inc(sender.GetRoutine())
		CurrentWVC[sender.GetRoutine()].Inc(sender.GetRoutine())
	}
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - fifo bool: true if the channel buffer is assumed to be fifo
func Send(ch *trace.ElementChannel, fifo bool) {
	routine := ch.GetRoutine()

	if ch.GetTPost() == 0 {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
		return
	}

	id := ch.GetID()
	qSize := ch.GetQSize()
	qCount := ch.GetQCount()

	if fifo {
		r := data.MostRecentSend[routine][id]
		if r.Elem != nil {
			CurrentVC[routine].Sync(r.Vc)
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
		v := s.GetVC()
		CurrentVC[routine].Sync(v)
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)

	chanBuffer[id][count] = data.BufferedVC{
		Occupied: true,
		Send:     ch,
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
	id := ch.GetID()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()

	if ch.GetTPost() == 0 {
		vc[routine].Inc(routine)
		wVc[routine].Inc(routine)
		return
	}

	newBuffer(id, qSize)
	s := chanBuffer[id][0].Send

	if s != nil {
		vc[routine] = vc[routine].Sync(s.GetVC())
	}

	if fifo {
		r := data.MostRecentReceive[routine][id]
		if r.Elem != nil {
			vc[routine] = vc[routine].Sync(r.Vc)
		}
	}

	chanBuffer[id] = append(chanBuffer[id][1:], data.BufferedVC{
		Occupied: false,
		Send:     nil,
	})

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)
}

// StuckChan updates and calculates the vector clocks for a stuck channel element
//
// Parameter:
//   - routine int: the route of the operation
func StuckChan(routine int) {
	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
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

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
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
	routine := ch.GetRoutine()

	if _, ok := data.CloseData[id]; ok {
		c := data.CloseData[id]
		CurrentVC[routine].Sync(c.GetVC())
	}

	CurrentVC[routine].Inc(routine)
	CurrentVC[routine].Inc(routine)
}

// Create a new map of buffered vector clocks for a channel if not already in
// data.BufferedVCs.
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
