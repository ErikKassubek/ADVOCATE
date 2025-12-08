// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for channels
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/baseA"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
)

var (
	chanBuffer     = make(map[int]([]baseA.BufferedVC))
	chanBufferSize = make(map[int]int)
)

// UpdateHBChannel updates the vector clocks to a channel element
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *trace.TraceElementChannel: the channel element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func UpdateHBChannel(graph *PoGraph, ch *trace.ElementChannel, recorded bool) {
	if recorded && ch.GetTPost() == 0 {
		return
	}

	opC := ch.GetType(true)
	cl := ch.GetClosed()

	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			Send(graph, ch)
		case trace.ChannelRecv:
			if cl { // recv on closed channel
				RecvC(graph, ch, true)
			} else {
				Recv(graph, ch)
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
				Unbuffered(graph, ch, partner)
				// increase index for recv is done in analysis/elements/channel.go
			}

		case trace.ChannelRecv: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(graph, partner, ch)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if cl { // recv on closed channel
					RecvC(graph, ch, false)
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
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - se *trace.TraceElementSelect: the select element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func UpdateHBSelect(graph *PoGraph, se *trace.ElementSelect, recorded bool) {
	noChannel := se.GetChosenDefault() || se.GetTPost() == 0

	if !noChannel {
		chosenCase := se.GetChosenCase()
		chosenCase.SetVc(se.GetVC())

		UpdateHBChannel(graph, chosenCase, recorded)
	}
}

// Unbuffered updates and calculates the vector clocks given a send/receive pair on a unbuffered
// channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - sender trace.Element: sender node
//   - recv trace.Element: receiver node
func Unbuffered(graph *PoGraph, sender trace.Element, recv trace.Element) {
	if sender.GetTPost() != 0 && recv.GetTPost() != 0 {
		if graph != nil {
			graph.AddEdge(sender, recv)
		} else {
			AddEdge(sender, recv, false)
		}
	}
}

// Send updates and calculates the pog given a send on a buffered channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *TraceElementChannel: The trace element
func Send(graph *PoGraph, ch *trace.ElementChannel) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetObjId()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()
	qCount := ch.GetQCount()

	if !flags.IgnoreFifo {
		r := baseA.MostRecentSend[routine][id]
		if r.Elem != nil {
			if graph != nil {
				graph.AddEdge(r.Elem, ch)
			} else {
				AddEdge(r.Elem, ch, false)
			}
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
		chanBuffer[id] = append(chanBuffer[id], baseA.BufferedVC{
			Occupied: false,
			Send:     nil})
	}

	// if count > qSize || chanBuffer[id][count].Occupied {
	// 	log.Error("Write to occupied buffer position or to big count")
	// }

	s := chanBuffer[id][count].Send
	if s != nil {
		if graph != nil {
			graph.AddEdge(s, ch)
		} else {
			AddEdge(s, ch, false)
		}
	}

	if !flags.IgnoreFifo {
		r := baseA.MostRecentSend[routine][id]
		if r.Elem != nil {
			if graph != nil {
				graph.AddEdge(r.Elem, ch)
			} else {
				AddEdge(r.Elem, ch, false)
			}
		}
	}

	chanBuffer[id][count] = baseA.BufferedVC{
		Occupied: true,
		Send:     ch,
	}
}

// Recv updates and calculates the vector clocks given a receive on a buffered channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *TraceElementChannel: The trace element
func Recv(graph *PoGraph, ch *trace.ElementChannel) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetObjId()
	routine := ch.GetRoutine()
	qSize := ch.GetQSize()

	newBuffer(id, qSize)

	s := chanBuffer[id][0].Send

	if s != nil {
		if graph != nil {
			graph.AddEdge(s, ch)
		} else {
			AddEdge(s, ch, false)
		}
	}

	if !flags.IgnoreFifo {
		r := baseA.MostRecentReceive[routine][id]
		if r.Elem != nil {
			if graph != nil {
				graph.AddEdge(r.Elem, ch)
			} else {
				AddEdge(r.Elem, ch, false)
			}
		}
	}

	chanBuffer[id] = append(chanBuffer[id][1:], baseA.BufferedVC{
		Occupied: false,
		Send:     nil,
	})
}

// RecvC updates and calculates the vector clocks given a receive on a closed channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *TraceElementChannel: The trace element
//   - buffered bool: true if the channel is buffered
func RecvC(graph *PoGraph, ch *trace.ElementChannel, buffered bool) {
	if ch.GetTPost() == 0 {
		return
	}

	id := ch.GetObjId()

	if _, ok := baseA.CloseData[id]; ok {
		c := baseA.CloseData[id]

		if graph != nil {
			graph.AddEdge(c, ch)
		} else {
			AddEdge(c, ch, false)
		}
	}

}

// Create a new map of buffered vector clocks for a channel if not already in
// baseA.bufferedVCs.
//
// Parameter:
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
func newBuffer(id int, qSize int) {
	if _, ok := chanBuffer[id]; !ok {
		chanBuffer[id] = make([]baseA.BufferedVC, 1)
		chanBufferSize[id] = qSize
		chanBuffer[id][0] = baseA.BufferedVC{
			Occupied: false,
			Send:     nil,
		}
	}
}
