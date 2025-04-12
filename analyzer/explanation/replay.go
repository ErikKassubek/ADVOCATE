// Copyright (c) 2024 Erik Kassubek
//
// File: replay.go
// Brief: Read the info about the rewrite and replay of the bug
//
// Author: Erik Kassubek
// Created: 2024-06-18
//
// License: BSD-3-Clause

package explanation

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Get the text for the rewrite and replay info for the explanation
//
// Parameter:
//   - bugType (string): bug code
//   - code (map[string]string): replay output codes
//   - index (int): index of the bug that is explained
//
// Returns:
//   - map[string]string: Information for the rewrite/replay part of the explanation file
func getRewriteInfo(bugType string, codes map[string]string, index string) map[string]string {
	res := make(map[string]string)

	rewType := getRewriteType(bugType)

	res["description"] = ""
	res["exitCode"] = ""
	res["exitCodeExplanation"] = ""
	res["replaySuc"] = "was not possible"

	var err error

	if rewType == "Actual" {
		res["description"] += "The bug is an actual bug. Therefore no rewrite is possibel."
		codes[fmt.Sprint(index)] = "fail"
	} else if rewType == "Possible" {
		res["description"] += "The bug is a potential bug.\n"
		res["description"] += "The analyzer has tries to rewrite the trace in such a way, "
		res["description"] += "that the bug will be triggered when replaying the trace."
	} else if rewType == "LeakPos" {
		res["description"] += "The analyzer found a leak in the recorded trace.\n"
		res["description"] += "The analyzer found a way to resolve the leak, meaning the "
		res["description"] += "leak should not reappear in the rewritten trace."
	} else if rewType == "Leak" {
		res["description"] += "The analyzer found a leak in the recorded trace.\n"
		res["description"] += "The analyzer could not find a way to resolve the leak. "
		res["description"] += "No rewritten trace was created. This does not need to mean, "
		res["description"] += "that the leak can not be resolved, especially because the "
		res["description"] += "analyzer is only aware of executed operations."
		codes[fmt.Sprint(index)] = "fail"
	}
	res["exitCode"], res["exitCodeExplanation"], res["replaySuc"], err = getReplayInfo(codes, index)

	if err != nil {
		fmt.Println("Error getting replay info: ", err)
	}

	return res

}

// From the bug code get wether the bug was actual, possible or a (possible) leak
//
// Parameter:
//   - bugCode (string): the bug code
//
// Returns:
//   - string: Actual, Possible, Leak or Pos Leak
func getRewriteType(bugCode string) string {
	switch bugCode[:1] {
	case "A":
		return "Actual"
	case "P":
		return "Possible"
	case "L":
		res := "Leak"
		if bugCode == "L01" || bugCode == "L03" || bugCode == "L06" ||
			bugCode == "L08" || bugCode == "L09" || bugCode == "L10" {
			res += "Pos"
		}
		return res
	}
	return ""
}

// Get the output codes from the output.log file
//
// Parameter:
//   - path (string): path to the folder containing the output.log file
//
// Returns: map[string]stringL exit codes
func getOutputCodes(path string) map[string]string {
	output := filepath.Join(path, "output.log")
	if _, err := os.Stat(output); os.IsNotExist(err) {
		res := "No replay info available. Output.log does not exist."
		return map[string]string{"AdvocateFailExplanationInfo": res, "AdvocateFailResplaySucInfo": "information not available"}
	}

	// read the output file
	content, err := os.ReadFile(output)
	if err != nil {
		res := "No replay info available. Could not read output.log file"
		return map[string]string{"AdvocateFailExplanationInfo": res, "AdvocateFailResplaySucInfo": "information not available"}
	}

	lines := strings.Split(string(content), "\n")

	replayPos := make(map[string]bool)
	replayCode := make(map[string]string)
	bugrepPrefix := "Bugreport info: "
	replayReadPrefix := "Reading trace from rewrittenTrace_"
	exitCodePrefix := "Exit Replay with code"

	lastReplayIndex := ""
	lastReplayIndexInfoFound := true

	for _, line := range lines {
		if strings.HasPrefix(line, bugrepPrefix) {
			line = strings.TrimPrefix(line, bugrepPrefix)
			lineSplit := strings.Split(line, ",")
			if len(lineSplit) == 2 {
				index := lineSplit[0]
				if lineSplit[1] == "suc" {
					replayPos[index] = true
				} else {
					replayPos[index] = false
				}
				if lineSplit[1] == "double" {
					replayCode[index] = "double"
				}
				if lineSplit[1] == "fail" {
					replayCode[index] = "fail"
				}
			}
		} else if strings.HasPrefix(line, replayReadPrefix) {
			if !lastReplayIndexInfoFound {
				replayCode[lastReplayIndex] = "panic"
			}
			lastReplayIndex = strings.TrimPrefix(line, replayReadPrefix)
			// if !strings.Contains(lastReplayIndex, "_") {
			// 	lastReplayIndex = "0_" + lastReplayIndex
			// }
			lastReplayIndexInfoFound = false
		} else if strings.HasPrefix(line, exitCodePrefix) {
			line = strings.TrimPrefix(line, exitCodePrefix)
			line = strings.TrimSpace(line)
			replayCode[lastReplayIndex] = strings.Split(line, " ")[0]
			lastReplayIndexInfoFound = true
		}
	}

	return replayCode
}

// Get the text for the replay info for the explanation
//
// Parameter:
//   - code (map[string]string): replay output codes
//   - index (int): index of the bug that is explained
//
// Returns:
//   - string: exit code
//   - string: replay result explanation
//   - string: replay success info
//   - error
func getReplayInfo(codes map[string]string, index string) (string, string, string, error) {
	if _, ok := codes["AdvocateFailExplanationInfo"]; ok {
		fmt.Println("Could not read")
		return "", codes["AdvocateFailExplanationInfo"], codes["AdvocateFailResplaySucInfo"], fmt.Errorf("Could not read output file")
	}

	exitCode := codes[index]
	replaySuc := "failed"
	if exitCode == "double" {
		replaySuc = "was already performed for this bug in another test"
		return "double", "", replaySuc, nil
	}
	if exitCode == "fail" {
		return "fail", exitCodeExplanation["fail"], "was not run", nil
	}
	if exitCode == "panic" {
		return "panic", exitCodeExplanation["panic"], "was terminated unexpectedly", nil
	}

	exitCodeInt, err := strconv.Atoi(exitCode)
	if err != nil {
		return "fail", exitCodeExplanation["fail"], "was not run", nil
	}
	if exitCodeInt == 0 {
		replaySuc = "ended without confirming the bug"
	} else if exitCodeInt >= 20 {
		replaySuc = "was successful"
	}

	return exitCode, exitCodeExplanation[exitCode], replaySuc, nil
}
