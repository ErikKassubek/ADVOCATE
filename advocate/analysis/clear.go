// Copyright (c) 2025 Erik Kassubek
//
// File: clear.go
// Brief: Clear trace and data
//
// Author: Erik Kassubek
// Created: 2025-03-03
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/clock"
	"advocate/memory"
	"advocate/results"
	"advocate/trace"
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
	closeData = make(map[int]*trace.TraceElementChannel)
	lastSendRoutine = make(map[int]map[int]elemWithVc)
	lastRecvRoutine = make(map[int]map[int]elemWithVc)
	hasSend = make(map[int]bool)
	mostRecentSend = make(map[int]map[int]ElemWithVcVal)
	hasReceived = make(map[int]bool)
	mostRecentReceive = make(map[int]map[int]ElemWithVcVal)
	bufferedVCs = make(map[int][]bufferedVC)
	wgAdd = make(map[int][]trace.TraceElement)
	wgDone = make(map[int][]trace.TraceElement)
	allLocks = make(map[int][]trace.TraceElement)
	allUnlocks = make(map[int][]trace.TraceElement)
	lockSet = make(map[int]map[int]string)
	mostRecentAcquire = make(map[int]map[int]elemWithVc)
	mostRecentAcquireTotal = make(map[int]ElemWithVcVal)
	relW = make(map[int]*clock.VectorClock)
	relR = make(map[int]*clock.VectorClock)
	lw = make(map[int]*clock.VectorClock)
	currentlyWaiting = make(map[int][]int)
	leakingChannels = make(map[int][]VectorClockTID2)
	selectCases = make([]allSelectCase, 0)
	allForks = make(map[int]*trace.TraceElementFork)
	exitCode = 0
	exitPos = ""
	replayTimeoutOldest = 0
	replayTimeoutDisabled = 0
	replayTimeoutAck = 0
	fuzzingFlowOnce = make([]ConcurrentEntry, 0)
	fuzzingFlowMutex = make([]ConcurrentEntry, 0)
	fuzzingFlowSend = make([]ConcurrentEntry, 0)
	fuzzingFlowRecv = make([]ConcurrentEntry, 0)
	executedOnce = make(map[int]*ConcurrentEntry)
	fuzzingCounter = make(map[int]map[string]int)

	currentVC = make(map[int]*clock.VectorClock)
	currentWVC = make(map[int]*clock.VectorClock)

	oSuc = make(map[int]*clock.VectorClock)

	holdSend = make([]holdObj, 0)
	holdRecv = make([]holdObj, 0)

	currentState = State{}

	numberSelectCasesWithPartner = 0

	waitingReceive = make([]*trace.TraceElementChannel, 0)
	maxOpID = make(map[int]int)
}
