// OOSC-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: oosc_exit.go
// Brief: Functionality for the exit codes
//
// Author: Erik Kassubek
// Created: 2025-03-18
//
// License: BSD-3-Clause

package runtime

var ooscExitCode = 0
var ooscExitCodePos = ""

// GetExitCode returns the exit code and exit position
//
// Returns:
//   - int: exit code
//   - string: exit position
func GetExitCode() (int, string) {
	return ooscExitCode, ooscExitCodePos
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
			ooscExitCode = ExitCodeSendClose
			skip = 5
		} else if m.Error() == "close of closed channel" {
			ooscExitCode = ExitCodeCloseClose
			skip = 4
		} else if m.Error() == "close of nil channel" {
			ooscExitCode = ExitCodeCloseNil
			skip = 4
		}
	case string:
		if m == "sync: negative WaitGroup counter" {
			ooscExitCode = ExitCodeNegativeWG
			skip = 5
		} else if hasPrefix(m, "test timed out") || hasPrefix(m, "Timeout") {
			ooscExitCode = ExitCodeTimeout
		} else if m == "sync: unlock of unlocked mutex" {
			ooscExitCode = ExitCodeUnlockBeforeLock
			skip = 6
		} else if m == "sync: Unlock of unlocked RWMutex" {
			ooscExitCode = ExitCodeUnlockBeforeLock
			skip = 4
		} else if m == "sync: RUnlock of unlocked RWMutex" {
			ooscExitCode = ExitCodeUnlockBeforeLock
			skip = 5
		} else if m == "Timeout" {
			ooscExitCode = ExitCodeTimeout
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
	ooscExitCodePos = file + posSep + intToString(line)
	if printDebug {
		println("AECP: ", ooscExitCodePos, " ", ooscExitCode)
	}

	if ooscExitCode == 0 {
		ooscExitCode = ExitCodePanic
	}
}
