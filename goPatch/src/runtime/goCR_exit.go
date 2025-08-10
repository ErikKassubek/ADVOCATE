// GOCP-FILE_START

// File: goCR_exit.go
// Brief: Functionality for the exit codes
//
// Created: 2025-03-18
//
// License: BSD-3-Clause

package runtime

var goCRExitCode = 0
var goCRExitCodePos = ""

// GetExitCode returns the exit code and exit position
//
// Returns:
//   - int: exit code
//   - string: exit position
func GetExitCode() (int, string) {
	return goCRExitCode, goCRExitCodePos
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
			goCRExitCode = ExitCodeSendClose
			skip = 5
		} else if m.Error() == "close of closed channel" {
			goCRExitCode = ExitCodeCloseClose
			skip = 4
		} else if m.Error() == "close of nil channel" {
			goCRExitCode = ExitCodeCloseNil
			skip = 4
		}
	case string:
		if m == "sync: negative WaitGroup counter" {
			goCRExitCode = ExitCodeNegativeWG
			skip = 5
		} else if hasPrefix(m, "test timed out") || hasPrefix(m, "Timeout") {
			goCRExitCode = ExitCodeTimeout
		} else if m == "sync: unlock of unlocked mutex" {
			goCRExitCode = ExitCodeUnlockBeforeLock
			skip = 6
		} else if m == "sync: Unlock of unlocked RWMutex" {
			goCRExitCode = ExitCodeUnlockBeforeLock
			skip = 4
		} else if m == "sync: RUnlock of unlocked RWMutex" {
			goCRExitCode = ExitCodeUnlockBeforeLock
			skip = 5
		} else if m == "Timeout" {
			goCRExitCode = ExitCodeTimeout
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
	goCRExitCodePos = file + ":" + intToString(line)
	if printDebug {
		println("AECP: ", goCRExitCodePos, " ", goCRExitCode)
	}

	if goCRExitCode == 0 {
		goCRExitCode = ExitCodePanic
	}
}
