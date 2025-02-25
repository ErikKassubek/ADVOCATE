// Copyrigth (c) 2024 Erik Kassubek
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
	timemeasurement "analyzer/timeMeasurement"
	"analyzer/utils"
)

/*
 * Check if there are multiple concurrent receive operations on the same channel.
 * Such concurrent recv can lead to nondeterministic behaviour.
 * If such a situation is detected, it is logged.
 * Call this function on a recv.
 * Args:
 *  ch (*TraceElementChannel): The trace element
 *  routine (int): routine of the recv
 *  tID (int): tID of the recv operation
 *  vc (int): vector clock of the recv operation
 */
// TODO: make this like the others
func checkForConcurrentCom(ch *TraceElementChannel, vc map[int]clock.VectorClock) {
	timemeasurement.Start("other")

	checkForConcurrentRevc(ch, vc)

	timemeasurement.End("other")

	getConcurrentSendForFuzzing(ch)

}

func getConcurrentSendForFuzzing(ch *TraceElementChannel) {

}

func checkForConcurrentRevc(ch *TraceElementChannel, vc map[int]clock.VectorClock) {
	for r, elem := range lastRecvRoutine {
		if r == ch.routine {
			continue
		}

		if elem[ch.id].vc.GetClock() == nil {
			continue
		}

		happensBefore := clock.GetHappensBefore(elem[ch.id].vc, vc[ch.routine])
		if happensBefore == clock.Concurrent {

			file1, line1, tPre1, err := infoFromTID(ch.GetTID())
			if err != nil {
				utils.LogError(err.Error())
				return
			}

			elem2 := lastRecvRoutine[r][ch.id].elem
			file2, line2, err := posFromPosString(elem2.GetPos())

			if err != nil {
				utils.LogErrorf("Error in posFromPosString: %s", err.Error())
			}

			tPre2 := elem2.GetTPre()

			arg1 := results.TraceElementResult{
				RoutineID: ch.routine,
				ObjID:     ch.id,
				TPre:      tPre1,
				ObjType:   "CR",
				File:      file1,
				Line:      line1,
			}

			arg2 := results.TraceElementResult{
				RoutineID: r,
				ObjID:     ch.id,
				TPre:      tPre2,
				ObjType:   "CR",
				File:      file2,
				Line:      line2,
			}

			results.Result(results.WARNING, results.AConcurrentRecv,
				"recv", []results.ResultElem{arg1}, "recv", []results.ResultElem{arg2})
		}
	}

	if ch.tPost != 0 {
		if _, ok := lastRecvRoutine[ch.routine]; !ok {
			lastRecvRoutine[ch.routine] = make(map[int]elemWithVc)
		}

		lastRecvRoutine[ch.routine][ch.id] = elemWithVc{vc[ch.routine].Copy(), ch}
	}
}

func getConcurrentMutexForFuzzing(mu *TraceElementMutex) {
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
		fuzzingFlowMutex = append(fuzzingFlowMutex, ConcurrentEntry{Elem: elem, Counter: lockCounter[elem.id][elem.pos], Type: CEMutex})
	}

}

func getConcurrentOnceForFuzzing(on *TraceElementOnce) {
	id := on.GetID()
	vc := on.GetVC()
	pos := on.GetPos()

	if _, ok := onceCounter[id]; !ok {
		onceCounter[id] = make(map[string]int)
	}
	onceCounter[id][pos] = onceCounter[id][pos] + 1

	if on.GetSuc() {
		executedOnce[id] = &ConcurrentEntry{Elem: on, Counter: onceCounter[id][pos], Type: CEOnce}
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
