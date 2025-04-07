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
	"analyzer/utils"
)

func getConcurrentSendForFuzzing(ch *TraceElementChannel, vc map[int]clock.VectorClock) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	id := ch.id
	routine := ch.routine

	incFuzzingCounter(ch)

	if ch.GetTPost() != 0 {
		return
	}

	for r, elem := range lastSendRoutine {
		if r == routine {
			continue
		}

		if elem[ch.id].vc.GetClock() == nil {
			continue
		}

		happensBefore := clock.GetHappensBefore(elem[id].vc, vc[routine])
		if happensBefore == clock.Concurrent {
			elem2 := elem[id].elem
			fuzzingFlowSend = append(fuzzingFlowSend, ConcurrentEntry{Elem: elem2, Counter: getFuzzingCounter(elem2), Type: CERecv})
		}
	}

	if ch.tPost != 0 {
		if _, ok := lastSendRoutine[routine]; !ok {
			lastSendRoutine[routine] = make(map[int]elemWithVc)
		}

		lastSendRoutine[routine][id] = elemWithVc{vc[routine].Copy(), ch}
	}
}

func checkForConcurrentRecv(ch *TraceElementChannel, vc map[int]clock.VectorClock) {
	if analysisFuzzing {
		timer.Start(timer.FuzzingAna)
		defer timer.Stop(timer.FuzzingAna)
	}
	timer.Start(timer.AnaConcurrent)
	defer timer.Stop(timer.AnaConcurrent)

	id := ch.id
	routine := ch.routine

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

			if ch.tPost == 0 {
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

			results.Result(results.WARNING, results.AConcurrentRecv,
				"recv", []results.ResultElem{arg1}, "recv", []results.ResultElem{arg2})
		}
	}

	if ch.tPost != 0 {
		if _, ok := lastRecvRoutine[routine]; !ok {
			lastRecvRoutine[routine] = make(map[int]elemWithVc)
		}

		lastRecvRoutine[routine][id] = elemWithVc{vc[routine].Copy(), ch}
	}
}

func getConcurrentMutexForFuzzing(mu *TraceElementMutex) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	// operation executed normally
	if mu.IsSuc() {
		return
	}

	// not executed try lock
	// get currently hold lock because of witch the try lock failed

	if val, ok := currentlyHoldLock[mu.id]; !ok || val == nil {
		utils.LogError("Failed trylock even throw mutex is not locked: ", mu.ToString())
	}

	elem := currentlyHoldLock[mu.id]

	if clock.GetHappensBefore(mu.GetVC(), elem.GetVC()) == clock.Concurrent {
		fuzzingFlowMutex = append(fuzzingFlowMutex, ConcurrentEntry{Elem: elem, Counter: getFuzzingCounter(elem), Type: CEMutex})
	}

}

func getConcurrentOnceForFuzzing(on *TraceElementOnce) {
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

func GetConcurrentInfoForFuzzing() (*[]ConcurrentEntry, *[]ConcurrentEntry, *[]ConcurrentEntry, *[]ConcurrentEntry) {
	return &fuzzingFlowOnce, &fuzzingFlowMutex, &fuzzingFlowSend, &fuzzingFlowRecv
}

func getFuzzingCounter(te TraceElement) int {
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

func incFuzzingCounter(te TraceElement) {
	id := te.GetID()
	pos := te.GetPos()

	if _, ok := fuzzingCounter[id]; !ok {
		fuzzingCounter[id] = make(map[string]int)
	}

	fuzzingCounter[id][pos] = fuzzingCounter[id][pos] + 1
}
