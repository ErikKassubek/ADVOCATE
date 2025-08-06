// Copyright (c) 2024 Erik Kassubek
//
// File: explanation.go
// Brief: Create an explanation file for a found bug
//
// Author: Erik Kassubek
// Created: 2024-06-14
//
// License: BSD-3-Clause

package explanation

import (
	"advocate/utils/log"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type bug struct {
	bugType string
	pos     string
}

var printedExplanations = make(map[bug]struct{})

func writeBug(bugType string, positions map[int][]string) bool {
	posSet := make(map[string]struct{})
	for _, p := range positions {
		for _, l := range p {
			posSet[l] = struct{}{}
		}
	}

	unique := make([]string, 0, len(posSet))
	for s := range posSet {
		unique = append(unique, s)
	}
	sort.Strings(unique)

	// Join into single string
	pos := strings.Join(unique, ",")

	b := bug{
		bugType: bugType,
		pos:     pos,
	}

	if _, ok := printedExplanations[b]; ok {
		return false
	}
	printedExplanations[b] = struct{}{}
	return true
}

// create an overview over an analyzed, and if possible replayed
// bug. It is mostly meant to give an explanation of a found
// bug to people, who are not used to the internal structure an
// representation of the analyzer.

// It creates one file. This file has the following element:
// - The type of bug found
// - The test/program, where the bug was found
// - if possible, the command to run the program
// - if possible, the command to replay the bug
// - position of the bug elements
// - code of the bug elements in the trace (+- 10 lines)
// - info about replay (was it possible or not)

// CreateOverview creates an overview over a bug found by the analyzer.
// It reads the results of the analysis, the code of the bug elements and the replay info.
// It then writes all this information into a file.
//
// Parameter:
//   - path: the path to the folder, where the results of the analysis and the trace are stored
//   - index: the index of the bug in the results
//   - ignoreDouble: if true, only write one bug report for each bug
//
// Returns:
//   - error: if an error occurred
func CreateOverview(path string, ignoreDouble bool, fuzzing int) error {
	// get the code info (main file, test name, commands)
	log.Info("Create bug reports")

	// remove possible partner information if not searched for
	if fuzzing >= 0 {
		adaptExplanationMaps()
	}
	buildBugCodes()

	replayCodes := getOutputCodes(path)

	progInfo, err := readProgInfo(path)
	if err != nil {
		log.Error("Error reading prog info: ", err)
	}

	hl, err := strconv.Atoi(progInfo["headerLine"])
	if err != nil {
		log.Error("Cound not read header line: ", err)
	}

	resultsMachine, _ := filepath.Glob(filepath.Join(path, "results_machine_*.log"))
	resultsMachine = append(resultsMachine, filepath.Join(path, "results_machine.log"))

	for _, result := range resultsMachine {
		file, _ := os.ReadFile(result)
		numberResults := len(strings.Split(string(file), "\n"))

		// timeoutFound := false

		for index := 1; index < numberResults; index++ {
			id := ""
			if strings.HasSuffix(result, "results_machine.log") {
				id += strconv.Itoa(index)
			} else {
				elem := strings.Split(strings.Split(result, ".log")[0], "_")
				id += elem[len(elem)-1] + "_" + strconv.Itoa(index)
			}

			bugType, bugPos, bugElemType, err := readAnalysisResults(result, index, progInfo["file"], hl)
			if err != nil {
				continue
			}

			if strings.HasPrefix(bugType, "S") {
				break
			}

			// get the bug type description
			bugTypeDescription := getBugTypeDescription(bugType)

			// get the code of the bug elements
			code, err := getBugPositions(bugPos, progInfo)
			if err != nil {
				log.Error("Error getting bug positions: ", err)
			}

			// get the replay info
			replay := getRewriteInfo(bugType, replayCodes, id)

			if ignoreDouble && replay["exitCode"] == "double" {
				continue
			}

			if !writeBug(bugType, bugPos) {
				continue
			}

			err = writeFile(path, id, bugTypeDescription, bugPos, bugElemType, code,
				replay, progInfo, fuzzing)
		}
	}

	return err

}

// Read an result machine file file and get one result
//
// Parameter:
//   - path string: path to the result file
//   - index int: index of the relevant bug in the file
//   - fileWithHeader string: file that contains the header for the recording
//   - headerLine int: line number of the first line of the header
//
// Returns:
//   - string: bug type
//   - map[int][]string: bug element positions
//   - map[int]string: bug element types
//   - error
func readAnalysisResults(path string, index int, fileWithHeader string, headerLine int) (string, map[int][]string, map[int]string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", nil, nil, err
	}

	lines := strings.Split(string(file), "\n")

	index-- // the index is 1-based

	if index >= len(lines) {
		return "", nil, nil, errors.New("index out of range")
	}

	bugStr := string(lines[index])
	bugFields := strings.Split(bugStr, ",")
	bugType := bugFields[0]

	bugPos := make(map[int][]string)
	bugElemType := make(map[int]string)

	posAlreadyKnown := make([]string, 0)

	for i := 1; i < len(bugFields); i++ {
		bugElems := strings.Split(bugFields[i], ";")
		if len(bugElems) == 0 {
			continue
		}

		bugPos[i] = make([]string, 0)

		for j, elem := range bugElems {
			fields := strings.Split(elem, ":")

			if fields[0] != "T" {
				continue
			}

			if j == 0 {
				bugElemType[i] = getBugElementType(fields[4])
			}

			file := fields[5]
			line := fields[6]

			// correct the line number, if the file is the main file of the program
			// because of the inserted preamble
			if file == fileWithHeader {
				lineInt, _ := strconv.Atoi(line)
				if lineInt >= headerLine {
					line = fmt.Sprint(lineInt - 5) // import + header
				} else {
					line = fmt.Sprint(lineInt - 1) // only import
				}
			}

			pos := file + ":" + line

			if slices.Contains(posAlreadyKnown, pos) {
				continue
			}
			posAlreadyKnown = append(posAlreadyKnown, pos)

			bugPos[i] = append(bugPos[i], pos)
		}
	}

	return bugType, bugPos, bugElemType, nil
}

