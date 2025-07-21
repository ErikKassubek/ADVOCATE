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
	"strconv"
)

// UpdateHBChannel updates the vector clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateHBChannel(ch *trace.ElementChannel) {
	id := ch.GetID()
	routine := ch.GetRoutine()
	oID := ch.GetOID()
	opC := ch.GetOpC()
	cl := ch.GetClosed()

	ch.SetVc(CurrentVC[routine])
	ch.SetWVc(CurrentWVC[routine])

	if ch.GetTPost() == 0 {
		return
	}

	if ch.IsBuffered() {
		switch opC {
		case trace.SendOp:
			data.MaxOpID[id] = oID
		case trace.RecvOp:
			if oID > data.MaxOpID[id] && !cl {
				data.WaitingReceive = append(data.WaitingReceive, ch)
				return
			}
		}

		switch opC {
		case trace.SendOp:
			Send(ch, CurrentVC, CurrentWVC, data.Fifo)
		case trace.RecvOp:
			if cl { // recv on closed channel
				RecvC(ch, CurrentVC, CurrentWVC, true)
			} else {
				Recv(ch, CurrentVC, CurrentWVC, data.Fifo)
			}
		case trace.CloseOp:
			Close(ch, CurrentVC, CurrentWVC)
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
			} else {
				if !cl { // stuck channel
					StuckChan(routine, CurrentVC, CurrentWVC)
				}
			}

		case trace.RecvOp: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(partner, ch)
			} else {
				if cl { // recv on closed channel
					RecvC(ch, CurrentVC, CurrentWVC, false)
				} else {
					StuckChan(routine, CurrentVC, CurrentWVC)
				}
			}
		case trace.CloseOp:
			Close(ch, CurrentVC, CurrentWVC)
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
func Send(ch *trace.ElementChannel, cl, wCl map[int]*clock.VectorClock, fifo bool) {
	id := ch.GetID()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()

	if ch.GetTPost() == 0 {
		cl[routine].Inc(routine)
		wCl[routine].Inc(routine)
		return
	}

	if data.MostRecentReceive[routine] == nil {
		data.MostRecentReceive[routine] = make(map[int]data.ElemWithVcVal)
	}

	newBufferedVCs(id, qSize)

	count := BufferedVCsCount[id]

	if BufferedVCsSize[id] <= count {
		data.HoldSend = append(data.HoldSend, data.HoldObj{
			Ch:   ch,
			Vc:   cl,
			WVc:  wCl,
			Fifo: fifo,
		})
		log.Important("APPEND: ", BufferedVCsSize[id], count)
		return
	}

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(BufferedVCs[id]) >= count && len(BufferedVCs[id]) < BufferedVCsSize[id] {
		BufferedVCs[id] = append(BufferedVCs[id], data.BufferedVC{
			Occupied: false,
			Send:     nil})
	}

	if count > qSize || BufferedVCs[id][count].Occupied {
		log.Error("Write to occupied buffer position or to big count")
	}

	s := BufferedVCs[id][count].Send
	if s != nil {
		v := s.GetVC()
		cl[routine].Sync(v)
	}

	if fifo {
		r := data.MostRecentSend[routine][id]
		if r.Elem != nil {
			cl[routine].Sync(r.Vc)
		}
	}

	cl[routine].Inc(routine)
	wCl[routine].Inc(routine)

	BufferedVCs[id][count] = data.BufferedVC{
		Occupied: true,
		Send:     ch,
	}

	BufferedVCsCount[id]++
	log.Important("+ID: ", id)
}

// Recv updates and calculates the vector clocks given a receive on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - cl map[int]*VectorClock: the current vector clocks
//   - wCl map[int]*VectorClock: the current weak vector clocks
//   - fifo bool: true if the channel buffer is assumed to be fifo
func Recv(ch *trace.ElementChannel, cl, wCl map[int]*clock.VectorClock, fifo bool) {
	id := ch.GetID()
	routine := ch.GetRoutine()
	oID := ch.GetOID()
	qSize := ch.GetQSize()

	if ch.GetTPost() == 0 {
		cl[routine].Inc(routine)
		wCl[routine].Inc(routine)
		return
	}

	if data.MostRecentReceive[routine] == nil {
		data.MostRecentReceive[routine] = make(map[int]data.ElemWithVcVal)
	}

	newBufferedVCs(id, qSize)

	if BufferedVCsCount[id] == 0 {
		data.HoldRecv = append(data.HoldRecv, data.HoldObj{
			Ch:   ch,
			Vc:   cl,
			WVc:  wCl,
			Fifo: fifo,
		})
		return
		// results.Debug("Read operation on empty buffer position", results.ERROR)
	}
	BufferedVCsCount[id]--
	log.Important("-ID: ", id)

	if BufferedVCs[id][0].Send.GetOID() != oID {
		found := false
		for i := 1; i < len(BufferedVCs[id]); i++ {
			if BufferedVCs[id][i].Send.GetOID() == oID {
				found = true
				BufferedVCs[id][0] = BufferedVCs[id][i]
				BufferedVCs[id][i] = data.BufferedVC{
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

	s := BufferedVCs[id][0].Send

	cl[routine] = cl[routine].Sync(s.GetVC())

	if fifo {
		r := data.MostRecentReceive[routine][id]
		if r.Elem != nil {
			cl[routine] = cl[routine].Sync(r.Vc)
		}
	}

	BufferedVCs[id] = append(BufferedVCs[id][1:], data.BufferedVC{
		Occupied: false,
		Send:     nil,
	})

	cl[routine].Inc(routine)
	wCl[routine].Inc(routine)

}

// StuckChan updates and calculates the vector clocks for a stuck channel element
//
// Parameter:
//   - routine int: the route of the operation
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
func StuckChan(routine int, vc, wVc map[int]*clock.VectorClock) {
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

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)
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

	if _, ok := data.CloseData[id]; ok {
		c := data.CloseData[id]
		vc[routine] = vc[routine].Sync(c.GetVC())
	}

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)
}

// Create a new map of buffered vector clocks for a channel if not already in
// vc.BufferedVCs.
//
// Parameter:
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
func newBufferedVCs(id int, qSize int) {
	if _, ok := BufferedVCs[id]; !ok {
		BufferedVCs[id] = make([]data.BufferedVC, 1)
		BufferedVCsCount[id] = 0
		BufferedVCsSize[id] = qSize
		BufferedVCs[id][0] = data.BufferedVC{
			Occupied: false,
			Send:     nil,
		}
	}
}
