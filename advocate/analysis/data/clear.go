// Copyright (c) 2025 Erik Kassubek
//
// File: clear.go
// Brief: Clear trace and data
//
// Author: Erik Kassubek
// Created: 2025-03-03
//
// License: BSD-3-Clause

package data

import (
	"advocate/analysis/clock"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/memory"
)

// Clear the data structures used for the analysis
func Clear() {
	ClearTrace()
	ClearData()
	results.Reset()
	memory.Reset()
}

// ClearData resets all data structures used in th analysis
func ClearData() {
	CloseData = make(map[int]*trace.TraceElementChannel)
	LastSendRoutine = make(map[int]map[int]ElemWithVc)
	LastRecvRoutine = make(map[int]map[int]ElemWithVc)
	ForkOperations = make(map[int]*trace.TraceElementFork)
	LastAnalyzedElementPerRoutine = make(map[int]trace.TraceElement)
	HasSend = make(map[int]bool)
	MostRecentSend = make(map[int]map[int]ElemWithVcVal)
	HasReceived = make(map[int]bool)
	MostRecentReceive = make(map[int]map[int]ElemWithVcVal)
	BufferedVCs = make(map[int][]BufferedVC)
	WgAdd = make(map[int][]trace.TraceElement)
	WgDone = make(map[int][]trace.TraceElement)
	AllLocks = make(map[int][]trace.TraceElement)
	AllUnlocks = make(map[int][]trace.TraceElement)
	LockSet = make(map[int]map[int]string)
	MostRecentAcquire = make(map[int]map[int]ElemWithVc)
	MostRecentAcquireTotal = make(map[int]ElemWithVcVal)
	RelW = make(map[int]*clock.VectorClock)
	RelR = make(map[int]*clock.VectorClock)
	Lw = make(map[int]*clock.VectorClock)
	CurrentlyWaiting = make(map[int][]int)
	LeakingChannels = make(map[int][]VectorClockTID2)
	SelectCases = make([]AllSelectCase, 0)
	ForkOperations = make(map[int]*trace.TraceElementFork)
	ExitCode = 0
	ExitPos = ""
	replayTimeoutOldest = 0
	replayTimeoutDisabled = 0
	replayTimeoutAck = 0
	FuzzingFlowOnce = make([]ConcurrentEntry, 0)
	FuzzingFlowMutex = make([]ConcurrentEntry, 0)
	FuzzingFlowSend = make([]ConcurrentEntry, 0)
	FuzzingFlowRecv = make([]ConcurrentEntry, 0)
	ExecutedOnce = make(map[int]*ConcurrentEntry)
	FuzzingCounter = make(map[int]map[string]int)

	CurrentVC = make(map[int]*clock.VectorClock)
	CurrentWVC = make(map[int]*clock.VectorClock)

	OSuc = make(map[int]*clock.VectorClock)

	HoldSend = make([]HoldObj, 0)
	HoldRecv = make([]HoldObj, 0)

	CurrentState = State{}

	WaitingReceive = make([]*trace.TraceElementChannel, 0)
	MaxOpID = make(map[int]int)
}
