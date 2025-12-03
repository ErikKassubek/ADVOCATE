// Copyright (c) 2024 Erik Kassubek
//
// File: bugTypes.go
// Brief: Print informations for all bug types
//
// Author: Erik Kassubek
// Created: 2024-06-14
//
// License: BSD-3-Clause

package explanation

import "advocate/utils/helper"

// type (bug / diagnostics)
var bugCrit = map[helper.ResultType]string{
	helper.RUnknownPanic:           "Bug",
	helper.RTimeout:                "Bug",
	helper.ASendOnClosed:           "Bug",
	helper.ARecvOnClosed:           "Diagnostics",
	helper.ACloseOnClosed:          "Bug",
	helper.ACloseOnNilChannel:      "Bug",
	helper.ANegWG:                  "Bug",
	helper.AUnlockOfNotLockedMutex: "Bug",
	helper.ALeak:                   "Bug",
	helper.ADeadlock:               "Bug",
	helper.AConcurrentRecv:         "Diagnostics",
	helper.PSendOnClosed:           "Bug",
	helper.PRecvOnClosed:           "Diagnostic",
	helper.PNegWG:                  "Bug",
	helper.PUnlockBeforeLock:       "Bug",
	helper.PCyclicDeadlock:         "Bug",
	helper.LUnknown:                "Leak",
	helper.LUnbufferedWith:         "Leak",
	helper.LUnbufferedWithout:      "Leak",
	helper.LBufferedWith:           "Leak",
	helper.LBufferedWithout:        "Leak",
	helper.LNilChan:                "Leak",
	helper.LSelectWith:             "Leak",
	helper.LSelectWithout:          "Leak",
	helper.LMutex:                  "Leak",
	helper.LWaitGroup:              "Leak",
	helper.LCond:                   "Leak",
	helper.LContext:                "Leak",
}

var bugNames = map[helper.ResultType]string{
	helper.ASendOnClosed:           "Actual Send on Closed Channel",
	helper.ARecvOnClosed:           "Actual Receive on Closed Channel",
	helper.ACloseOnClosed:          "Actual Close on Closed Channel",
	helper.ACloseOnNilChannel:      "Actual close on nil channel",
	helper.ANegWG:                  "Actual negative Wait Group",
	helper.AUnlockOfNotLockedMutex: "Actual unlock of not locked mutex",
	helper.ALeak:                   "Actual Leak",
	helper.ADeadlock:               "Actual Deadlock",
	helper.AConcurrentRecv:         "Concurrent Receive",

	helper.PSendOnClosed:     "Possible Send on Closed Channel",
	helper.PRecvOnClosed:     "Possible Receive on Closed Channel",
	helper.PNegWG:            "Possible Negative WaitGroup cCounter",
	helper.PUnlockBeforeLock: "Possible unlock of not locked mutex",
	helper.PCyclicDeadlock:   "Possible cyclic deadlock",

	helper.LUnknown:           "Leak",
	helper.LUnbufferedWith:    "Leak on unbuffered channel with possible partner",
	helper.LUnbufferedWithout: "Leak on unbuffered channel without possible partner",
	helper.LBufferedWith:      "Leak on buffered Channel with possible partner",
	helper.LBufferedWithout:   "Leak on buffered Channel without possible partner",
	helper.LNilChan:           "Leak on nil channel",
	helper.LSelectWith:        "Leak on select with possible partner",
	helper.LSelectWithout:     "Leak on select without possible partner",
	helper.LMutex:             "Leak on sync.Mutex",
	helper.LWaitGroup:         "Leak on sync.WaitGroup",
	helper.LCond:              "Leak on sync.Cond",
	helper.LContext:           "Leak on channel or select on context",

	helper.RUnknownPanic: "Unknown Panic",
	helper.RTimeout:      "Timeout",
}

var bugCodes = make(map[string]helper.ResultType) // inverse of bugNames, initialized in init

