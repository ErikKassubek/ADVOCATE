// Copyright (c) 2024 Erik Kassubek
//
// File: io.go
// Brief: Functions to read ans write the fuzzing files
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package baseF

import (
	"advocate/analysis/baseA"
	"advocate/io"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/settings.go"
	"advocate/utils/timer"
	"fmt"
	"os"
	"path/filepath"
)

// WriteMutChain writes a chain based mutation the mutation to file and add it to the queue
//
// Parameter:
//   - mut Chain: the mutation to write
//   - first bool: set to true, if it is the first mutation for a given test/prog
//
// Returns:
//   - bool: true if max number muts in reached
//   - error
func WriteMutChain(mut Chain, first bool) (bool, error) {
	if MaxNumberRuns != -1 && NumberWrittenMutations > MaxNumberRuns {
		return true, nil
	}
	NumberWrittenMutations++

	traceCopy, err := baseA.CopyMainTrace()
	if err != nil {
		return false, err
	}

	t1 := -1
	for _, elem := range mut.Elems {
		tPost := elem.GetTPost()
		if t1 == -1 || tPost < t1 {
			t1 = tPost
		}
	}

	// remove all elements after the first elem in the chain
	traceCopy.ShortenTrace(t1, false)

	// add in all the elements in the chain
	mapping := make(map[string]trace.Element)
	for i, elem := range mut.Elems {
		c := elem.Copy(mapping)
		c.SetTSort(t1 + i*2)
		traceCopy.AddElement(c)
	}

	if first {
		AddFuzzingTraceFolder(paths.FuzzingTraces)
	}

	fuzzingTracePath := filepath.Join(paths.FuzzingTraces, fmt.Sprintf("fuzzingTrace_%d", NumberWrittenMutations))
	ChainFiles[NumberWrittenMutations] = mut

	err = io.WriteTrace(&traceCopy, fuzzingTracePath, true)
	if err != nil {
		return false, fmt.Errorf("Could not create mutation: %s", err.Error())
	}

	// write the active map to a "replay_active.log"
	if flags.FuzzingMode == GoPie || settings.WithoutReplay {
		WriteMutActive(fuzzingTracePath, &traceCopy, &mut, 0)
	} else {
		WriteMutActive(fuzzingTracePath, &traceCopy, &mut, mut.ElemWithSmallestTPost().GetTPost())
	}

	traceCopy.Clear()

	muta := Mutation{MutType: MutPiType, MutPie: NumberWrittenMutations}

	AddMutToQueue(muta, false, false)

	return false, nil
}

// Create the folder for the fuzzing traces
//
// Parameter:
//   - path string: path to the folder
func AddFuzzingTraceFolder(path string) {
	os.RemoveAll(path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Error("Could not create fuzzing folder")
	}
}

// WriteMutationToFile writes a given mutation to a mutation file.
// These files are used to run the mutation
//
// Parameter:
//   - pathToFolder string: path to where the mutation should be created
//   - mut mutation: the mutation to write
//
// Returns:
//   - error
func WriteMutationToFile(pathToFolder string, mut Mutation) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	// write for mut and mut type, for goPie it is already written
	if mut.MutType == MutSelType || mut.MutType == MutFlowType {
		fileName := filepath.Join(pathToFolder, paths.NameFuzzingData)
		sep := "#"

		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close() // Ensure the file is closed when we're done

		// Write the string to the file
		for pos, selects := range mut.MutSel {
			content := fmt.Sprintf("%s;", pos)

			for i, sel := range selects {
				if i != 0 {
					content += ","
				}
				content += fmt.Sprintf("%d", sel.ChosenCase)
			}

			content += "\n"

			_, err = file.WriteString(content)
			if err != nil {
				return err
			}
		}

		_, err = file.WriteString(sep + "\n")

		for pos, count := range mut.MutFlow {
			content := fmt.Sprintf("%s;%d\n", pos, count)
			_, err = file.WriteString(content)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetPath returns a path for a path
// Given a path, if it a dir, return the path, otherwise return the
// path to the dir the file is in
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path to the dir
func GetPath(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return path
	}

	if info.IsDir() {
		return path
	}

	return filepath.Dir(path)
}

// WriteMutActive writes the element in the chain into a rewriteActive.log
// file for use in GoPie and Guided
//
// Parameter
//   - fuzzingTracePath string: path to the trace folder
//   - tr *trace.Trace: the trace to write
//   - mut *chain: chain to write
//   - partTime int: if 0, the replay will partial replay from the beginning
//     otherwise it will switch to partial replay when the element with this
//     time is the next element to be replayed
func WriteMutActive(fuzzingTracePath string, tr *trace.Trace, mut *Chain, partTime int) {
	activePath := filepath.Join(fuzzingTracePath, paths.NameReplayActive)

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
