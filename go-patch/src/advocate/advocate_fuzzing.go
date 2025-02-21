// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_fuzzing.go
// Brief: Fuzzing
//
// Author: Erik Kassubek
// Created: 2024-12-10
//
// License: BSD-3-Clause

package advocate

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

/*
 * Initialize fuzzing
 * Args:
 * 	progName (string): name of the prog/test used to create the fuzzing file.
 * 		for test it must have the form progName_testName
 */
func InitFuzzing() {
	fuzzingSelectPath := "fuzzingData.log"
	prefSel, prefFlow, err := readFile(fuzzingSelectPath)

	if err != nil {
		println("Error in reading ", fuzzingSelectPath, ": ", err.Error())
		panic(err)
	}

	runtime.InitFuzzing(prefSel, prefFlow)
	InitTracing()
}

/*
 * Read the file containing the preferred select cases
 * Args:
 * 	pathSelect (string): path to the file containing the select
 * 		preferred cases
 * Returns:
 * 	map[string][]int: key: file:line of select, values: list of preferred cases in select
* 	map[string]int: key: file:line of select, values: counter of operation to delay
 * 	error
*/
func readFile(pathSelect string) (map[string][]int, map[string]int, error) {
	resSelect := make(map[string][]int)
	resFlow := make(map[string]int)

	file, err := os.Open(pathSelect)
	if err != nil {
		return resSelect, resFlow, err
	}

	mode := 1

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if line == "#" {
			mode = 2
			continue
		}

		elems := strings.Split(line, ";")
		if len(elems) != 2 {
			return resSelect, resFlow, fmt.Errorf("Incorrect line in fuzzing select file: %s", line)
		}

		if mode == 1 {
			ids := strings.Split(elems[1], ",")

			if len(ids) == 0 {
				continue
			}

			resSelect[elems[0]] = make([]int, len(ids))
			for i, id := range ids {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					return resSelect, resFlow, err
				}
				resSelect[elems[0]][i] = idInt
			}
		} else {
			count, err := strconv.Atoi(elems[1])
			if err != nil {
				return resSelect, resFlow, err
			}
			resFlow[elems[0]] = count
		}
	}

	return resSelect, resFlow, nil
}
