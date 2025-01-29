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
	"log"
	"os"
	"path/filepath"
)

func writeMutationsToFile(pathToFolder string, mut map[string][]fuzzingSelect) error {
	fileName := filepath.Join(pathToFolder, fmt.Sprintf("fuzzingData.log"))

	// Open the file for writing. If it doesn't exist, create it.
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure the file is closed when we're done

	// Write some content to the file
	log.Println("Len mut: ", len(mut))
	for id, selects := range mut {
		log.Println("Len sel: ", len(selects))
		content := fmt.Sprintf("%s;", id)

		for i, sel := range selects {
			if i != 0 {
				content += ","
			}
			content += fmt.Sprintf("%d", sel.chosenCase)
		}

		_, err = file.WriteString(content)
		if err != nil {
			return err
		}
	}

	return nil
}
