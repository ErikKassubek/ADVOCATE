//
// File: data.go
// Brief: Data for flow
//
// Created: 2025-07-03
//
// License: BSD-3-Clause

package flow

const maxFlowMut = 10

var (
	alreadyDelayedElems = make(map[string][]int)
)

// ClearData deletes the data of flow mutation
func ClearData() {
	alreadyDelayedElems = make(map[string][]int)
}