// explanations
var bugExplanations = map[helper.ResultType]string{
	helper.ASendOnClosed: "During the execution of the program, a send on a closed channel occurred.\n" +
		"The occurrence of a send on closed leads to a panic.",
	helper.ARecvOnClosed: "During the execution of the program, a receive on a closed channel occurred.\n",
	helper.ACloseOnClosed: "During the execution of the program, a close on a close channel occurred.\n" +
		"The occurrence of a close on a closed channel lead to a panic.",
	helper.ACloseOnNilChannel: "During the execution of the program, a close on a nil channel occurred.\n" +
		"The occurrence of a close on a nil channel lead to a panic.",
	helper.ANegWG: "During the execution, a negative waitgroup counter occurred.\n" +
		"The occurrence of a negative wait group counter lead to a panic.",
	helper.AUnlockOfNotLockedMutex: "During the execution, a not locked mutex was unlocked.\n" +
		"The occurrence of this lead to a panic.",
	helper.ALeak: "During the execution, a blocking bug was detected.\n" +
		"This means, there is a routine that is blocked, and there is not possibility of it being unblocked in the future",
	helper.ADeadlock: "During the execution, a deadlock was detected.\n" +
		"This means, there is are routine that are cyclicaly blocked, and there is not possibility of it being unblocked in the future",
	helper.AConcurrentRecv: "During the execution of the program, a channel waited to receive at multiple positions at the same time.\n" +
		"In this case, the actual receiver of a send message is chosen randomly.\n" +
		"This can lead to nondeterministic behavior.",
	helper.RUnknownPanic: "During the execution of the program, a unknown panic occurred",
	helper.RTimeout:      "The execution of the program timed out",
	helper.PSendOnClosed: "The analyzer detected a possible send on a closed channel.\n" +
		"Although the send on a closed channel did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation.\n" +
		"Such a send on a closed channel leads to a panic.",
	helper.PRecvOnClosed: "The analyzer detected a possible receive on a closed channel.\n" +
		"Although the receive on a closed channel did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation." +
		"This is not necessarily a bug, but it can be an indication of a bug.",
	helper.PNegWG: "The analyzer detected a possible negative WaitGroup counter.\n" +
		"Although the negative counter did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation.\n" +
		"A negative counter will lead to a panic.",
	helper.PUnlockBeforeLock: "The analyzer detected a possible unlock on a not locked mutex.\n" +
		"Although the unlock of a not locked mutex did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation.\n" +
		"A unlock of a not locked mutex will result in a panic.",
	helper.PCyclicDeadlock: "The analysis detected a possible cyclic deadlock.\n" +
		"If this deadlock contains or influences the run of the main routine, this can " +
		"result in the program getting stuck. Otherwise it can lead to an unnecessary use of " +
		"resources.",
	helper.LUnknown: "The analyzer detected a leak.\n" +
		"This means that the routine was terminated because of a panic in another routine " +
		"or because the main routine terminated while this routine was still running.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"This can be a desired behavior, but it can also be a signal for a not otherwise detected block.",
	helper.LUnbufferedWith: "The analyzer detected a Leak on an unbuffered channel with a possible partner.\n" +
		"A Leak on an unbuffered channel is a situation, where a unbuffered channel is " +
		"still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"The partner is a corresponding send or receive operation, which communicated with another operation, " +
		"but could communicated with the stuck operation instead, resolving the deadlock.",
	helper.LUnbufferedWithout: "The analyzer detected a Leak on an unbuffered channel without a possible partner.\n" +
		"A Leak on an unbuffered channel is a situation, where a unbuffered channel is " +
		"still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"The analyzer could not find a partner for the stuck operation, which would resolve the leak.",
	helper.LBufferedWith: "The analyzer detected a Leak on a buffered channel with a possible partner.\n" +
		"A Leak on a buffered channel is a situation, where a buffered channel is " +
		"still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"The partner is a corresponding send or receive operation, which communicated with another operation, " +
		"but could communicated with the stuck operation instead, resolving the leak.",
	helper.LBufferedWithout: "The analyzer detected a Leak on a buffered channel without a possible partner.\n" +
		"A Leak on a buffered channel is a situation, where a buffered channel is " +
		"still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"The analyzer could not find a partner for the stuck operation, which would resolve the leak.",
	helper.LNilChan: "The analyzer detected a leak on a nil channel.\n" +
		"A leak on a nil channel is a situation, where a nil channel is still blocking at the end of the program.\n" +
		"A nil channel is a channel, which was never initialized or set to nil." +
		"An operation on a nil channel will block indefinitely.",
	helper.LSelectWith: "The analyzer detected a Leak on a select with a possible partner.\n" +
		"A Leak on a select is a situation, where a select is still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"The partner is a corresponding send or receive operation, which communicated with another operation, " +
		"but could communicated with the stuck operation instead, resolving the leak.",
	helper.LSelectWithout: "The analyzer detected a Leak on a select without a possible partner.\n" +
		"A Leak on a select is a situation, where a select is still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"The analyzer could not find a partner for the stuck operation, which would resolve the leak.",
	helper.LMutex: "The analyzer detected a leak on a sync.Mutex.\n" +
		"A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.",
	helper.LWaitGroup: "The analyzer detected a leak on a sync.WaitGroup.\n" +
		"A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"A sync.WaitGroup wait is blocking, because the counter is not zero.",
	helper.LCond: "The analyzer detected a leak on a sync.Cond.\n" +
		"A leak on a sync.Cond is a situation, where a sync.Cond wait is still blocking at the end of the program.\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"A sync.Cond wait is blocking, because the condition is not met.",
	helper.LContext: "The analyzer detected a leak on a channel or select on a context\n" +
		"A Leak could potentially resolve itself, if the program would run longer.\n" +
		"This may or may not be a blocking bug.",
}

