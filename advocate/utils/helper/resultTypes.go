// Copyright (c) 2025 Erik Kassubek
//
// File: constants.go
// Brief: List of global constants
//
// Author: Erik Kassubek
// Created: 2025-04-14
//
// License: BSD-3-Clause

// ResultType is an ID for a type of result

package helper

// ResultType is an enum for the type of found bug
type ResultType string

// possible values for ResultType
const (
	Empty ResultType = ""

	// actual
	ASendOnClosed           ResultType = "A01"
	ARecvOnClosed           ResultType = "A02"
	ACloseOnClosed          ResultType = "A03"
	ACloseOnNilChannel      ResultType = "A04"
	ANegWG                  ResultType = "A05"
	AUnlockOfNotLockedMutex ResultType = "A06"
	ADeadlock               ResultType = "A07"
	AConcurrentRecv         ResultType = "A08"
	ASelCaseWithoutPartner  ResultType = "A09"

	// possible
	PSendOnClosed     ResultType = "P01"
	PRecvOnClosed     ResultType = "P02"
	PNegWG            ResultType = "P03"
	PUnlockBeforeLock ResultType = "P04"
	PCyclicDeadlock   ResultType = "P05"

	// leaks
	LUnknown           ResultType = "L00"
	LUnbufferedWith    ResultType = "L01"
	LUnbufferedWithout ResultType = "L02"
	LBufferedWith      ResultType = "L03"
	LBufferedWithout   ResultType = "L04"
	LNilChan           ResultType = "L05"
	LSelectWith        ResultType = "L06"
	LSelectWithout     ResultType = "L07"
	LMutex             ResultType = "L08"
	LWaitGroup         ResultType = "L09"
	LCond              ResultType = "L10"
	LContext           ResultType = "L11"

	// recording
	RUnknownPanic ResultType = "R01"
	RTimeout      ResultType = "R02"

	// not executed select
	// SNotExecutedWithPartner = "S00"
)

// Values for the possible program exit codes
const (
	ExitCodeNone             = -1
	ExitCodePanic            = 3
	ExitCodeTimeout          = 10
	ExitCodeLeakUnbuf        = 20
	ExitCodeLeakBuf          = 21
	ExitCodeLeakMutex        = 22
	ExitCodeLeakCond         = 23
	ExitCodeLeakWG           = 24
	ExitCodeSendClose        = 30
	ExitCodeRecvClose        = 31
	ExitCodeCloseClose       = 32
	ExitCodeCloseNil         = 33
	ExitCodeNegativeWG       = 34
	ExitCodeUnlockBeforeLock = 35
	ExitCodeCyclic           = 41
)

// MinExitCodeSuc is the minimum exit code for successful replay
const MinExitCodeSuc = ExitCodeLeakUnbuf
