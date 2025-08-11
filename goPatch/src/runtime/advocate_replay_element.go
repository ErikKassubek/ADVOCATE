// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_replay_element.go
// Brief: Replay element
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package runtime

// Build the key (id) of a replay element
//
// Parameters:
//   - routine int: replay routine of the element
//   - file string: code position file of the element
//   - line int: code position line of the element
func BuildReplayKey(routine int, file string, line int) string {
	return intToString(routine) + ":" + file + ":" + intToString(line)
}

// The replay data structure.
// The replay data structure is used to store the routine local trace of the replay.
//
// Fields:
// - Routine int: id of the represented routine
//   - op: identifier of the operation
//   - time: time of the operation
//   - timePre: pre time
//   - file: file in which the operation is executed
//   - line: line number of the operation
//   - blocked: true if the operation is blocked (never finised, tpost=0), false otherwise
//   - suc: success of the opeartion
//     for mutexes: trylock operations true if the lock was acquired, false otherwise
//     for other operations always true
//     for once: true if the once was chosen (was the first), false otherwise
//     for others: always true
//   - Index: Index of the select case (only for select) or index of the new routine (only for spawn), otherwise 0
type ReplayElement struct {
	Routine int
	Op      Operation
	Time    int
	TimePre int
	File    string
	Line    int
	Blocked bool
	Suc     bool
	Index   int
}

// Get the Key (id) of a replay element
//
// Returns:
//   - the Key of elem
func (elem *ReplayElement) Key() string {
	return BuildReplayKey(elem.Routine, elem.File, elem.Line)
}

// Given an replay element key, create the corresponding replay element
//
// Parameter:
//   - key (string): te replay key
//
// Returns:
//   - ReplayElement: a replay element that fits to the key
func replayElemFromKey(key string) ReplayElement {
	keySplit := split(key, ':')
	return ReplayElement{
		Routine: stringToInt(keySplit[0]),
		File:    keySplit[1],
		Line:    stringToInt(keySplit[2]),
		Suc:     true,
		Blocked: false,
	}
}
