// Copyright (c) 2024 Erik Kassubek
//
// File: analysisConcurrentCommunication.go
// Brief: Find concurrent operations on the same element
//   For concurrent receive: add panic
//   For concurrent send, receive, (try)(r)lock, once.Do: store to use in fuzzing
//
// Author: Erik Kassubek
// Created: 2024-01-27
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/results"
	"analyzer/timer"
	"analyzer/trace"
	"analyzer/utils"
)

// getConcurrentSendForFuzzing checks if for the given send, if there is a
// concurrent send on the same channel. If there is, the information is stored
// in fuzzingFlowSend. This is used for fuzzing.
//
// Parameter:
//   - sender *TraceElementChannel: Send trace element
func getConcurrentSendForFuzzing(sender *trace.TraceElementChannel) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	id := sender.GetID()
	routine := sender.GetRoutine()

	incFuzzingCounter(sender)

	if sender.GetTPost() != 0 {
		return
	}

	for r, elem := range lastSendRoutine {
		if r == routine {
			continue
		}

		if elem[id].vc.GetClock() == nil {
			continue
		}

		happensBefore := clock.GetHappensBefore(elem[id].vc, currentVC[routine])
		if happensBefore == clock.Concurrent {
			elem2 := elem[id].elem
			fuzzingFlowSend = append(fuzzingFlowSend, ConcurrentEntry{Elem: elem2, Counter: getFuzzingCounter(elem2), Type: CERecv})
		}
	}

	if sender.GetTPost() != 0 {
		if _, ok := lastSendRoutine[routine]; !ok {
			lastSendRoutine[routine] = make(map[int]elemWithVc)
		}

		lastSendRoutine[routine][id] = elemWithVc{currentVC[routine].Copy(), sender}
	}
}

// checkForConcurrentRecv checks if for the given recv, if there is a
// concurrent recv on the same channel. If there is, the information is stored
// in fuzzingFlowRecv.
//
// Parameter:
//   - ch *TraceElementChannel: recv trace element
func checkForConcurrentRecv(ch *trace.TraceElementChannel, vc map[int]*clock.VectorClock) {
	if analysisFuzzing {
		timer.Start(timer.FuzzingAna)
		defer timer.Stop(timer.FuzzingAna)
	}
	timer.Start(timer.AnaConcurrent)
	defer timer.Stop(timer.AnaConcurrent)

	id := ch.GetID()
	routine := ch.GetRoutine()

	incFuzzingCounter(ch)

	for r, elem := range lastRecvRoutine {
		if r == routine {
			continue
		}

		if elem[id].vc.GetClock() == nil {
			continue
		}

		happensBefore := clock.GetHappensBefore(elem[id].vc, vc[routine])
		if happensBefore == clock.Concurrent {

			elem2 := elem[id].elem

			if ch.GetTPost() == 0 {
				fuzzingFlowRecv = append(fuzzingFlowRecv, ConcurrentEntry{Elem: elem2, Counter: getFuzzingCounter(elem2), Type: CERecv})
			}

			arg1 := results.TraceElementResult{
				RoutineID: routine,
				ObjID:     id,
				TPre:      ch.GetTPre(),
				ObjType:   "CR",
				File:      ch.GetFile(),
				Line:      ch.GetLine(),
			}

			arg2 := results.TraceElementResult{
				RoutineID: r,
				ObjID:     id,
				TPre:      elem2.GetTPre(),
				ObjType:   "CR",
				File:      elem2.GetFile(),
				Line:      elem2.GetLine(),
			}

			results.Result(results.WARNING, utils.AConcurrentRecv,
				"recv", []results.ResultElem{arg1}, "recv", []results.ResultElem{arg2})
		}
	}

	if ch.GetTPost() != 0 {
		if _, ok := lastRecvRoutine[routine]; !ok {
			lastRecvRoutine[routine] = make(map[int]elemWithVc)
		}

		lastRecvRoutine[routine][id] = elemWithVc{vc[routine].Copy(), ch}
	}
}

