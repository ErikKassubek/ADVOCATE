// Copyright (c) 2024 Erik Kassubek
//
// File: blocked.go
// Brief: Trace analysis for routine blocks
//
// Author: Erik Kassubek
// Created: 2024-01-28
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/results/results"
	"advocate/utils/types"
)

func Blocked() error {
	tr := &a_base.MainTrace
	blocked := tr.GetBlocked()
	ref := tr.GetObjAware()

	// blocked routines
	b := make(map[int]struct{})
	i := 0
	for rout := range blocked {
		b[rout] = struct{}{}
		i++
	}

	l := a_base.MainTrace.GetNotReturned(true)

	log.Debug(b)
	log.Debug(l)
	log.Debug(ref)

	for {
		r := make([]int, 0)

		for routB := range b {
			for _, routL := range l {
				if types.Contains(ref[routL], routB) {
					log.Debug(routL, " -> ", routB)
					r = append(r, routB)
					break
				}
			}
		}

		for _, routR := range r {
			l = append(l, routR)
			delete(b, routR)
		}

		if len(r) == 0 {
			break
		}
	}

	log.Debug(b)

	cyclic := checkCyclic(b, ref)

	for rout := range cyclic {
		delete(b, rout)
	}

	reportBlocking(cyclic, helper.ADeadlock)
	reportBlocking(cyclic, helper.ABlocking)

	return nil
}

// check for cyclic dependencies
func checkCyclic(b map[int]struct{}, ref map[int][]int) map[int]struct{} {
	graph := map[int][]int{}
	selfLoop := map[int]bool{}

	for rID := range b {
		for _, rID2 := range ref[rID] {
			graph[rID] = append(graph[rID], rID2)
			if rID == rID2 {
				selfLoop[rID] = true
			}
		}
	}

	// Tarjan SCC
	index := 0
	stack := []int{}
	onStack := map[int]bool{}
	indices := map[int]int{}
	lowlink := map[int]int{}

	result := map[int]struct{}{}

	var strongConnect func(int)
	strongConnect = func(v int) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range graph[v] {
			if _, seen := indices[w]; !seen {
				strongConnect(w)
				lowlink[v] = min(lowlink[v], lowlink[w])
			} else if onStack[w] {
				lowlink[v] = min(lowlink[v], indices[w])
			}
		}

		// Root of SCC
		if lowlink[v] == indices[v] {
			scc := map[int]struct{}{}
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc[w] = struct{}{}
				if w == v {
					break
				}
			}

			// must be cycle with multiple elements
			if len(scc) == 1 {
				return
			}

			// must be closed
			for r := range scc {
				for _, r2 := range ref[r] {
					if _, ok := scc[r2]; !ok {
						return
					}
				}
			}

			for r := range scc {
				result[r] = struct{}{}
			}
		}
	}

	for r := range b {
		if _, seen := indices[r]; !seen {
			strongConnect(r)
		}
	}

	return result
}

func reportBlocking(routs map[int]struct{}, rt helper.ResultType) {
	obj := make([]results.ResultElem, 0)
	tr := &a_base.MainTrace

	for r := range routs {
		elem := tr.GetLastElemInRout(r)

		objRes := results.TraceElementResult{
			RoutineID: r,
			ObjID:     -1,
			TPre:      -1,
			ObjType:   elem.GetType(true),
			File:      elem.GetFile(),
			Line:      elem.GetLine(),
		}

		obj = append(obj, objRes)
	}

	if len(obj) != 0 {
		results.Result(results.CRITICAL, rt,
			"Blocked", obj, "", []results.ResultElem{})
	}
}

// func Blocked() error {
// 	output := filepath.Join(paths.ProgDir, paths.NameOutput)

// 	file, err := os.Open(output)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)

// 	buf := make([]byte, 0, 1024*1024) // 1 MB initial buffer
// 	scanner.Buffer(buf, 10*1024*1024)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if strings.HasPrefix(line, "LEAK_GC@") {
// 			err = readGCBlocked(line, false)
// 		} else if strings.HasPrefix(line, "DEADLOCK_GC@") {
// 			err = readGCBlocked(line, true)
// 		}
// 		if err != nil {
// 			log.Errorf(err.Error())
// 		}
// 	}

// 	// if err := scanner.Err(); err != nil {
// 	// 	return err
// 	// }

// 	reportGCBlocked()
// 	reportNonDeadlockLeaks()

// 	return nil
// }

// func readGCBlocked(line string, deadlock bool) error {
// 	fields := strings.Split(line, "@")

// 	if len(fields) != 4 {
// 		return fmt.Errorf("Could not process deadlock %s", line)
// 	}

// 	routineID, err := strconv.Atoi(fields[1])
// 	if err != nil {
// 		return err
// 	}

// 	// only count deadlocks that are also in the trace
// 	if obj, ok := leaks[routineID]; ok {
// 		var objRes results.ResultElem
// 		objRes = obj.arg1[0]
// 		delete(leaks, routineID)
// 		if deadlock {
// 			GCDeadlock = append(GCDeadlock, objRes)
// 		} else {
// 			GCLeak = append(GCLeak, objRes)
// 		}
// 	}

// 	// if !objResSet {
// 	// 	objRes = results.TraceElementResult{
// 	// 		RoutineID: routineID,
// 	// 		ObjID:     -1,
// 	// 		TPre:      -1,
// 	// 		ObjType:   getObjectType(fields[3]),
// 	// 		File:      file,
// 	// 		Line:      line,
// 	// 	}
// 	// }

// 	return nil
// }

// reportGCBlocked creates a result for all elements that are in a deadlock
// func reportGCBlocked() {
// 	if len(GCLeak) > 0 {
// 		results.Result(results.CRITICAL, helper.ALeak,
// 			"Blocked", GCLeak, "", []results.ResultElem{})
// 	}

// 	if len(GCDeadlock) > 0 {
// 		results.Result(results.CRITICAL, helper.ADeadlock,
// 			"Blocked", GCDeadlock, "", []results.ResultElem{})
// 	}
// }

// // reportNonDeadlockLeaks creates results for all elements that have a leek
// // without being in a deadlock
// func reportNonDeadlockLeaks() {
// 	for _, leak := range leaks {
// 		results.Result(results.CRITICAL, leak.resultType,
// 			leak.argType1, leak.arg1, leak.argType1, leak.arg2)
// 	}
// }
