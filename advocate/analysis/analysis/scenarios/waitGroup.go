// Copyright (c) 2024 Erik Kassubek
//
// File: analysisWaitGroup.go
// Brief: Trace analysis for possible negative wait group counter
//
// Author: Erik Kassubek
// Created: 2023-11-24
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/data"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
	"advocate/utils/types"
	"fmt"
)

// CheckForDoneBeforeAddChange collect all adds and dones for the analysis
//
// Parameter:
//   - wa *TraceElementWait: the trace wait or done element
func CheckForDoneBeforeAddChange(wa *trace.ElementWait) {
	timer.Start(timer.AnaWait)
	defer timer.Stop(timer.AnaWait)

	delta := wa.GetDelta()

	if delta > 0 {
		CheckForDoneBeforeAddAdd(wa)
	} else if delta < 0 {
		CheckForDoneBeforeAddDone(wa)
	} else {
		// checkForImpossibleWait(routine, id, pos, vc)
	}
}

// CheckForDoneBeforeAddAdd collect all adds for the analysis
//
// Parameter:
//   - wa *TraceElementWait: the trace wait element
func CheckForDoneBeforeAddAdd(wa *trace.ElementWait) {
	id := wa.GetID()

	// if necessary, create maps and lists
	if _, ok := data.WGAddData[id]; !ok {
		data.WGAddData[id] = make([]trace.Element, 0)
	}

	// add the vector clock and position to the list
	data.WGAddData[id] = append(data.WGAddData[id], wa)
}

// CheckForDoneBeforeAddDone collect all dones for the analysis
//
// Parameter:
//   - wa *TraceElementWait: the trace done element
func CheckForDoneBeforeAddDone(wa *trace.ElementWait) {
	id := wa.GetID()

	// if necessary, create maps and lists
	if _, ok := data.WgDoneData[id]; !ok {
		data.WgDoneData[id] = make([]trace.Element, 0)

	}

	// add the vector clock and position to the list
	data.WgDoneData[id] = append(data.WgDoneData[id], wa)
}

// CheckForDoneBeforeAdd checks if a wait group counter could become negative
// For each done operation, build a bipartite st graph.
// Use the Ford-Fulkerson algorithm to find the maximum flow.
// If the maximum flow is smaller than the number of done operations, a negative wait group counter is possible.
func CheckForDoneBeforeAdd() {
	timer.Start(timer.AnaWait)
	defer timer.Stop(timer.AnaWait)

	for id := range data.WGAddData { // for all waitgroups
		graph := buildResidualGraph(data.WGAddData[id], data.WgDoneData[id])

		maxFlow, graph, err := calculateMaxFlow(graph)
		if err != nil {
			fmt.Println("Could not check for done before add: ", err)
		}
		nrDone := len(data.WgDoneData[id])

		addsNegWg := make([]trace.Element, 0)
		donesNegWg := make([]trace.Element, 0)

		if maxFlow < nrDone {
			// sort the adds and dones, that do not have a partner is such a way,
			// that the i-th add in the result message is concurrent with the
			// i-th done in the result message

			for _, add := range data.WGAddData[id] {
				if !types.Contains(graph[drain], add) {
					addsNegWg = append(addsNegWg, add)
				}
			}

			for _, dones := range graph[source] {
				donesNegWg = append(donesNegWg, dones)
			}

			addsNegWgSorted := make([]trace.Element, 0)
			donesNEgWgSorted := make([]trace.Element, 0)

			for i := 0; i < len(addsNegWg); {
				removed := false
				for j := 0; j < len(donesNegWg); {
					if clock.GetHappensBefore(addsNegWg[i].GetVC(), donesNegWg[j].GetVC()) == hb.Concurrent {
						addsNegWgSorted = append(addsNegWgSorted, addsNegWg[i])
						donesNEgWgSorted = append(donesNEgWgSorted, donesNegWg[j])
						// remove the element from the list
						addsNegWg = append(addsNegWg[:i], addsNegWg[i+1:]...)
						donesNegWg = append(donesNegWg[:j], donesNegWg[j+1:]...)
						// fix the index
						removed = true
						break
					} else {
						j++
					}
				}
				if !removed {
					i++
				}
			}

			args1 := []results.ResultElem{} // dones
			args2 := []results.ResultElem{} // adds

			for _, done := range donesNEgWgSorted {
				if done.GetTID() == "\n" {
					continue
				}
				file, line, tPre, err := trace.InfoFromTID(done.GetTID())
				if err != nil {
					log.Error(err.Error())
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
				file, line, tPre, err := trace.InfoFromTID(add.GetTID())
				if err != nil {
					log.Error(err.Error())
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

			results.Result(results.CRITICAL, helper.PNegWG,
				"done", args1, "add", args2)
		}
	}
}
