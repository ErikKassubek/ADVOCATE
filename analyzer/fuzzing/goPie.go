// Copyright (c) 2025 Erik Kassubek
//
// File: goPie.go
// Brief: Main file for goPie fuzzing
//
// Author: Erik Kassubek
// Created: 2025-03-22
//
// License: BSD-3-Clause

package fuzzing

import (
	"analyzer/analysis"
	"analyzer/io"
	"fmt"
	"path/filepath"
	"sort"
)

func createGoPieMut(pkgPath string) error {
	energy := getEnergy()

	mutations := make(map[string]chain)

	for _, sc := range schedulingChains {
		muts := mutate(sc, energy)
		for key, mut := range muts {
			mutations[key] = mut
		}
	}

	for _, mut := range mutations {
		traces := analysis.GetTraces()
		oldTrace, err := analysis.CopyCurrentTrace()
		if err != nil {
			return fmt.Errorf("Could not copy trace: ", err.Error())
		}

		tPosts := make([]int, 0)
		routines := make(map[int]struct{})
		for _, elem := range mut.elems {
			tPosts = append(tPosts, elem.GetTPost())
			routines[elem.GetRoutine()] = struct{}{}
		}

		sort.Ints(tPosts)

		for i, elem := range mut.elems {
			routine, index := elem.GetTraceIndex()
			traces[routine][index].SetTSort(tPosts[i])
		}

		// TODO: is this sort necessary
		for routine := range routines {
			traces[routine] = analysis.SortTrace(traces[routine])
		}

		// remove all elements after the last elem in the chain
		lastTPost := tPosts[len(tPosts)-1]
		analysis.RemoveLater(lastTPost + 1)
		// add a replayEndElem
		analysis.AddTraceElementReplay(lastTPost+2, 0)

		fileName := filepath.Join(pkgPath, fmt.Sprintf("fuzzingTrace_%d", numberOfWrittenGoPieMuts))
		numberOfWrittenGoPieMuts++

		err = io.WriteTrace(fileName, analysis.GetNoRoutines())
		if err != nil {
			analysis.SetTrace(oldTrace)
			return fmt.Errorf("Could not create pie mutation")
		}

		mutationQueue = append(mutationQueue, mutation{mutType: mutPiType, mutPie: fileName})

		analysis.SetTrace(oldTrace)

	}

	return nil
}

func getEnergy() int {
	score := 100 // TODO: calculate this score, not in paper and not really clear from the code

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100

}
