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
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const fileBlockSep = "###########"

/*
 * read the fuzzing file and store in data
 * Args:
 *   path: path to the fuzzing file
 */
func readFile(filePath string) error {
	// If this is the first run and no fuzzing file exists yet
	if filePath == "" {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			createNewFuzzingFile(filePath)
		} else {
			println("Error opening file: " + filePath)
			panic(err)
		}
	}

	scanner := bufio.NewScanner(file)

	state := 0
	for scanner.Scan() {
		line := scanner.Text()

		// # if the block separator was found, change to the next state
		if line == fileBlockSep {
			state++
			continue
		}

		switch state {
		case 0:
			lineSplit := strings.Split(line, ";")
			if len(lineSplit) != 2 {
				return fmt.Errorf("Fuzzing File invalid format: %s", err.Error())
			}
			numberOfPreviousRuns, err = strconv.Atoi(lineSplit[0])
			if err != nil {
				return fmt.Errorf("Fuzzing File invalid format: %s", err.Error())
			}
			maxScore, err = strconv.ParseFloat(lineSplit[1], 64)
		case 1:
			err = readChannelInfo(line)
		case 2:
			err = readPairInfo(line)
		}
		if err != nil {
			return fmt.Errorf("Fuzzing File invalid format: %s", err.Error())
		}
	}

	if state != 2 {
		return fmt.Errorf("Fuzzing File invalid format: Incorrect number of blocks")
	}

	return nil
}

func createNewFuzzingFile(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	content := "0;0\n###########\n###########\n"
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

/*
 * Read the info for a channel. The line should have the following format:
 * fileCreate:lineCreate;closeInfo;qSize;maxQSize
 * closeInfo can be
 *   a: always been closed
 *   n: never been closed
 *   s: in some runs it was closed, but not in all
 * qSize is the buffer size
 * maxQSize is the maximum fullness of the buffer over all operations
 */
func readChannelInfo(line string) error {
	lineSplit := strings.Split(line, ";")
	if len(lineSplit) != 4 {
		return fmt.Errorf("Fuzzing File invalid format: %s. Invalid number of elements: %d", line, len(lineSplit))
	}

	var ci closeInfo
	switch lineSplit[1] {
	case "a":
		ci = always
	case "n":
		ci = never
	case "s":
		ci = sometimes
	default:
		return fmt.Errorf("Fuzzing File invalid format: %s: Invalid close Info %s", line, lineSplit[1])
	}

	qSize, err := strconv.Atoi(lineSplit[2])
	if err != nil {
		return fmt.Errorf("Fuzzing File invalid format: %s: Invalid qSize %s", line, lineSplit[2])
	}

	maxQSize, err := strconv.Atoi(lineSplit[3])
	if err != nil {
		return fmt.Errorf("Fuzzing File invalid format: %s: Invalid maxQSize %s", line, lineSplit[2])
	}

	addFuzzingChannel(lineSplit[0], ci, qSize, maxQSize)

	return nil
}

/*
 * Read the info for a send/receive pair
 * It must have the following form:
 * fileSend:lineSend:selCaseSend;fileRecv:lineRecv:selCaseRecv;avgNumberCom
 * selCaseSend and selCaseRecv identify the cases in a select. If the send/recv is in a select, the value is set to the number of the case. If it is not part of a select, it is set to -1.
 */
func readPairInfo(line string) error {
	lineSplit := strings.Split(line, ";")
	if len(lineSplit) != 3 {
		return fmt.Errorf("Fuzzing File invalid format: %s. Invalid number of elements: %d", line, len(lineSplit))
	}

	com, err := strconv.ParseFloat(lineSplit[2], 64)
	if err != nil {
		return fmt.Errorf("Fuzzing File invalid format: %s: Invalid com %s", line, lineSplit[2])
	}

	addFuzzingPair(lineSplit[0], lineSplit[1], com)

	return nil
}

/*
 * Write the current info to a file
 */
func writeFileInfo(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// write the number of previous runs
	file.WriteString(fmt.Sprintf("%d;$f", numberOfPreviousRuns, maxScore))

	file.WriteString(fileBlockSep)

	// write the channel block
	for _, line := range channelInfoFile {
		file.WriteString(line.toString())
	}

	file.WriteString(fileBlockSep)

	// write the operation pair block
	for _, line := range pairInfoFile {
		file.WriteString(line.toString())
	}

	return nil
}

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

/*
 * Store the results of previously found fuzzing runs.
 * This file is mainly used to prevent replay of already found bugs
 */
func storeFuzzingResults() {

}
