// Copyright (c) 2025 Erik Kassubek
//
// File: io.go
// Brief: Write goPie active to file
//
// Author: Erik Kassubek
// Created: 2025-07-03
//
// License: BSD-3-Clause

package gopie

import (
	"advocate/trace"
	"fmt"
	"os"
	"path/filepath"
)

// WriteMutActive writes the element in the chain into a rewriteActive.log
// file for use in GoPie
//
// Parameter
//   - fuzzingTracePath string: path to the trace folder
//   - tr *trace.Trace: the trace to write
//   - mut *chain: chain to write
//   - partTime int: if 0, the replay will partial replay from the beginning
//     otherwise it will switch to partial replay when the element with this
//     time is the next element to be replayed
func writeMutActive(fuzzingTracePath string, tr *trace.Trace, mut *Chain, partTime int) {
	activePath := filepath.Join(fuzzingTracePath, "replay_active.log")

	f, err := os.Create(activePath)
	if err != nil {
		return
	}

	defer f.Close()

	f.WriteString(fmt.Sprintf("%d\n", partTime))

	// find the counter for all elements in the mut
	mutCounter := make(map[int]int)
	posCounter := make(map[string]int)
	mutTime := make(map[int]int)
	for _, elem := range mut.Elems {
		mutCounter[elem.GetTraceID()] = 0
	}

	traceIter := tr.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		traceID := elem.GetTraceID()
		pos := elem.GetPos()
		posCounter[pos]++
		if _, ok := mutCounter[traceID]; ok { // is in chain
			mutCounter[traceID] = posCounter[pos]
			mutTime[traceID] = elem.GetTSort()
		}
	}

	for _, elem := range mut.Elems {
		traceID := elem.GetTraceID()
		// key := fmt.Sprintf("%d:%s,%d,%d\n", elem.GetRoutine(), elem.GetPos(), mutTPre[traceID], mutCounter[traceID])
		key := fmt.Sprintf("%d:%s,%d,%d\n", elem.GetRoutine(), elem.GetPos(), mutTime[traceID], mutCounter[traceID])
		f.WriteString(key)
	}
}