var exitCodeExplanation = map[string]string{
	"panic": "The replay was started but was terminated unexpectedly.\nThe main reason could be, that the runtime exceeded the timeout of the test",
	"fail": "The analyzer was not able to rewrite the bug.\nThis can be because the bug is an actual bug, " +
		"because the bug is a leak without a possible partner or blocking operations " +
		"or because the analyzer was not able to rewrite the trace for other reasons.",
	"0": "The replay finished without being able to confirm the predicted bug. If the given trace was a directly recorded trace, this is the " +
		"expected behavior. If it was rewritten by the analyzer, this could be an indication " +
		"that something went wrong during rewrite or replay.",
	"3": "During the replay, the program panicked unexpectedly.\n" +
		"This can be expected behavior, e.g. if the program tries to replay a recv on closed " +
		"but the recv on closed is necessarily preceded by a send on closed.",
	"10": "The replay got stuck during the execution.\n" +
		"The main routine has already finished, but the trace still contains not executed operations.\n" +
		"This can be caused by a stuck replay.\n" +
		"Possible causes are:\n" +
		"    - The program was altered between recording and replay\n" +
		"    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n" +
		"    - The program execution path depends on the order of not tracked operations\n" +
		"    - The program execution depends on outside input, that was not exactly reproduced\n" +
		"	 - The program encountered a deadlock earlier in the trace than expected",
	"11": "The replay got stuck during the execution.\n" +
		"A waiting trace element was not executed for a long time.\n" +
		"This can be caused by a stuck replay.\n" +
		"Possible causes are:\n" +
		"    - The program was altered between recording and replay\n" +
		"    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n" +
		"    - The program execution path depends on the order of not tracked operations\n" +
		"    - The program execution depends on outside input, that was not exactly reproduced\n" +
		"	 - The program encountered an unexpected deadlock",
	"12": "The replay got stuck during the execution.\n" +
		"No trace element was executed for a long tim.\n" +
		"This can be caused by a stuck replay.\n" +
		"Possible causes are:\n" +
		"    - The program was altered between recording and replay\n" +
		"    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n" +
		"    - The program execution path depends on the order of not tracked operations\n" +
		"    - The program execution depends on outside input, that was not exactly reproduced" +
		"	 - The program encountered an unexpected deadlock",
	"13": "The replay got stuck during the execution.\n" +
		"The program tried to execute an operation, even though all elements in the trace have already been executed.\n" +
		"This can be caused by a stuck replay.\n" +
		"Possible causes are:\n" +
		"    - The program was altered between recording and replay\n" +
		"    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n" +
		"    - The program execution path depends on the order of not tracked operations\n" +
		"    - The program execution depends on outside input, that was not exactly reproduced",
	"20": "The replay was able to get the leaking unbuffered channel or select unstuck.",
	"21": "The replay was able to get the leaking buffered channel unstuck.",
	"22": "The replay was able to get the leaking mutex unstuck.",
	"23": "The replay was able to get the leaking conditional variable unstuck.",
	"24": "The replay was able to get the leaking wait-group unstuck.",
	"30": "The replay resulted in an expected send on close triggering a panic. The bug was triggered. " +
		"The replay was therefore able to confirm, that the send on closed can actually occur.",
	"31": "The replay resulted in an expected receive on close. The bug was triggered." +
		"The replay was therefore able to confirm, that the receive on closed can actually occur.",
	"32": "The replay resulted in an expected negative wait group triggering a panic. The bug was triggered. " +
		"The replay was therefore able to confirm, that the negative wait group can actually occur.",
	"33": "The replay resulted in an expected lock of an unlocked mutex triggering a panic. The bug was triggered. " +
		"The replay was therefore able to confirm, that the unlock of a not locked mutex can actually occur.",
	"41": "The replay reached the expected point and found stuck mutexes." + "The replay was therefore able to confirm that a deadlock can actually occur.",
}

