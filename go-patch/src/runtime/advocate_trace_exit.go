// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace.go
// Brief: Functionality for the exit codes
//
// Author: Erik Kassubek
// Created: 2025-03-18
//
// License: BSD-3-Clause

package runtime

var advocatePanicWriteBlock chan struct{}
var advocatePanicDone chan struct{}

var advocateExitCode = 0
var advocateExitCodePos = ""

const (
	exitCodeSendOnClosed          = 1
	exitCodeCloseOnClosed         = 2
	exitCodeCloseOnNilChannel     = 3
	exitCodeNegWG                 = 4
	exitCodeUnlockOfUnlockedMutex = 5
	exitCodeUnknownPanic          = 6
	exitCodeTimeOut               = 7
)


/*
 * Get the channels used to write the trace on certain panics
 * Args:
 *    apwb (chan struct{}): advocatePanicWriteBlock
 *    apd (chan struct{}): advocatePanicDone
 */
 func GetAdvocatePanicChannels(apwb, apd chan struct{}) {
	advocatePanicWriteBlock = apwb
	advocatePanicDone = apd
}

func GetExitCode() (int, string) {
	return advocateExitCode, advocateExitCodePos
}

func SetExitCodeFromPanicString(msg any) {
	_, file, line, _ := Caller(2)
	advocateExitCodePos = file + ":" + intToString(line)

	switch m := msg.(type) {
	case plainError:
		if m.Error() == "send on closed channel" {
			advocateExitCode = exitCodeSendOnClosed
		} else if m.Error() == "close of closed channel" {
			advocateExitCode = exitCodeCloseOnClosed
		} else if m.Error() == "close of nil channel" {
			advocateExitCode = exitCodeCloseOnNilChannel
		}
	case string:
		if m == "sync: negative WaitGroup counter" {
			advocateExitCode = exitCodeNegWG
		} else if hasPrefix(m, "test timed out") {
			advocateExitCode = exitCodeTimeOut
		} else if expectedExitCode == ExitCodeUnlockBeforeLock {
			if m == "sync: RUnlock of unlocked RWMutex" ||
				m == "sync: Unlock of unlocked RWMutex" ||
				m == "sync: unlock of unlocked mutex" {
				advocateExitCode = exitCodeUnlockOfUnlockedMutex
			}
		}
	}

	if advocateExitCode == 0 {
		advocateExitCode = exitCodeUnknownPanic
	}
}