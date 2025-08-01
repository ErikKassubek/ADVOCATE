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

var isFuzzing = false
var finishFuzzingStarted = false

// Initialize fuzzing
//
// Parameter:
//   - tracePath string: For fuzzing approaches that use trace, add the path to the
//     trace, otherwise set to ""
//   - timeout int: Timeout in seconds
func InitFuzzing(tracePath string, timeout int) {
	prefSel := make(map[string][]int)
	prefFlow := make(map[string][]int)

	InitTracing(0) // timeout will be done in startReplay

	if tracePath == "" { // GoFuzz and Flow
		fuzzingSelectPath := "fuzzingData.log"
		var err error
		prefSel, prefFlow, err = readFuzzingSelectFile(fuzzingSelectPath)
		if err != nil {
			println("Error in reading ", fuzzingSelectPath, ": ", err.Error())
			panic(err)
		}
		runtime.InitFuzzingDelay(prefSel, prefFlow, FinishFuzzing)
	} else { // GoPie
		runtime.InitFuzzingReplay(FinishFuzzing)
		tracePathRewritten = tracePath
		runtime.SetReplayAtomic(true)
		startReplay(timeout)
	}
}

// Run when fuzzing is finished (normally as defer)
// This records the traces and some additional info
func FinishFuzzing() {
	if finishFuzzingStarted {
		return
	}
	finishFuzzingStarted = true
	runtime.WaitForReplayFinish(true)
	runtime.DisableReplay()
	FinishTracing()
}

// Read the file containing the preferred select cases
//
// Parameter:
//   - pathSelect string: path to the file containing the select
//     preferred cases
//
// Returns:
//   - map[string][]int: key: file:line of select, values: list of preferred cases in select
//   - map[string]int: key: file:line of select, values: counter of operation to delay
//   - error
func readFuzzingSelectFile(pathSelect string) (map[string][]int, map[string][]int, error) {
	resSelect := make(map[string][]int)
	resFlow := make(map[string][]int)

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

		path := elems[0]

		if mode == 1 {
			ids := strings.Split(elems[1], ",")

			if len(ids) == 0 {
				continue
			}

			resSelect[path] = make([]int, len(ids))
			for i, id := range ids {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					return resSelect, resFlow, err
				}
				resSelect[path][i] = idInt
			}
		} else {
			counts := strings.Split(elems[1], ",")
			for _, count := range counts {
				countInt, err := strconv.Atoi(count)
				if err != nil {
					return resSelect, resFlow, err
				}
				if _, ok := resFlow[path]; !ok {
					resFlow[path] = make([]int, 0)
				}

				resFlow[path] = append(resFlow[path], countInt)
			}
		}
	}

	return resSelect, resFlow, nil
}
