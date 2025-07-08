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
	"advocate/analysis/concurrent/clock"
	"advocate/analysis/data"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// getConcurrentSendForFuzzing checks if for the given send, if there is a
// concurrent send on the same channel. If there is, the information is stored
// in data.FuzzingFlowSend. This is used for fuzzing.
//
// Parameter:
//   - sender *TraceElementChannel: Send trace element
func getConcurrentSendForFuzzing(sender *trace.ElementChannel) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	id := sender.GetID()
	routine := sender.GetRoutine()

	incFuzzingCounter(sender)

	if sender.GetTPost() != 0 {
		return
	}

	for r, elem := range data.LastSendRoutine {
		if r == routine {
			continue
		}

		if elem[id].Vc == nil || elem[id].Vc.GetClock() == nil {
			continue
		}

		happensBefore := clock.GetHappensBefore(elem[id].Vc, data.CurrentVC[routine])
		if happensBefore == clock.Concurrent {
			elem2 := elem[id].Elem
			data.FuzzingFlowSend = append(data.FuzzingFlowSend, data.ConcurrentEntry{Elem: elem2, Counter: getFuzzingCounter(elem2), Type: data.CERecv})
		}
	}

	if sender.GetTPost() != 0 {
		if _, ok := data.LastSendRoutine[routine]; !ok {
			data.LastSendRoutine[routine] = make(map[int]data.ElemWithVc)
		}

		data.LastSendRoutine[routine][id] = data.ElemWithVc{Vc: data.CurrentVC[routine].Copy(), Elem: sender}
	}
}

// checkForConcurrentRecv checks if for the given recv, if there is a
// concurrent recv on the same channel. If there is, the information is stored
// in data.FuzzingFlowRecv.
//
// Parameter:
//   - ch *TraceElementChannel: recv trace element
func checkForConcurrentRecv(ch *trace.ElementChannel, vc map[int]*clock.VectorClock) {
	if data.AnalysisFuzzing {
		timer.Start(timer.FuzzingAna)
		defer timer.Stop(timer.FuzzingAna)
	}
	timer.Start(timer.AnaConcurrent)
	defer timer.Stop(timer.AnaConcurrent)

	id := ch.GetID()
	routine := ch.GetRoutine()

	incFuzzingCounter(ch)

	for r, elem := range data.LastRecvRoutine {
		if r == routine {
			continue
		}

		if elem[id].Vc == nil || elem[id].Vc.GetClock() == nil {
			continue
		}

		happensBefore := clock.GetHappensBefore(elem[id].Vc, vc[routine])
		if happensBefore == clock.Concurrent {

			elem2 := elem[id].Elem

			if ch.GetTPost() == 0 {
				data.FuzzingFlowRecv = append(data.FuzzingFlowRecv, data.ConcurrentEntry{Elem: elem2, Counter: getFuzzingCounter(elem2), Type: data.CERecv})
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

			results.Result(results.WARNING, helper.AConcurrentRecv,
				"recv", []results.ResultElem{arg1}, "recv", []results.ResultElem{arg2})
		}
	}

	if ch.GetTPost() != 0 {
		if _, ok := data.LastRecvRoutine[routine]; !ok {
			data.LastRecvRoutine[routine] = make(map[int]data.ElemWithVc)
		}

		data.LastRecvRoutine[routine][id] = data.ElemWithVc{Vc: vc[routine].Copy(), Elem: ch}
	}
}

// getConcurrentMutexForFuzzing checks if for the given mutex operations, if there is a
// concurrent mutex operations on the same mutex. If there is, the information is stored
// in data.FuzzingFlowMutex.
//
// Parameter:
//   - mu *TraceElementMutex: mutex operations
func getConcurrentMutexForFuzzing(mu *trace.ElementMutex) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	// operation executed normally
	if mu.IsSuc() {
		return
	}

	id := mu.GetID()

	// not executed try lock
	// get currently hold lock because of witch the try lock failed

	if val, ok := data.CurrentlyHoldLock[id]; !ok || val == nil {
		log.Error("Failed trylock even throw mutex is not locked: ", mu.ToString())
	}

	elem := data.CurrentlyHoldLock[id]

	if clock.GetHappensBefore(mu.GetVC(), elem.GetVC()) == clock.Concurrent {
		data.FuzzingFlowMutex = append(data.FuzzingFlowMutex, data.ConcurrentEntry{Elem: elem, Counter: getFuzzingCounter(elem), Type: data.CEMutex})
	}

}

// getConcurrentOnceForFuzzing checks if for the given once operations, if there is a
// concurrent once operations on the same primitive. If there is, the information is stored
// in data.FuzzingFlowOnce.
//
// Parameter:
//   - on *TraceElementOnce: once.Do operations
func getConcurrentOnceForFuzzing(on *trace.ElementOnce) {
	timer.Start(timer.FuzzingAna)
	timer.Stop(timer.FuzzingAna)

	id := on.GetID()
	vc := on.GetVC()

	incFuzzingCounter(on)

	if on.GetSuc() {
		data.ExecutedOnce[id] = &data.ConcurrentEntry{Elem: on, Counter: getFuzzingCounter(on), Type: data.CEOnce}
		return
	}

	if exec, ok := data.ExecutedOnce[id]; ok {
		if clock.GetHappensBefore(exec.Elem.GetVC(), vc) == clock.Concurrent {
			data.FuzzingFlowOnce = append(data.FuzzingFlowOnce, *exec)
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
func GetConcurrentInfoForFuzzing() (*[]data.ConcurrentEntry, *[]data.ConcurrentEntry, *[]data.ConcurrentEntry, *[]data.ConcurrentEntry) {
	return &data.FuzzingFlowOnce, &data.FuzzingFlowMutex, &data.FuzzingFlowSend, &data.FuzzingFlowRecv
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
func getFuzzingCounter(te trace.Element) int {
	id := te.GetID()
	pos := te.GetPos()

	if _, ok := data.FuzzingCounter[id]; !ok {
		return 0
	}

	if val, ok := data.FuzzingCounter[id][pos]; ok {
		return val
	}
	return 0
}

// incFuzzingCounter increases the fuzzing counter of a given element
//
// Parameter:
//   - te TraceElement: The element to increase the counter for
func incFuzzingCounter(te trace.Element) {
	id := te.GetID()
	pos := te.GetPos()

	if _, ok := data.FuzzingCounter[id]; !ok {
		data.FuzzingCounter[id] = make(map[string]int)
	}

	data.FuzzingCounter[id][pos] = data.FuzzingCounter[id][pos] + 1
}