// getConcurrentMutexForFuzzing checks if for the given mutex operations, if there is a
// concurrent mutex operations on the same mutex. If there is, the information is stored
// in fuzzingFlowMutex.
//
// Parameter:
//   - mu *TraceElementMutex: mutex operations
func getConcurrentMutexForFuzzing(mu *trace.TraceElementMutex) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	// operation executed normally
	if mu.IsSuc() {
		return
	}

	id := mu.GetID()

	// not executed try lock
	// get currently hold lock because of witch the try lock failed

	if val, ok := currentlyHoldLock[id]; !ok || val == nil {
		utils.LogError("Failed trylock even throw mutex is not locked: ", mu.ToString())
	}

	elem := currentlyHoldLock[id]

	if clock.GetHappensBefore(mu.GetVC(), elem.GetVC()) == clock.Concurrent {
		fuzzingFlowMutex = append(fuzzingFlowMutex, ConcurrentEntry{Elem: elem, Counter: getFuzzingCounter(elem), Type: CEMutex})
	}

}

// getConcurrentOnceForFuzzing checks if for the given once operations, if there is a
// concurrent once operations on the same primitive. If there is, the information is stored
// in fuzzingFlowOnce.
//
// Parameter:
//   - on *TraceElementOnce: once.Do operations
func getConcurrentOnceForFuzzing(on *trace.TraceElementOnce) {
	timer.Start(timer.FuzzingAna)
	timer.Stop(timer.FuzzingAna)

	id := on.GetID()
	vc := on.GetVC()

	incFuzzingCounter(on)

	if on.GetSuc() {
		executedOnce[id] = &ConcurrentEntry{Elem: on, Counter: getFuzzingCounter(on), Type: CEOnce}
		return
	}

	if exec, ok := executedOnce[id]; ok {
		if clock.GetHappensBefore(exec.Elem.GetVC(), vc) == clock.Concurrent {
			fuzzingFlowOnce = append(fuzzingFlowOnce, *exec)
		}
	}
}

// GetConcurrentInfoForFuzzing returns the required fuzzing information for
// the flow fuzzing mutation.
//
// Returns:
//   - *[]ConcurrentEntry: once that can be delayed in flow fuzzing
//   - *[]ConcurrentEntry: mutex operations that can be delayed in flow fuzzing
//   - *[]ConcurrentEntry: send that can be delayed in flow fuzzing
//   - *[]ConcurrentEntry: recv that can be delayed in flow fuzzing
func GetConcurrentInfoForFuzzing() (*[]ConcurrentEntry, *[]ConcurrentEntry, *[]ConcurrentEntry, *[]ConcurrentEntry) {
	return &fuzzingFlowOnce, &fuzzingFlowMutex, &fuzzingFlowSend, &fuzzingFlowRecv
}

// getFuzzingCounter returns the fuzzing counter for an element. If the element
// has no counter it is set to 0. The fuzzing counter gives for a given
// primitive how often an operation has been executed on the primitive before.
//
// Parameter:
//   - te TraceElement: The trace element to get the counter for
//
// Returns:
//   - int: the current fuzzing counter for the element
func getFuzzingCounter(te trace.TraceElement) int {
	id := te.GetID()
	pos := te.GetPos()

	if _, ok := fuzzingCounter[id]; !ok {
		return 0
	}

	if val, ok := fuzzingCounter[id][pos]; ok {
		return val
	}
	return 0
}

// incFuzzingCounter increases the fuzzing counter of a given element
//
// Parameter:
//   - te TraceElement: The element to increase the counter for
func incFuzzingCounter(te trace.TraceElement) {
	id := te.GetID()
	pos := te.GetPos()

	if _, ok := fuzzingCounter[id]; !ok {
		fuzzingCounter[id] = make(map[string]int)
	}

	fuzzingCounter[id][pos] = fuzzingCounter[id][pos] + 1
}
