// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_exit.go
// Brief: Functionality for the exit codes
//
// Author: Erik Kassubek
// Created: 2025-03-18
//
// License: BSD-3-Clause

package runtime

var gocdrExitCode = 0
var gocdrExitCodePos = ""

// GetExitCode returns the exit code and exit position
//
// Returns:
//   - int: exit code
//   - string: exit position
func GetExitCode() (int, string) {
	return gocdrExitCode, gocdrExitCodePos
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
			gocdrExitCode = ExitCodeSendClose
			skip = 5
		} else if m.Error() == "close of closed channel" {
			gocdrExitCode = ExitCodeCloseClose
			skip = 4
		} else if m.Error() == "close of nil channel" {
			gocdrExitCode = ExitCodeCloseNil
			skip = 4
		}
	case string:
		if m == "sync: negative WaitGroup counter" {
			gocdrExitCode = ExitCodeNegativeWG
			skip = 5
		} else if hasPrefix(m, "test timed out") || hasPrefix(m, "Timeout") {
			gocdrExitCode = ExitCodeTimeout
		} else if m == "sync: unlock of unlocked mutex" {
			gocdrExitCode = ExitCodeUnlockBeforeLock
			skip = 6
		} else if m == "sync: Unlock of unlocked RWMutex" {
			gocdrExitCode = ExitCodeUnlockBeforeLock
			skip = 4
		} else if m == "sync: RUnlock of unlocked RWMutex" {
			gocdrExitCode = ExitCodeUnlockBeforeLock
			skip = 5
		} else if m == "Timeout" {
			gocdrExitCode = ExitCodeTimeout
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
	gocdrExitCodePos = file + posSep + intToString(line)
	if printDebug {
		println("AECP: ", gocdrExitCodePos, " ", gocdrExitCode)
	}

	if gocdrExitCode == 0 {
		gocdrExitCode = ExitCodePanic
	}
}
