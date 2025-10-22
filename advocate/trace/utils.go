// Copyright (c) 2024 Erik Kassubek
//
// File: analysisUtil.go
// Brief: Collection of utility functiond for trace analysis
//
// Author: Erik Kassubek
// Created: 2024-05-29
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// TODO: remove tID as much as possible

// InfoFromTID returns the info from a TID
//
// Parameter:
//   - tID string: The TID
//
// Returns:
//   - string: the file
//   - int: the line
//   - int: the tPre
//   - error: the error
func InfoFromTID(tID string) (string, int, int, error) {
	spilt1 := strings.Split(tID, "@")

	if len(spilt1) != 3 {
		return "", 0, 0, errors.New(fmt.Sprint("TID not correct: no @: ", tID))
	}

	tPre, err := strconv.Atoi(spilt1[2])
	if err != nil {
		return "", 0, 0, err
	}

	// for windows test
	sp := spilt1[1]
	split3 := strings.Split(sp, ":/")
	if len(split3) == 2 {
		sp = split3[1]
	}

	split2 := strings.Split(sp, ":")
	if len(split2) < 2 {
		return "", 0, 0, errors.New(fmt.Sprint("TID not correct: no ':': ", tID))
	}

	line, err := strconv.Atoi(split2[1])
	if err != nil {
		return "", 0, 0, err
	}
	file := split2[0]

	return file, line, tPre, nil
}

// SameRoutine determines if for aal trace elements in the list, if they are
// operations on the same primitive, they have the same routine
//
// Parameter:
//   - elems ...[]TraceElement: lists of trace elements
//
// Returns:
//   - true if for each primitive, the element in elems are always in the same routine
func SameRoutine(elems ...[]Element) bool {
	ids := make(map[int]int)
	for _, elem := range elems {
		for i, e := range elem {
			if _, ok := ids[i]; !ok {
				ids[i] = e.GetRoutine()
			} else if ids[i] != e.GetRoutine() {
				return false
			}
		}
	}

	return true
}

// PosFromPosString returns the file and line from a pos string
//
// Parameter:
//   - pos string: [file]:[line]
//
// Returns:
//   - string: file
//   - int: line
//   - error
func PosFromPosString(pos string) (string, int, error) {
	posSplit := strings.Split(pos, ":")
	if len(posSplit) != 2 {
		return "", 0, fmt.Errorf("Invalid pos %s", pos)
	}

	line, err := strconv.Atoi(posSplit[1])
	if err != nil {
		return "", 0, fmt.Errorf("Invalid pos %s: %s", pos, err.Error())
	}

	return posSplit[0], line, nil
}

// PosStringFromPos returns the pos string from a file and line
//
// Parameter:
//   - string: file
//   - int: line
//
// Returns:
//   - pos string: [file]:[line]
func PosStringFromPos(file string, line int) string {
	return fmt.Sprintf("%s:%d", file, line)
}
