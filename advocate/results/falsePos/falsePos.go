// Copyright (c) 2025 Erik Kassubek
//
// File: falsePos.go
// Brief: Check for false positives
//
// Author: Erik Kassubek
// Created: 2025-11-17
//
// License: BSD-3-Clause

package falsepos

import (
	"advocate/utils/helper"
	"advocate/utils/log"
	"fmt"
	"strings"
)

// IsFalsePositive checks if the given bug is likely a false positive based on
// the program code and gc based leak detection
func IsFalsePositive(resultType helper.ResultType, fileName string, line int, blocked map[string]map[int]struct{}, contextCancel map[int]struct{}, contextDone map[string]map[int]int) (bool, error) {
	// is confirmed dead by GC
	if _, ok := blocked[fileName][line]; ok {
		return false, nil
	}

	fset, file, info, err := parseFile(fileName)
	if err != nil {
		return false, err
	}

	parents := buildParentMap(file)

	// fix header shift
	lineShift := line
	if strings.HasSuffix(fileName, "main.go") || strings.HasSuffix(fileName, "_test.go") {
		lineShift -= 5
	}

	node := findNodeAtLine(fset, file, lineShift)
	if node == nil {
		return false, fmt.Errorf("Could not find node")
	}

	endlessLoop := isInsideEndlessLoop(node, parents)

	if endlessLoop {
		return true, nil
	}

	rangeOverChan, err := isRangeOverChan(node, parents, info)
	if err != nil {
		log.Errorf("Could not check for endless loop")
	}
	if rangeOverChan {
		return true, nil
	}

	contextWithCancel := isContextDoneWithCancel(node, info, fileName, line, contextCancel, contextDone)
	if contextWithCancel {
		return true, nil
	}

	return false, nil
}
