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

import (
	"analyzer/utils"
	"math"
)

const maxRuntimeRecordingSec = 7 * 60 // 7 min

// A run is considered interesting, if at least one of the following conditions is met
//  1. The run contains a new pair of channel operations (new meaning it has not been seen in any of the previous runs)
//  2. An operation pair's execution counter changes significantly (by at least 50%) from previous avg over all runs.
//  3. A new channel operation is triggered, such as creating, closing or not closing a channel for the first time
//  4. A buffered channel gets a larger maximum fullness than in all previous executions (MaxChBufFull)
//  5. A select case is executed for the first time
func isInterestingSelect() bool {
	// 1. The run contains a new pair of channel operations (new meaning it has not been seen in any of the previous runs)
	for keyTrace, pit := range pairInfoTrace {
		pif, ok := pairInfoFile[keyTrace]
		if !ok {
			return true
		}

		// 2. An operation pair's execution counter changes significantly from previous order.
		change := math.Abs((pit.com - pif.com) / pif.com)
		if change > 0.5 {
			return true
		}
	}

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

	if useHBInfoFuzzing {
		// 5. A select choses a case it has never been selected before
		for id, data := range selectInfoTrace {
			alreadyExecCase, ok := selectInfoFile[id]
			if !ok { // select has never been seen before
				return true
			}

			for _, sel := range data { // case has been executed for the first time
				if !utils.Contains(alreadyExecCase, sel.chosenCase) {
					return true
				}
			}
		}
	}

	return false
}
