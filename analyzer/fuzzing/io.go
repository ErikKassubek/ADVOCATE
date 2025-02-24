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
	"fmt"
	"os"
	"path/filepath"
)

func writeMutationToFile(pathToFolder string, mut mutation) error {
	mutSel := mut.mutSel
	mutFlow := mut.mutFlow

	fileName := filepath.Join(pathToFolder, fmt.Sprintf("fuzzingData.log"))
	sep := "#"

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure the file is closed when we're done

	// Write the string to the file
	for pos, selects := range mutSel {
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

	for pos, count := range mutFlow {
		content := fmt.Sprintf("%s;%d\n", pos, count)
		_, err = file.WriteString(content)
		if err != nil {
			return err
		}
	}

	return nil
}

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