// writeFile(path, id, bugTypeDescription, bugPos, bugElemType, code,
// 	replay, progInfo, fuzzing)

// Write an bug explanation file
//
// Parameter:
//   - path string: path where the explanation file should be created
//   - index int: index of the bug file (name is e.g. bug_[index])
//   - description map[string]string: description of the bug, containing e.g. bug/diagnostics, name, explanation
//   - positions map[int][]string: positions of the bug elements
//   - bugElemType map[int]string: types of the bug elements
//   - code map[int][]string: program codes that contains the bug elements
//   - replay map[string]string: information about the replay
//   - progInfo map[string]sting: Info about the prog, e.g. prog/test name
//   - fuzzing int: Fuzzing run number
//
// Returns:
//   - error
func writeFile(path string, index string, description map[string]string,
	positions map[int][]string, bugElemType map[int]string, code map[int][]string,
	replay map[string]string, progInfo map[string]string, fuzzing int) error {

	// write the bug type description
	res := "# " + description["crit"] + ": " + description["name"] + "\n\n"
	res += description["explanation"] + "\n\n"

	// write the positions of the bug
	res += "## Test/Program\n"
	res += "The bug was found in the following test/program:\n\n"
	if progInfo["name"] != "" {
		res += "- Test/Prog: " + progInfo["name"] + "\n"
	} else {
		res += "- Test: unknown" + "\n"
	}

	if progInfo["file"] != "" {
		res += "- File: " + progInfo["file"] + "\n\n"
	} else {
		res += "- File: unknown" + "\n\n"
	}

	// write the code of the bug elements
	res += "## Bug Elements\n"
	res += "The elements involved in the found "
	res += strings.ToLower(description["crit"])
	res += " are located at the following positions:\n\n"

	for key := range positions {
		res += "###  "
		res += bugElemType[key] + "\n"

		for j, pos := range positions[key] {
			if pos == ":-1" {
				return nil
			}
			code := code[key][j]
			res += "-> " + pos + "\n"
			res += code + "\n\n"
		}
	}

	// write the info about the replay, if possible including the command to read the bug
	replayPossible := replay["replaySuc"] != "was not possible" && replay["replaySuc"] != "was not run"
	replayDouble := replay["exitCode"] == "double"

	res += "## Replay\n"
	if replayPossible && !replayDouble {
		res += replay["description"] + "\n\n"
	}

	if replayDouble {
		res += "The replay was not performed, because the same bug had been found before."
	} else {
		res += "**Replaying " + replay["replaySuc"] + "**.\n\n"
		if replayPossible {
			// res += "The replayed trace can be found in: "
			// res += "rewrittenTrace_" + index + "\n\n"
			if replay["replaySuc"] == "panicked" {
				res += "It panicked with the following message:\n\n"
				res += replay["exitCode"] + "\n\n"
			} else if replay["exitCode"] == "fail" {
				res += replay["exitCodeExplanation"] + "\n\n"
			} else {
				res += "It exited with the following code: "
				res += replay["exitCode"] + "\n\n"
				res += replay["exitCodeExplanation"] + "\n\n"
			}
		}
	}

	confirmed := false
	if description["crit"] == "Bug" {
		if replayDouble || replay["replaySuc"] == "was successful" ||
			strings.HasPrefix(description["name"], "Actual") {
			confirmed = true
		}
	}

	id := progInfo["file"] + "#" + progInfo["name"]
	if replay["replaySuc"] == "was not run" {
		log.Resultf(true, confirmed, id, "Found %s.", description["name"])
	} else {
		log.Resultf(true, confirmed, id, "Found %s. Replay %s.", description["name"], replay["replaySuc"])
	}

	// if in path, the folder "bugs" does not exist, create it
	if _, err := os.Stat(path + "/bugs"); os.IsNotExist(err) {
		err := os.Mkdir(path+"/bugs", 0755)
		if err != nil {
			return err
		}
	}

	folderName := path + "/bugs"
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		err := os.Mkdir(folderName, 0755)
		if err != nil {
			return err
		}
	}

	// create the file
	fileName := ""
	if fuzzing == -1 {
		fileName = filepath.Join(folderName, strings.ToLower(description["crit"])) + "_" + index + ".md"
	} else {
		fileName = filepath.Join(folderName, fmt.Sprintf("%s_%d_%s.md", strings.ToLower(description["crit"]), fuzzing, index))
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	_, err = file.WriteString(res)
	return err

}
