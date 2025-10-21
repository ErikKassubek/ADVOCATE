// Copyright (c) 2024 Erik Kassubek
//
// File: io.go
// Brief: Functions to read ans write the fuzzing files
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package fuzzingHelper

import (
	"advocate/fuzzing/baseF"
	"advocate/trace"
	"advocate/utils/paths"
	"advocate/utils/timer"
	"fmt"
	"os"
	"path/filepath"
)

// WriteMutationToFile writes a given mutation to a mutation file.
// These files are used to run the mutation
//
// Parameter:
//   - pathToFolder string: path to where the mutation should be created
//   - mut mutation: the mutation to write
//
// Returns:
//   - error
func WriteMutationToFile(pathToFolder string, mut baseF.Mutation) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	// write for mut and mut type, for goPie it is already written
	if mut.MutType == baseF.MutSelType || mut.MutType == baseF.MutFlowType {
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
func WriteMutActive(fuzzingTracePath string, tr *trace.Trace, mut *baseF.Chain, partTime int) {
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
