// Copyright (c) 2024 Erik Kassubek
//
// File: programParts.go
// Brief: Read the program code at the positions of the bug
//
// Author: Erik Kassubek
// Created: 2024-06-17
//
// License: BSD-3-Clause

package explanation

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Get the positions of the bug elements in the program
//
// Parameter:
//   - traceElem1 map[int]string: The trace elements of the bug
//
// Returns:
//   - map[int][]string: Dict for the code snippets
func getBugPositions(traceElems map[int][]string, progInfo map[string]string) (map[int][]string, error) {
	res := make(map[int][]string)

	for i, elem := range traceElems {
		for _, e := range elem {
			pos := strings.Split(e, ":")
			file := pos[0]
			line, err := strconv.Atoi(pos[1])
			if err != nil {
				fmt.Println("Invalid line: ", pos[1])
			}

			// headerLine, _ := strconv.Atoi(progInfo["headerLine"])

			// if file == progInfo["file"] {
			// 	if line > headerLine {
			// 		line -= 5 // header and import
			// 	} else {
			// 		line-- // only import
			// 	}
			// }

			code, err := GetProgramCode(file, line, true)
			if err != nil {
				res[i] = append(res[i], "")
			} else {
				res[i] = append(res[i], code)
			}
		}
	}

	return res, nil
}

// GetProgramCode returns the code snippet of a program file at a specific line
//
// Parameter:
//   - file string: The path to the file
//   - line int: The line number
//   - numbers bool: If line numbers should be included
//
// Returns:
//   - string: The code snippet
//   - error: An error if the file could not be read
func GetProgramCode(file string, line int, numbers bool) (string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if line < 0 || line >= len(lines) {
		return "", errors.New("line number out of range")
	}

	res := "```go\n"

	start := line - 10
	if start < 0 {
		start = 0
	} else {
		res += "...\n\n"
	}
	end := line + 10
	isEnd := false
	if end >= len(lines) {
		end = len(lines)
		isEnd = true
	}

	res += strings.Join(lines[start:end], "\n")

	if !isEnd {
		res += "\n\n..."
	}
	res += "\n```"

	if !numbers {
		return res, nil
	}

	// add line numbers
	resWithLines := ""
	for i, l := range strings.Split(res, "\n") {
		if i == 0 || i == len(strings.Split(res, "\n"))-1 {
			resWithLines += l + "\n"
			continue
		}
		resWithLines += strconv.Itoa(i+start-2) + " " + l
		if i+start-2 == line {
			resWithLines += "           // <-------\n"
		} else {
			resWithLines += "\n"
		}
	}

	return resWithLines, nil
}
