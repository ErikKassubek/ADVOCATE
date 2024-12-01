// Copyrigth (c) 2024 Erik Kassubek
//
// File: bugTypes.go
// Brief: Print informations for all bug types
//
// Author: Erik Kassubek
// Created: 2024-06-14
//
// License: BSD-3-Clause

package explanation

import (
	"fmt"
)

// type (bug / diagnostics)
var bugCrit = map[string]string{
	"A01": "Bug",
	"A02": "Diagnostics",
	"A03": "Bug",
	"A04": "Diagnostics",
	"A05": "Diagnostics",
	"P01": "Bug",
	"P02": "Diagnostic",
	"P03": "Bug",
	"P04": "Bug",
	"L00": "Leak",
	"L01": "Leak",
	"L02": "Leak",
	"L03": "Leak",
	"L04": "Leak",
	"L05": "Leak",
	"L06": "Leak",
	"L07": "Leak",
	"L08": "Leak",
	"L09": "Leak",
	"L10": "Leak",
}

var bugNames = map[string]string{
	"A01": "Actual Send on Closed Channel",
	"A02": "Actual Receive on Closed Channel",
	"A03": "Actual Close on Closed Channel",
	"A04": "Concurrent Receive",
	"A05": "Select Case without Partner",

	"P01": "Possible Send on Closed Channel",
	"P02": "Possible Receive on Closed Channel",
	"P03": "Possible Negative WaitGroup cCounter",
	"P04": "Possible unlock of not locked mutex",

	"L00": "Leak on routine without blocking operation",
	"L01": "Leak of unbuffered Channel with possible partner",
	"L02": "Leak on unbuffered Channel without possible partner",
	"L03": "Leak of buffered Channel with possible partner",
	"L04": "Leak on buffered Channel without possible partner",
	"L05": "Leak on nil channel",
	"L06": "Leak of select with possible partner",
	"L07": "Leak on select without possible partner",
	"L08": "Leak on sync.Mutex",
	"L09": "Leak on sync.WaitGroup",
	"L10": "Leak on sync.Cond",
}

var bugCodes = make(map[string]string) // inverse of bugNames, initialized in init

// explanations
var bugExplanations = map[string]string{
	"A01": "During the execution of the program, a send on a closed channel occurred.\n" +
		"The occurrence of a send on closed leads to a panic.",
	"A02": "During the execution of the program, a receive on a closed channel occurred.\n",
	"A03": "During the execution of the program, a close on a close channel occurred.\n" +
		"The occurrence of a close on a closed channel leads to a panic.",
	"A04": "During the execution of the program, a channel waited to receive at multiple positions at the same time.\n" +
		"In this case, the actual receiver of a send message is chosen randomly.\n" +
		"This can lead to nondeterministic behavior.",
	"A05": "During the execution of the program, a select was executed, where, based " +
		"on the happens-before relation, at least one case could never be triggered.\n" +
		"This can be a desired behavior, especially considering, that only executed " +
		"operations are considered, but it can also be an hint of an unnecessary select case.",
	"P01": "The analyzer detected a possible send on a closed channel.\n" +
		"Although the send on a closed channel did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation.\n" +
		"Such a send on a closed channel leads to a panic.",
	"P02": "The analyzer detected a possible receive on a closed channel.\n" +
		"Although the receive on a closed channel did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation." +
		"This is not necessarily a bug, but it can be an indication of a bug.",
	"P03": "The analyzer detected a possible negative WaitGroup counter.\n" +
		"Although the negative counter did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation.\n" +
		"A negative counter will lead to a panic.",
	"P04": "The analyzer detected a possible unlock on a not locked mutex.\n" +
		"Although the unlock of a not locked mutex did not occur during the recording, " +
		"it is possible that it will occur, based on the happens before relation.\n" +
		"A unlock of a not locked mutex will result in a panic.",
	"L00": "The analyzer detected a leak on a routine without a blocking operations.\n" +
		"This means that the routine was terminated because of a panic in another routine " +
		"or because the main routine terminated while this routine was still running.\n" +
		"This can be a desired behavior, but it can also be a signal for a not otherwise detected block.",
	"L01": "The analyzer detected a leak of an unbuffered channel with a possible partner.\n" +
		"A leak of an unbuffered channel is a situation, where a unbuffered channel is " +
		"still blocking at the end of the program.\n" +
		"The partner is a corresponding send or receive operation, which communicated with another operation, " +
		"but could communicated with the stuck operation instead, resolving the deadlock.",
	"L02": "The analyzer detected a leak of an unbuffered channel without a possible partner.\n" +
		"A leak of an unbuffered channel is a situation, where a unbuffered channel is " +
		"still blocking at the end of the program.\n" +
		"The analyzer could not find a partner for the stuck operation, which would resolve the leak.",
	"L03": "The analyzer detected a leak of a buffered channel with a possible partner.\n" +
		"A leak of a buffered channel is a situation, where a buffered channel is " +
		"still blocking at the end of the program.\n" +
		"The partner is a corresponding send or receive operation, which communicated with another operation, " +
		"but could communicated with the stuck operation instead, resolving the leak.",
	"L04": "The analyzer detected a leak of a buffered channel without a possible partner.\n" +
		"A leak of a buffered channel is a situation, where a buffered channel is " +
		"still blocking at the end of the program.\n" +
		"The analyzer could not find a partner for the stuck operation, which would resolve the leak.",
	"L05": "The analyzer detected a leak on a nil channel.\n" +
		"A leak on a nil channel is a situation, where a nil channel is still blocking at the end of the program.\n" +
		"A nil channel is a channel, which was never initialized or set to nil." +
		"An operation on a nil channel will block indefinitely.",
	"L06": "The analyzer detected a leak of a select with a possible partner.\n" +
		"A leak of a select is a situation, where a select is still blocking at the end of the program.\n" +
		"The partner is a corresponding send or receive operation, which communicated with another operation, " +
		"but could communicated with the stuck operation instead, resolving the leak.",
	"L07": "The analyzer detected a leak of a select without a possible partner.\n" +
		"A leak of a select is a situation, where a select is still blocking at the end of the program.\n" +
		"The analyzer could not find a partner for the stuck operation, which would resolve the leak.",
	"L08": "The analyzer detected a leak on a sync.Mutex.\n" +
		"A leak on a sync.Mutex is a situation, where a sync.Mutex lock operations is still blocking at the end of the program.\n" +
		"A sync.Mutex lock operation is a operation, which is blocking, because the lock is already acquired.",
	"L09": "The analyzer detected a leak on a sync.WaitGroup.\n" +
		"A leak on a sync.WaitGroup is a situation, where a sync.WaitGroup is still blocking at the end of the program.\n" +
		"A sync.WaitGroup wait is blocking, because the counter is not zero.",
	"L10": "The analyzer detected a leak on a sync.Cond.\n" +
		"A leak on a sync.Cond is a situation, where a sync.Cond wait is still blocking at the end of the program.\n" +
		"A sync.Cond wait is blocking, because the condition is not met.",
}

