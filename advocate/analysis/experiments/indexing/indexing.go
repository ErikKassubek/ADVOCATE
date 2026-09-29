// Copyright (c) 2024 Erik Kassubek
//
// File: indexing.go
// Brief: Implement id abstraction based on light-weight execution indexing
//        Based on P. Joshi, C.-S. Park, K. Sen, and M. Naik, “A randomized dynamic program analysis technique for detecting real deadlocks,” SIGPLAN Not., vol. 44, no. 6, pp. 110–120, Jun. 2009, doi: 10.1145/1543135.1542489.
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package indexing

import (
	"advocate/analysis/a_base"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
	"fmt"
	"strings"
)

const k = 3

type stackElem struct {
	isC bool
	c   string
	q   int
}

func (this *stackElem) String() string {
	if this.isC {
		return this.c
	}

	return fmt.Sprintf("%d", this.q)
}

var depth = make(map[int]int)                         // rout id -> depth
var counters = make(map[int]map[int]map[string]int)   //  rout id -> d_t -> c (pos string) -> value
var callStack = make(map[int]*types.Stack[stackElem]) // rout id -> call stack of c

var indices = make(map[trace.Element]string)

func initValues(tr *trace.Trace) {
	for rout := range tr.GetRoutines() {
		depth[rout] = 0
		counters[rout] = make(map[int]map[string]int)
		callStack[rout] = types.NewStack[stackElem]()
	}
}

func initCounters(routID, d int, c string) {
	if _, ok := counters[routID][d]; !ok {
		counters[routID][d] = make(map[string]int)
	}
	if _, ok := counters[routID][d][c]; ok {
		counters[routID][d][c] = 0
	}
}

func pushToCS(routID, d int, c string) {
	c_stack_elem := stackElem{true, c, 0}
	q_stacl_elem := stackElem{false, "", counters[routID][d][c]}
	callStack[routID].Push(c_stack_elem)
	callStack[routID].Push(q_stacl_elem)
}

func BuildExecutionIndexing() {
	tr := &a_base.MainTrace
	traceIter := tr.AsIterator()

	initValues(tr)

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		routID := elem.RoutineID()
		// c := elem.Pos().Short()
		c := fmt.Sprintf("%d", elem.Line())
		d := depth[routID]
		initCounters(routID, d, c)

		switch elem.(type) {
		case *trace.ElementFunc:

			counters[routID][d][c] = counters[routID][d][c] + 1

			pushToCS(routID, d, c)

			depth[routID] = depth[routID] + 1

			d = depth[routID]
			for c_iter := range counters[routID][d] {
				counters[routID][d][c_iter] = 0
			}
		case *trace.ElementReturn:
			depth[routID] = depth[routID] - 1
			callStack[routID].Pop()
			callStack[routID].Pop()
		case *trace.ElementAlloc, *trace.ElementFork: // in the original paper this is applied only to new (alloc). We are also interested in forks to give routines ids
			counters[routID][d][c] = counters[routID][d][c] + 1

			pushToCS(routID, d, c)

			arr := callStack[routID].AsArray()

			var index strings.Builder

			lastInd := len(arr) - 1
			for count := 0; count < len(arr); count++ {
				if count >= 2*k {
					return
				}

				i := len(arr) - 1 - count

				if i != lastInd {
					index.WriteString(",")
				}
				val := arr[i]
				index.WriteString(val.String())

			}

			indices[elem] = index.String()

			callStack[routID].Pop()
			callStack[routID].Pop()
		}
	}

}

func PrintIndexes() {
	log.Debug("INDICES")
	for elem, index := range indices {
		if elem.File() == "/home/advocate/Advocate/Experiments/Indexing/simpleIndexing/main.go" {
			log.Debug(elem, " -> ", index)
		}
	}
}
