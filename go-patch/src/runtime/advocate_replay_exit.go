// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_replay_exit.go
// Brief: Replay exit and exit codes
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package runtime

// Set replayForceExit
//
// Parameter:
//
//	force: force exit
// func SetForceExit(force bool) {
// 	replayForceExit = force
// }

// Set the expected exit code
//
// Parameters:
//   - code: the expected exit code
func SetExpectedExitCode(code int) {
	expectedExitCode = code
}

// Exit the program with the given code.
//
// Parameter:
//   - code: the exit code
//   - msg any: message
func ExitReplayWithCode(code int, msg any) {
	if !hasReturnedExitCode {
		// if !isExitCodeConfOnEndElem(code) && !stuckReplayExecutedSuc {
		// 	return
		// }
		println("\nExit Replay with code ", code, ExitCodeNames[code])
		hasReturnedExitCode = true
	} else {
		println("Exit code already returned")
	}

	if msg != "" {
		print("Exit Message: ")

		var p _panic
		p.arg = msg
		preprintpanics(&p)
		printpanics(&p)
		print("\n")
	}

	exit(int32(code))
}

//	For some exit codes, the replay is seen as confirmed, if the replay end
//	element is reached. This function returns wether the exit code is
//	such a code
//	The codes are
//	   20 - 29: Leak
//
// Parameter:
//   - code int: the exit code
//
// Returns:
//   - bool: true if the code is a leak code
func isExitCodeConfOnEndElem(code int) bool {
	return (code >= 20 && code < 30) || (code >= 40 && code < 50)
}

var hasPanicked = false

// Exit the program with the given code if the program panics.
//
// Parameter:
//   - msg: the panic message
func ExitReplayPanic(msg any) {
	if hasPanicked {
		exit(1)
	}
	hasPanicked = true

	SetExitCodeFromPanicMsg(msg)
	if IsAdvocateFuzzingEnabled() {
		finishFuzzingFunc()
	} else if IsTracingEnabled() {
		finishTracingFunc()
	}

	// if !IsReplayEnabled() {
	// 	return
	// }

	ExitReplayWithCode(advocateExitCode, msg)
}

// ExitReplayTimeout exits the program, when a timeout in tracing or replay
// was triggered
func ExitReplayTimeout() {
	// top := ""
	// currentRunning := getCurrentOps()
	// for id, elem := range currentRunning {
	// 	id_str := uint64ToString(id)
	// 	elem_str := elem.getPos()
	// 	top += id_str + "->" + elem_str
	// }

	// println("ExitPosition:" + top)

	ExitReplayPanic("Timeout")
}