// examples
var bugExamples map[string]string = map[string]string{
	"A01": "func main() {\n" +
		"    c := make(chan int)\n" +
		"    close(c)          // <-------\n" +
		"    c <- 1            // <-------\n}",
	"A02": "func main() {\n" +
		"    c := make(chan int)\n" +
		"    close(c)          // <-------\n" +
		"    <-c               // <-------\n}",
	"A03": "func main() {\n" +
		"    c := make(chan int)\n" +
		"    close(c)          // <-------\n" +
		"    close(c)          // <-------\n}",
	"A04": "func main() {\n" +
		"    c := make(chan int, 1)\n\n" +
		"    go func() {\n" +
		"        <-c             // <-------\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        <-c             // <-------\n" +
		"    }()\n\n" +
		"    c <- 1\n" +
		"}",
	"A05": "func main() {\n" +
		"    c := make(chan int)\n" +
		"    d := make(chan int)\n" +
		"    go func() {\n" +
		"        <-c\n" +
		"    }()\n\n" +
		"    select{\n" +
		"    case c1 := <- c:\n" +
		"        print(c1)\n" +
		"    case d <- 1:      // <-------\n" +
		"        print(\"d\")\n" +
		"    }\n",
	"P01": "func main() {\n" +
		"    c := make(chan int)\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <-------\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        <- c\n" +
		"    }()\n\n" +
		"    close(c)            // <-------\n}",
	"P02": "func main() {\n" +
		"    c := make(chan int)\n\n" +
		"    go func() {\n" +
		"        c <- 1\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        <- c            // <-------\n" +
		"    }()\n\n" +
		"    close(c)            // <-------\n}",
	"P03": "func main() {\n" +
		"    var wg sync.WaitGroup\n\n" +
		"    go func() {\n" +
		"        wg.Add(1)       // <-------\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        wg.Done()       // <-------\n" +
		"    }()\n\n" +
		"    wg.Wait()\n}",
	"P04": "func main() {\n" +
		"    var m sync.Mutex\n\n" +
		"    go func() {\n" +
		"        m.Lock()       // <-------\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        m.Unlock()     // <-------\n" +
		"    }()\n\n}",
	"L00": "func main() {\n" +
		"    go func() {\n" +
		"        time.Sleep(time.Second)          // <------- Is still running when main routine terminates\n" +
		"    }()\n\n" +
		"}",
	"L01": "func main() {\n" +
		"    c := make(chan int)\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Communicates\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        <- c            // <------- Communicates, possible partner\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Leak\n" +
		"    }()\n" +
		"}",
	"L02": "func main() {\n" +
		"    c := make(chan int)\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Leak, no possible partner\n" +
		"    }()\n" +
		"}",
	"L03": "func main() {\n" +
		"    c := make(chan int, 1)\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Communicates\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        <- c            // <------- Communicates, possible partner\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Leak\n" +
		"    }()\n" +
		"}",
	"L04": "func main() {\n" +
		"    c := make(chan int, 1)\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Leak, no possible partner\n" +
		"    }()\n" +
		"}",
	"L05": "func main() {\n" +
		"    var c chan int      // <------- Not initialized -> c = nil\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Leak\n" +
		"    }()\n",
	"L06": "func main() {\n" +
		"    c := make(chan int)\n\n" +
		"    go func() {\n" +
		"        c <- 1          // <------- Communicates\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        <- c            // <------- Communicates, possible partner\n" +
		"    }()\n\n" +
		"    go func() {\n" +
		"        select {        // <------- Leak\n" +
		"        case c <- 1:    // <------- Possible partner\n" +
		"        }\n" +
		"    }()\n" +
		"}",
	"L07": "func main() {\n" +
		"    c := make(chan int)\n\n" +
		"    go func() {\n" +
		"        select {        // <------- Leak, no possible partner\n" +
		"        case c <- 1:\n" +
		"        }\n" +
		"    }()\n" +
		"}",
	"L08": "func main() {\n" +
		"    var m sync.Mutex\n\n" +
		"    go func() {\n" +
		"        m.Lock()        // <------- Leak\n" +
		"    }()\n\n" +
		"    m.Lock()            // <------- Lock, no unlock\n" +
		"}",
	"L09": "func main() {\n" +
		"    var wg sync.WaitGroup\n\n" +
		"    wg.Add(1)           // <------- Add, no Done\n" +
		"    wg.Wait()           // <------- Leak\n" +
		"}",
	"L10": "func main() {\n" +
		"    var c sync.Cond\n\n" +
		"    c.Wait()            // <------- Leak, no signal/broadcast\n" +
		"}",
}

