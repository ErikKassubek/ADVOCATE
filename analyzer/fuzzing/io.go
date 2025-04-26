// Copyright (c) 2024 Erik Kassubek
//
// File: io.go
// Brief: Functions to read ans write the fuzzing files
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package fuzzing

import (
	"analyzer/timer"
	"analyzer/trace"
	"fmt"
	"os"
	"path/filepath"
)

// Write a given mutation to a mutation file. These files are used to run the mutation
//
// Parameter:
//   - pathToFolder string: path to where the mutation should be created
//   - mut mutation: the mutation to write
//
// Returns:
//   - error
func writeMutationToFile(pathToFolder string, mut mutation) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	// write for mut and mut type, for goPie it is already written
	if mut.mutType == mutSelType || mut.mutType == mutFlowType {
		fileName := filepath.Join(pathToFolder, fmt.Sprintf("fuzzingData.log"))
		sep := "#"

		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close() // Ensure the file is closed when we're done

		// Write the string to the file
		for pos, selects := range mut.mutSel {
			content := fmt.Sprintf("%s;", pos)

			for i, sel := range selects {
				if i != 0 {
					content += ","
				}
				content += fmt.Sprintf("%d", sel.chosenCase)
			}

			content += "\n"

			_, err = file.WriteString(content)
			if err != nil {
				return err
			}
		}

		_, err = file.WriteString(sep + "\n")

		for pos, count := range mut.mutFlow {
			content := fmt.Sprintf("%s;%d\n", pos, count)
			_, err = file.WriteString(content)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Given a path, if it a dir, return the path, otherwise return the
// path to the dir the file is in
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path to the dir
func getPath(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return path
	}

	if info.IsDir() {
		return path
	}

	return filepath.Dir(path)
}

// writeMutActive writes the element in the chain into a rewriteActive.log
// file for use in GoPie (only in original, not in GoPi+ or GoPieHB)
//
// Parameter
//   - fuzzingTracePath string: path to the trace folder
//   - tr *trace.Trace: the trace to write
//   - mut *chain: chain to write
func writeMutActive(fuzzingTracePath string, tr *trace.Trace, mut *chain) {
	activePath := filepath.Join(fuzzingTracePath, "rewriteActive.log")

	f, err := os.Create(activePath)
	if err != nil {
		return
	}

	defer f.Close()

	// find the counter for all elements in the mut
	mutCounter := make(map[trace.TraceElement]int)
	for _, elem := range mut.elems {
		mutCounter[elem] = 0
	}

	traceIter := tr.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		if _, ok := mutCounter[elem]; ok { // is in chain
			mutCounter[elem]++
		}
	}

	for _, elem := range mut.elems {
		key := fmt.Sprintf("%d:%s,%d,%d\n", elem.GetRoutine(), elem.GetPos(), elem.GetTPre(), mutCounter[elem])
		f.WriteString(key)
	}
}
