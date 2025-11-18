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
	ABlocking               ResultType = "A07"
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

var ResultTypes = []ResultType{
	ASendOnClosed,
	// ARecvOnClosed,
	ACloseOnClosed,
	ACloseOnNilChannel,
	ANegWG,
	AUnlockOfNotLockedMutex,
	ABlocking,
	AConcurrentRecv,
	ASelCaseWithoutPartner,
	PSendOnClosed,
	PRecvOnClosed,
	PNegWG,
	PUnlockBeforeLock,
	PCyclicDeadlock,
	LUnknown,
	LUnbufferedWith,
	LUnbufferedWithout,
	LBufferedWith,
	LBufferedWithout,
	LNilChan,
	LSelectWith,
	LSelectWithout,
	LMutex,
	LWaitGroup,
	LCond,
	LContext,
}

var ResultTypesActual = []ResultType{
	ASendOnClosed,
	ARecvOnClosed,
	ACloseOnClosed,
	ACloseOnNilChannel,
	ANegWG,
	AUnlockOfNotLockedMutex,
	ABlocking,
	AConcurrentRecv,
	ASelCaseWithoutPartner,
}

var ResultTypesPotential = []ResultType{
	PSendOnClosed,
	// PRecvOnClosed,
	PNegWG,
	PUnlockBeforeLock,
	PCyclicDeadlock,
}

var ResultTypesLeak = []ResultType{
	LUnknown,
	LUnbufferedWith,
	LUnbufferedWithout,
	LBufferedWith,
	LBufferedWithout,
	LNilChan,
	LSelectWith,
	LSelectWithout,
	LMutex,
	LWaitGroup,
	LCond,
	LContext,
}

var ResultTypesRecording = []ResultType{
	RUnknownPanic,
	RTimeout,
}

func ResultTypeFromString(code string) ResultType {
	switch code {
	case "A01":
		return ASendOnClosed
	case "A02":
		return ARecvOnClosed
	case "A03":
		return ACloseOnClosed
	case "A04":
		return ACloseOnNilChannel
	case "A05":
		return ANegWG
	case "A06":
		return AUnlockOfNotLockedMutex
	case "A07":
		return ABlocking
	case "A08":
		return AConcurrentRecv
	case "A09":
		return ASelCaseWithoutPartner
	case "P01":
		return PSendOnClosed
	case "P02":
		return PRecvOnClosed
	case "P03":
		return PNegWG
	case "P04":
		return PUnlockBeforeLock
	case "P05":
		return PCyclicDeadlock
	case "L00":
		return LUnknown
	case "L01":
		return LUnbufferedWith
	case "L02":
		return LUnbufferedWithout
	case "L03":
		return LBufferedWith
	case "L04":
		return LBufferedWithout
	case "L05":
		return LNilChan
	case "L06":
		return LSelectWith
	case "L07":
		return LSelectWithout
	case "L08":
		return LMutex
	case "L09":
		return LWaitGroup
	case "L10":
		return LCond
	case "L11":
		return LContext
	case "R01":
		return RUnknownPanic
	case "R02":
		return RTimeout
	default:
		return Empty // Return Empty for codes not found
	}
}

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

func (rt ResultType) IsLeak() bool {
	return string(rt)[0] == 'L'
}
