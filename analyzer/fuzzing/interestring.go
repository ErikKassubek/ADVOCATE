// Copyright (c) 2024 Erik Kassubek
//
// File: interestring.go
// Brief: Functions to determine whether a run was interesting
//
// Author: Erik Kassubek
// Created: 2024-11-28
//
// License: BSD-3-Clause

package fuzzing

import "log"

/*
 * A run is considered interesting, if at least one of the following conditions is met
 * 	1. The run contains a new pair of channel operations (new meaning it has not been seen in any of the previous runs)
 * 	2. An operation pair's execution counter changes significantly from previous order.
 * 	3. A new channel operation is triggered, such as creating, closing or not closing a channel for the first time
 * 	4. A buffered channel gets a larger maximum fullness than in all previous executions (MaxChBufFull)
 */
func isInteresting() bool {
	// 1. The run contains a new pair of channel operations (new meaning it has not been seen in any of the previous runs)
	log.Println("PairInfoTrace: ", len(pairInfoTrace))
	for keyTrace, _ := range pairInfoTrace {
		if _, ok := pairInfoFile[keyTrace]; !ok {
			return true
		}
	}

	// 2. An operation pair's execution counter changes significantly from previous order.
	// TODO: implement

	for _, data := range channelInfoTrace {
		fileData, ok := channelInfoFile[data.globalID]

		// 3. A new channel operation is triggered, such as creating, closing or not closing a channel for the first time
		// never created before
		if !ok {
			return true
		}
		// first time closed
		if data.closeInfo == always && fileData.closeInfo == never {
			return true
		}
		// first time not closed
		if data.closeInfo == never && fileData.closeInfo == always {
			return true
		}

		// 4. A buffered channel gets a larger maximum fullness than in all previous executions (MaxChBufFull)
		if data.maxQCount > fileData.maxQCount {
			return true
		}
	}

	return false
}
