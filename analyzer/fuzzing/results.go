// Copyright (c) 2025 Erik Kassubek
//
// File: results.go
// Brief: Handle already found results to prevent double replay
//
// Author: Erik Kassubek
// Created: 2025-01-19
//
// License: BSD-3-Clause

package fuzzing

var foundBugs = make(map[string]bool) // string from bug.Bug -> replay suc or not necessary

func AddFoundBug(bugString string, suc bool) {
	_, ok := foundBugs[bugString]
	if !ok { // first time found
		foundBugs[bugString] = suc
	} else { // already seen
		if suc {
			foundBugs[bugString] = suc
		}
	}
}

func HasAlreadyBeenReplayed(bugString string) bool {
	suc, ok := foundBugs[bugString]
	return ok && suc
}
