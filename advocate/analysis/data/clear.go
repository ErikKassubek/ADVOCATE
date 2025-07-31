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
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/control"
)

// Clear the data structures used for the analysis
func Clear() {
	ClearTrace()
	ClearData()
	results.Reset()
	control.Reset()
}

// ClearData resets all data structures used in th analysis
func ClearData() {
	CloseData = make(map[int]*trace.ElementChannel)
	LastSendRoutine = make(map[int]map[int]ElemWithVc)
	LastRecvRoutine = make(map[int]map[int]ElemWithVc)
	ForkOperations = make(map[int]*trace.ElementFork)
	LastAnalyzedElementPerRoutine = make(map[int]trace.Element)
	HasSend = make(map[int]bool)
	MostRecentSend = make(map[int]map[int]ElemWithVcVal)
	HasReceived = make(map[int]bool)
	MostRecentReceive = make(map[int]map[int]ElemWithVcVal)
	WGAddData = make(map[int][]trace.Element)
	WgDoneData = make(map[int][]trace.Element)
	AllLocks = make(map[int][]trace.Element)
	AllUnlocks = make(map[int][]trace.Element)
	LockSet = make(map[int]map[int]string)
	MostRecentAcquire = make(map[int]map[int]ElemWithVc)
	MostRecentAcquireTotal = make(map[int]ElemWithVcVal)
	LastAtomicWriter = make(map[int]*trace.ElementAtomic)
	CurrentlyWaiting = make(map[int][]*trace.ElementCond)
	LeakingChannels = make(map[int][]VectorClockTID2)
	SelectCases = make([]AllSelectCase, 0)
	ForkOperations = make(map[int]*trace.ElementFork)
	LastChangeWG = make(map[int]*trace.ElementWait)
	RelR = make(map[int]*ElemWithVc)
	RelW = make(map[int]*ElemWithVc)
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

	OSuc = make(map[int]*trace.ElementOnce)

	HoldSend = make([]HoldObj, 0)
	HoldRecv = make([]HoldObj, 0)

	CurrentState = State{}

	WaitingReceive = make([]*trace.ElementChannel, 0)
	MaxOpID = make(map[int]int)
}
