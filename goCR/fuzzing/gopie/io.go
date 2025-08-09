//
// File: io.go
// Brief: Write goPie active to file
//
// Created: 2025-07-03
//
// License: BSD-3-Clause

package gopie

import (
	"fmt"
	"goCR/trace"
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
	mutCounter := make(map[string]int)
	posCounter := make(map[string]int)
	mutTime := make(map[string]int)
	for _, elem := range mut.Elems {
		mutCounter[getRoutPos(elem)] = 0
	}

	traceIter := tr.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		routPos := getRoutPos(elem)
		posCounter[routPos]++
		if _, ok := mutCounter[routPos]; ok { // is in chain
			mutCounter[routPos] = posCounter[routPos]
			mutTime[routPos] = elem.GetTSort()
		}
	}

	for _, elem := range mut.Elems {
		routPos := getRoutPos(elem)
		// key := fmt.Sprintf("%d:%s,%d,%d\n", elem.GetRoutine(), elem.GetPos(), mutTPre[traceID], mutCounter[traceID])
		key := fmt.Sprintf("%d:%s,%d,%d\n", elem.GetRoutine(), elem.GetPos(), mutTime[routPos], mutCounter[routPos])
		f.WriteString(key)
	}
}

func getRoutPos(elem trace.Element) string {
	return fmt.Sprintf("%d:%s", elem.GetRoutine(), elem.GetPos())
}
