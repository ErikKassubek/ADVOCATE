// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for channels
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_vc

import (
	"gocct/analysis/a_base"
	"gocct/analysis/hb/a_clock"
	"gocct/trace"
	"gocct/utils/log"
)

// UpdateHBChannel updates the vector clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateHBChannel(ch *trace.ElementChannel) {
	routine := ch.Routine()

	ch.Vc(a_clock.Strong, CurrentVC[routine])
	ch.Vc(a_clock.Weak, CurrentWVC[routine])

	if !ch.Committed() {
		return
	}

	opC := ch.Type(true)
	cl := ch.GetClosed()

	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			Send(ch)
		case trace.ChannelRecv:
			if cl { // recv on closed channel
				RecvC(ch, true)
			} else {
				Recv(ch, CurrentVC, CurrentWVC)
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
				partner.Vc(a_clock.Strong, CurrentVC[partnerRout])
				sel := partner.GetSelect()
				if sel != nil {
					sel.Vc(a_clock.Strong, CurrentVC[partnerRout])
				}
				Unbuffered(ch, partner)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if !cl {
					StuckChan(routine)
				}
			}

		case trace.ChannelRecv: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.Routine()
				partner.Vc(a_clock.Strong, CurrentVC[partnerRout])
				Unbuffered(partner, ch)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if cl { // recv on closed channel
					RecvC(ch, false)
				} else {
					StuckChan(routine)
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

// UpdateHBSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateHBSelect(se *trace.ElementSelect) {
	noChannel := se.GetChosenDefault() || !se.Committed()

	routine := se.Routine()

	se.Vc(a_clock.Strong, CurrentVC[routine])
	se.Vc(a_clock.Weak, CurrentVC[routine])

	if noChannel {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
	} else {
		chosenCase := se.GetChosenCase()
		chosenCase.Vc(a_clock.Strong, se.GetVC(a_clock.Strong))

		UpdateHBChannel(chosenCase)
	}

	cases := se.GetCases()

	for _, c := range cases {
		c.Vc(a_clock.Strong, se.GetVC(a_clock.Strong))
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
	if sender.Committed() && recv.Committed() {
		CurrentVC[recv.Routine()].Sync(CurrentVC[sender.Routine()])
		CurrentVC[sender.Routine()] = CurrentVC[recv.Routine()].Copy()
		CurrentVC[sender.Routine()].Inc(sender.Routine())
		CurrentVC[recv.Routine()].Inc(recv.Routine())
		CurrentWVC[sender.Routine()].Inc(sender.Routine())
		CurrentWVC[recv.Routine()].Inc(recv.Routine())
	} else {
		CurrentVC[sender.Routine()].Inc(sender.Routine())
		CurrentWVC[sender.Routine()].Inc(sender.Routine())
	}
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func Send(ch *trace.ElementChannel) {
	routine := ch.Routine()

	if !ch.Committed() {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
		return
	}

	id := ch.ObjID()
	qSize := ch.GetQSize()
	qCount := ch.GetQCount()

	r := a_base.MostRecentSend[routine][id]
	if r.Elem != nil {
		CurrentVC[routine].Sync(r.Vc)
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
	if len(chanBuffer[id]) >= count {
		chanBuffer[id] = append(chanBuffer[id], a_base.BufferedVC{
			Occupied: false,
			Send:     nil})
	}

	// if count > qSize || chanBuffer[id][count].Occupied {
	// 	log.Error("Write to occupied buffer position or to big count")
	// }

	s := chanBuffer[id][count].Send
	if s != nil {
		v := s.GetVC(a_clock.Strong)
		CurrentVC[routine].Sync(v)
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)

	chanBuffer[id][count] = a_base.BufferedVC{
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
func Recv(ch *trace.ElementChannel, vc, wVc map[int]*a_clock.VectorClock) {
	id := ch.ObjID()
	routine := ch.Routine()
	qSize := ch.GetQSize()

	if !ch.Committed() {
		vc[routine].Inc(routine)
		wVc[routine].Inc(routine)
		return
	}

	newBuffer(id, qSize)
	s := chanBuffer[id][0].Send

	if s != nil {
		vc[routine] = vc[routine].Sync(s.GetVC(a_clock.Strong))
	}

	r := a_base.MostRecentReceive[routine][id]
	if r.Elem != nil {
		vc[routine] = vc[routine].Sync(r.Vc)
	}

	chanBuffer[id] = append(chanBuffer[id][1:], a_base.BufferedVC{
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
	if !ch.Committed() {
		return
	}

	routine := ch.Routine()

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// RecvC updates and calculates the vector clocks given a receive on a closed channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - buffered bool: true if the channel is buffered
func RecvC(ch *trace.ElementChannel, buffered bool) {
	if !ch.Committed() {
		return
	}

	id := ch.ObjID()
	routine := ch.Routine()

	if _, ok := a_base.CloseData[id]; ok {
		c := a_base.CloseData[id]
		CurrentVC[routine].Sync(c.GetVC(a_clock.Strong))
	}

	CurrentVC[routine].Inc(routine)
	CurrentVC[routine].Inc(routine)
}

// Create a new map of buffered vector clocks for a channel if not already in
// baseA.BufferedVCs.
//
// Parameter:
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
func newBuffer(id int, qSize int) {
	if _, ok := chanBuffer[id]; !ok {
		chanBuffer[id] = make([]a_base.BufferedVC, 1)
		chanBufferSize[id] = qSize
		chanBuffer[id][0] = a_base.BufferedVC{
			Occupied: false,
			Send:     nil,
		}
	}
}