var rewriteType = map[string]string{
	"A01": "Actual",
	"A02": "Actual",
	"A03": "Actual",
	"A04": "Actual",
	"A05": "Actual",
	"P01": "Possible",
	"P02": "Possible",
	"P03": "Possible",
	"P04": "Possible",
	"L01": "LeakPos",
	"L02": "Leak",
	"L03": "LeakPos",
	"L04": "Leak",
	"L05": "Leak",
	"L06": "LeakPos",
	"L07": "Leak",
	"L08": "LeakPos",
	"L09": "LeakPos",
	"L10": "LeakPos",
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
		"    - The program execution depends on outside input, that was not exactly reproduced",
	"11": "The replay got stuck during the execution.\n" +
		"A waiting trace element was not executed for a long time.\n" +
		"This can be caused by a stuck replay.\n" +
		"Possible causes are:\n" +
		"    - The program was altered between recording and replay\n" +
		"    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n" +
		"    - The program execution path depends on the order of not tracked operations\n" +
		"    - The program execution depends on outside input, that was not exactly reproduced",
	"12": "The replay got stuck during the execution.\n" +
		"No trace element was executed for a long tim.\n" +
		"This can be caused by a stuck replay.\n" +
		"Possible causes are:\n" +
		"    - The program was altered between recording and replay\n" +
		"    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n" +
		"    - The program execution path depends on the order of not tracked operations\n" +
		"    - The program execution depends on outside input, that was not exactly reproduced",
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
	// "41": "cyclic",
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
	"GF": "Routine: Fork",
	"GE": "Routine",
}

func init() {
	for key, value := range bugNames {
		bugCodes[value] = key
	}
}

/*
 * Get the code key from the description
 * Args:
 *     description (string): bug description
 * Returns:
 *     string: code if exists, otherwise empty string
 */
func GetCodeFromDescription(description string) string {
	if value, ok := bugCodes[description]; ok {
		return value
	}
	return ""
}

func getBugTypeDescription(bugType string) map[string]string {
	return map[string]string{
		"crit":        bugCrit[bugType],
		"name":        bugNames[bugType],
		"explanation": bugExplanations[bugType],
		"example":     bugExamples[bugType],
	}
}

func printBugTypeDescription(bugType string) {
	fmt.Println(bugCrit[bugType] + ": " + bugNames[bugType] + "\n")
	fmt.Println(bugExplanations[bugType] + "\n")
	fmt.Println(bugExamples[bugType])
}

func getBugElementType(elemType string) string {
	if _, ok := objectTypes[elemType]; !ok {
		return "Unknown element type"
	}
	return objectTypes[elemType]
}