var objectTypes = map[string]string{
	"AL": "Atomic Load",
	"AS": "Atomic Store",
	"AA": "Atomic Add",
	"AW": "Atomic Swap",
	"AC": "Atomic CompSwap",
	"CS": "Channel: Send",
	"CR": "Channel: Receive",
	"CC": "Channel: Close",
	"ML": "Mutex: Lock",
	"MR": "Mutex: RLock",
	"MT": "Mutex: TryLock",
	"MY": "Mutex: TryRLock",
	"MU": "Mutex: Unlock",
	"MN": "Mutex: RUnlock",
	"WA": "Waitgroup: Add",
	"WD": "Waitgroup: Done",
	"WW": "Waitgroup: Wait",
	"SS": "Select:",
	"DW": "Conditional Variable: Wait",
	"DB": "Conditional Variable: Broadcast",
	"DS": "Conditional Variable: Signal",
	"OE": "Once: Done Executed",
	"ON": "Once: Done Not Executed (because the once was already executed)",
	"RF": "Routine: Fork",
	"RE": "Routine: End",
	"DH": "Mutex: Causing deadlock",
	"DC": "Mutex: Part of deadlock",
	"XX": "Unknown",
}

// adaptExplanationMaps changes the bugNames and bugExplanation maps, if
// the analysis did not search for possible partners
func adaptExplanationMaps() {
	bugNames[helper.LUnbufferedWith] = "Leak on unbuffered channel"
	bugNames[helper.LUnbufferedWithout] = "Leak on unbuffered channel"
	bugNames[helper.LBufferedWith] = "Leak on buffered channel"
	bugNames[helper.LBufferedWithout] = "Leak on buffered channel"
	bugNames[helper.LNilChan] = "Leak on nil channel"
	bugNames[helper.LSelectWith] = "Leak on select"
	bugNames[helper.LSelectWithout] = "Leak on select"

	bugExplanations[helper.LUnbufferedWith] = "The analyzer detected a Leak on an unbuffered channel .\n" +
		"A Leak on an unbuffered channel is a situation, where a unbuffered channel is " +
		"still blocking at the end of the program.\n"
	bugExplanations[helper.LUnbufferedWithout] = "The analyzer detected a Leak on an unbuffered channel.\n" +
		"A Leak on an unbuffered channel is a situation, where a unbuffered channel is " +
		"still blocking at the end of the program.\n"
	bugExplanations[helper.LBufferedWith] = "The analyzer detected a Leak on a buffered channel.\n" +
		"A Leak on a buffered channel is a situation, where a buffered channel is " +
		"still blocking at the end of the program.\n"
	bugExplanations[helper.LBufferedWithout] = "The analyzer detected a Leak on a buffered channel.\n" +
		"A Leak on a buffered channel is a situation, where a buffered channel is " +
		"still blocking at the end of the program.\n"
	bugExplanations[helper.LSelectWith] = "The analyzer detected a Leak on a select.\n" +
		"A Leak on a select is a situation, where a select is still blocking at the end of the program.\n"
	bugExplanations[helper.LSelectWithout] = "The analyzer detected a Leak on a select.\n" +
		"A Leak on a select is a situation, where a select is still blocking at the end of the program.\n"
}

// GetCodeFromDescription returns the code key from the description
//
// Parameter:
//   - description string: bug description
//
// Returns:
//   - helper.ResultType: code if exists, otherwise empty string
func GetCodeFromDescription(description string) helper.ResultType {
	if value, ok := bugCodes[description]; ok {
		return value
	}
	return ""
}

// Get the bug type descriptions from the bug type codes
//
// Parameter:
//   - bugType string: bug type code
//
// Returns:
//   - map[string]string: bug type descriptions
func getBugTypeDescription(bugType helper.ResultType) map[string]string {
	return map[string]string{
		"crit":        bugCrit[bugType],
		"name":        bugNames[bugType],
		"explanation": bugExplanations[bugType],
	}
}

// Get bug element (operation) type from elem type code
//
// Parameter:
//   - elemType string: code for the object (operation)
//
// Returns: string: description of the bug type
func getBugElementType(elemType string) string {
	if _, ok := objectTypes[elemType]; !ok {
		return "Unknown element type"
	}
	return objectTypes[elemType]
}

// buildBugCodes builds the bugCodes as an inversion of bugNames
// If bugCodes is already build, this does nothing
func buildBugCodes() {
	if len(bugCodes) > 0 {
		return
	}

	for key, desc := range bugNames {
		bugCodes[desc] = key
	}
}
