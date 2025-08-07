//
// File: io.go
// Brief: Functions to read ans write the fuzzing files
//
// Created: 2024-11-28
//
// License: BSD-3-Clause

package data

import (
	"fmt"
	"goCR/utils/timer"
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
func WriteMutationToFile(pathToFolder string, mut Mutation) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	// write for mut and mut type, for goPie it is already written
	if mut.MutType == MutSelType || mut.MutType == MutFlowType {
		fileName := filepath.Join(pathToFolder, "fuzzingData.log")
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
