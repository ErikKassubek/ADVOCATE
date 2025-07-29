// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_exit.go
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

// Get the channels used to write the trace on certain panics
//
// Parameters:
//   - apwb (chan struct{}): advocatePanicWriteBlock
//   - apd (chan struct{}): advocatePanicDone
func GetAdvocatePanicChannels(apwb, apd chan struct{}) {
	advocatePanicWriteBlock = apwb
	advocatePanicDone = apd
}

// GetExitCode returns the exit code and exit position
//
// Returns:
//   - int: exit code
//   - string: exit position
func GetExitCode() (int, string) {
	return advocateExitCode, advocateExitCodePos
}

// SetExitCodeFromPanicMsg sets the panic info from the panic message
//
// Parameter:
//   - msg any: the panic message
func SetExitCodeFromPanicMsg(msg any) {
	skip := 3

	switch m := msg.(type) {
	case plainError:
		if m.Error() == "send on closed channel" {
			advocateExitCode = ExitCodeSendClose
			skip = 5
		} else if m.Error() == "close of closed channel" {
			advocateExitCode = ExitCodeCloseClose
			skip = 4
		} else if m.Error() == "close of nil channel" {
			advocateExitCode = ExitCodeCloseNil
			skip = 4
		}
	case string:
		if m == "sync: negative WaitGroup counter" {
			advocateExitCode = ExitCodeNegativeWG
			skip = 5
		} else if hasPrefix(m, "test timed out") || hasPrefix(m, "Timeout") {
			advocateExitCode = ExitCodeTimeout
		} else if m == "sync: unlock of unlocked mutex" {
			advocateExitCode = ExitCodeUnlockBeforeLock
			skip = 6
		} else if m == "sync: Unlock of unlocked RWMutex" {
			advocateExitCode = ExitCodeUnlockBeforeLock
			skip = 4
		} else if m == "sync: RUnlock of unlocked RWMutex" {
			advocateExitCode = ExitCodeUnlockBeforeLock
			skip = 5
		} else if m == "Timeout" {
			advocateExitCode = ExitCodeTimeout
			skip = 0
		}
	default:
		var p _panic
		p.arg = msg
		preprintpanics(&p)
		printpanics(&p)
		print("\n")
		printAllGoroutines()
	}

	_, file, line, _ := Caller(skip)
	advocateExitCodePos = file + ":" + intToString(line)
	if printDebug {
		println("AECP: ", advocateExitCodePos)
	}

	if advocateExitCode == 0 {
		advocateExitCode = ExitCodePanic
	}
}
