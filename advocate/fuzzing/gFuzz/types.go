// Copyright (c) 2025 Erik Kassubek
//
// File: types.go
// Brief: Types for gFuzz
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package gFuzz

// For each pair of channel operations, that have communicated, store the following information:
//
//   - sendID: file:line:caseSend of the send
//   - caseSend: If the send is in a select, the case ID, otherwise 0
//   - recvID: file:line:Recv of the recv
//   - caseRecv: If the recv is in a select, the case ID, otherwise 0
//   - chanID: local ID of the channel
//   - sendSel: id of the select case, if not part of select: -2
//   - recvSel: id of the select case, if not part of select: -2
//   - com: number of communication in this run of avg of communications over all runs
type FuzzingPair struct {
	SendID  string
	RecvID  string
	ChanID  int
	SendSel int
	RecvSel int
	Com     float64
}

// For each channel that has ever been created, store the
// following information:
//
//   - globalId: file:line of creation with new
//   - localId: id in this run
//   - qSize: buffer size of the channel
//   - maxQSize: maximum buffer fullness over all runs
//   - whether the channel has always/never/sometimes been closed
type FuzzingChannel struct {
	GlobalID  string
	LocalID   int
	CloseInfo closeInfo
	QSize     int
	MaxQCount int
}

// Info for a channel wether it was closed in all runs,
// never closed or in some runs closed and in others not
type closeInfo string

const (
	Always    closeInfo = "a"
	Never     closeInfo = "n"
	Sometimes closeInfo = "s"
)
