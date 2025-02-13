// Copyrigth (c) 2024 Erik Kassubek
//
// File: analysisWaitGroup.go
// Brief: Trace analysis for possible negative wait group counter
//
// Author: Erik Kassubek
// Created: 2023-11-24
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/results"
	"analyzer/utils"
	"errors"
	"fmt"
	"log"
)

/*
 * Collect all adds and dones for the analysis
 * Args:
 *    wa *TraceElementWait: the trace wait or done element
 */
func checkForDoneBeforeAddChange(wa *TraceElementWait) {
	if wa.delta > 0 {
		checkForDoneBeforeAddAdd(wa)
	} else if wa.delta < 0 {
		checkForDoneBeforeAddDone(wa)
	} else {
		// checkForImpossibleWait(routine, id, pos, vc)
	}
}

/*
 * Collect all adds for the analysis
 * Args:
 *    wa *TraceElementWait: the trace wait element
 */
func checkForDoneBeforeAddAdd(wa *TraceElementWait) {
	// if necessary, create maps and lists
	if _, ok := wgAdd[wa.id]; !ok {
		wgAdd[wa.id] = make([]TraceElement, 0)
	}

	// add the vector clock and position to the list
	for i := 0; i < wa.delta; i++ {
		wgAdd[wa.id] = append(wgAdd[wa.id], wa)
	}
}

/*
 * Collect all dones for the analysis
 * Args:
 *    wa *TraceElementWait: the trace done element
 */
func checkForDoneBeforeAddDone(wa *TraceElementWait) {
	// if necessary, create maps and lists
	if _, ok := wgDone[wa.id]; !ok {
		wgDone[wa.id] = make([]TraceElement, 0)

	}

	// add the vector clock and position to the list
	wgDone[wa.id] = append(wgDone[wa.id], wa)
}

/*
 * Check if a wait group counter could become negative
 * For each done operation, build a bipartite st graph.
 * Use the Ford-Fulkerson algorithm to find the maximum flow.
 * If the maximum flow is smaller than the number of done operations, a negative wait group counter is possible.
 */
func checkForDoneBeforeAdd() {
	for id := range wgAdd { // for all waitgroups

		graph := buildResidualGraph(wgAdd[id], wgDone[id])

		maxFlow, graph, err := calculateMaxFlow(graph)
		if err != nil {
			fmt.Println("Could not check for done before add: ", err)
		}
		nrDone := len(wgDone[id])

		addsNegWg := []TraceElement{}
		donesNegWg := []TraceElement{}

		if maxFlow < nrDone {
			// sort the adds and dones, that do not have a partner is such a way,
			// that the i-th add in the result message is concurrent with the
			// i-th done in the result message

			for _, add := range wgAdd[id] {
				if !utils.ContainsString(graph["t"], add.GetTID()) {
					addsNegWg = append(addsNegWg, add)
				}
			}

			for _, dones := range graph["s"] {
				doneVcTID, err := getDoneElemFromTID(id, dones)
				if err != nil {
					log.Print(err.Error())
				} else {
					donesNegWg = append(donesNegWg, doneVcTID)
				}
			}

			addsNegWgSorted := make([]TraceElement, 0)
			donesNEgWgSorted := make([]TraceElement, 0)

			for i := 0; i < len(addsNegWg); i++ {
				for j := 0; j < len(donesNegWg); j++ {
					if clock.GetHappensBefore(addsNegWg[i].GetVC(), donesNegWg[j].GetVC()) == clock.Concurrent {
						addsNegWgSorted = append(addsNegWgSorted, addsNegWg[i])
						donesNEgWgSorted = append(donesNEgWgSorted, donesNegWg[j])
						// remove the element from the list
						addsNegWg = append(addsNegWg[:i], addsNegWg[i+1:]...)
						donesNegWg = append(donesNegWg[:j], donesNegWg[j+1:]...)
						// fix the index
						i--
						j = 0
					}
				}
			}

			args1 := []results.ResultElem{} // dones
			args2 := []results.ResultElem{} // adds

			for _, done := range donesNEgWgSorted {
				if done.GetTID() == "\n" {
					continue
				}
				file, line, tPre, err := infoFromTID(done.GetTID())
				if err != nil {
					log.Print(err.Error())
					return
				}

				args1 = append(args1, results.TraceElementResult{
					RoutineID: done.GetRoutine(),
					ObjID:     id,
					TPre:      tPre,
					ObjType:   "WD",
					File:      file,
					Line:      line,
				})
			}

			for _, add := range addsNegWgSorted {
				if add.GetTID() == "\n" {
					continue
				}
				file, line, tPre, err := infoFromTID(add.GetTID())
				if err != nil {
					log.Print(err.Error())
					continue
				}

				args2 = append(args2, results.TraceElementResult{
					RoutineID: add.GetRoutine(),
					ObjID:     id,
					TPre:      tPre,
					ObjType:   "WA",
					File:      file,
					Line:      line,
				})

			}

			results.Result(results.CRITICAL, results.PNegWG,
				"done", args1, "add", args2)
		}
	}
}

func getDoneElemFromTID(id int, tID string) (TraceElement, error) {
	for _, done := range wgDone[id] {
		if done.GetTID() == tID {
			return done, nil
		}
	}
	return nil, errors.New("Could not find done operation with tID " + tID)
}
