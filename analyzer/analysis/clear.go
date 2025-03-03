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
	"analyzer/clock"
)

/*
 * Clear the data structures used for the analysis
 */
func Clear() {
	ClearTrace()
	ClearData()
}

func ClearTrace() {
	traces = make(map[int][]TraceElement)
	currentIndex = make(map[int]int)
}

func ClearData() {
	closeData = make(map[int]*TraceElementChannel)
	lastSendRoutine = make(map[int]map[int]elemWithVc)
	lastRecvRoutine = make(map[int]map[int]elemWithVc)
	hasSend = make(map[int]bool)
	mostRecentSend = make(map[int]map[int]ElemWithVcVal)
	hasReceived = make(map[int]bool)
	mostRecentReceive = make(map[int]map[int]ElemWithVcVal)
	bufferedVCs = make(map[int][]bufferedVC)
	wgAdd = make(map[int][]TraceElement)
	wgDone = make(map[int][]TraceElement)
	allLocks = make(map[int][]TraceElement)
	allUnlocks = make(map[int][]TraceElement)
	lockSet = make(map[int]map[int]string)
	mostRecentAcquire = make(map[int]map[int]elemWithVc)
	mostRecentAcquireTotal = make(map[int]ElemWithVcVal)
	relW = make(map[int]clock.VectorClock)
	relR = make(map[int]clock.VectorClock)
	leakingChannels = make(map[int][]VectorClockTID2)
	selectCases = make([]allSelectCase, 0)
	allForks = make(map[int]*TraceElementFork)
	exitCode = 0
	exitPos = ""
	fuzzingFlowOnce = make([]ConcurrentEntry, 0)
	fuzzingFlowMutex = make([]ConcurrentEntry, 0)
	fuzzingFlowSend = make([]ConcurrentEntry, 0)
	fuzzingFlowRecv = make([]ConcurrentEntry, 0)
	executedOnce = make(map[int]*ConcurrentEntry)
	fuzzingCounter = make(map[int]map[string]int)

	currentVCHb = make(map[int]clock.VectorClock)
	currentVCWmhb = make(map[int]clock.VectorClock)
	channelWithoutPartner = make(map[int]map[int]*TraceElementChannel)

	numberOfRoutines = 0
	timeoutHappened = false

	currentState = State{}

	wasCanceled.Store(false)
	wasCanceledRam.Store(false)
}
