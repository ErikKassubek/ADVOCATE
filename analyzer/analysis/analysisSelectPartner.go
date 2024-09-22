// Copyrigth (c) 2024 Erik Kassubek
//
// File: analysisSelectPartner.go
// Brief: Trace analysis for detection of select cases without any possible partners
//
// Author: Erik Kassubek <kassubek.erik@gmail.com>
// Created: 2024-03-04
// LastChange: 2024-09-01
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/logging"
)

/*
* CheckForSelectCaseWithoutPartner checks for select cases without a valid
* partner. Call when all elements have been processed.
 */
func CheckForSelectCaseWithoutPartner() {
	// check if not selected cases could be partners
	for i, c1 := range selectCases {
		for j := i + 1; j < len(selectCases); j++ {
			c2 := selectCases[j]

			if c1.partner && c2.partner {
				continue
			}

			if c1.chanID != c2.chanID || c1.vcTID.TID == c2.vcTID.TID || c1.send == c2.send {
				continue
			}

			if c2.send { // c1 should be send, c2 should be recv
				c1, c2 = c2, c1
			}

			hb := clock.GetHappensBefore(c1.vcTID.Vc, c2.vcTID.Vc)
			found := false
			if c1.buffered && (hb == clock.Concurrent || hb == clock.After) {
				found = true
			} else if !c1.buffered && hb == clock.Concurrent {
				found = true
			}

			if found {
				selectCases[i].partner = true
				selectCases[j].partner = true
			}
		}
	}

	if len(selectCases) == 0 {
		return
	}

	// collect all cases with no partner
	casesWithoutPartner := make(map[string][]logging.ResultElem) // tID -> cases
	casesWithoutPartnerInfo := make(map[string][]int)            // tID -> [routine, selectID]

	for _, c := range selectCases {
		if c.partner {
			continue
		}

		opjType := "C"
		if c.send {
			opjType += "S"
		} else {
			opjType += "R"
		}

		arg2 := logging.SelectCaseResult{
			SelID:   c.selectID,
			ObjID:   c.chanID,
			ObjType: opjType,
			Routine: c.vcTID.Routine,
		}

		if _, ok := casesWithoutPartner[c.vcTID.TID]; !ok {
			casesWithoutPartner[c.vcTID.TID] = make([]logging.ResultElem, 0)
			casesWithoutPartnerInfo[c.vcTID.TID] = []int{c.vcTID.Routine, c.selectID}
		}

		casesWithoutPartner[c.vcTID.TID] = append(casesWithoutPartner[c.vcTID.TID], arg2)
	}

	for tID, cases := range casesWithoutPartner {
		if len(cases) == 0 {
			continue
		}

		info := casesWithoutPartnerInfo[tID]
		if len(info) != 2 {
			logging.Debug("info should have 2 elements", logging.ERROR)
			continue
		}

		file, line, tPre, err := infoFromTID(tID)
		if err != nil {
			logging.Debug(err.Error(), logging.ERROR)
			continue
		}

		arg1 := logging.TraceElementResult{
			RoutineID: info[0],
			ObjID:     info[1],
			TPre:      tPre,
			ObjType:   "SS",
			File:      file,
			Line:      line,
		}

		logging.Result(logging.WARNING, logging.ASelCaseWithoutPartner,
			"select", []logging.ResultElem{arg1}, "case", cases)
	}
}

/*
* CheckForSelectCaseWithoutPartnerSelect checks for select cases without a valid
* partner. Call whenever a select is processed.
* Args:
*   se (*TraceElementSelect): The trace element
*   ids ([]int): The ids of the channels
*   bufferedInfo ([]bool): The buffer status of the channels
*   sendInfo ([]bool): The send status of the channels
*   vc (VectorClock): The vector clock
 */
//  func CheckForSelectCaseWithoutPartnerSelect(routine int, selectID int, caseChanIds []int, bufferedInfo []bool,
func CheckForSelectCaseWithoutPartnerSelect(se *TraceElementSelect, caseChanIds []int, bufferedInfo []bool,
	sendInfo []bool, vc clock.VectorClock) {
	for i, id := range caseChanIds {
		buffered := bufferedInfo[i]
		send := sendInfo[i]

		found := false

		if i == se.chosenIndex {
			// no need to check if the channel is the chosen case
			found = true
		} else {
			// not select cases
			if send {
				for _, mrr := range mostRecentReceive {
					if possiblePartner, ok := mrr[id]; ok {
						hb := clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hb == clock.Concurrent || hb == clock.Before) {
							found = true
							break
						} else if !buffered && hb == clock.Concurrent {
							found = true
							break
						}
					}
				}
			} else { // recv
				for _, mrs := range mostRecentSend {
					if possiblePartner, ok := mrs[id]; ok {
						hb := clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hb == clock.Concurrent || hb == clock.After) {
							found = true
						} else if !buffered && hb == clock.Concurrent {
							found = true
						}
					}
				}
			}
		}

		selectCases = append(selectCases,
			allSelectCase{se.id, id, VectorClockTID{vc, se.tID, se.routine}, send, buffered, found})

	}
}

/*
* CheckForSelectCaseWithoutPartnerChannel checks for select cases without a valid
* partner. Call whenever a channel operation is processed.
* Args:
*   id (int): The id of the channel
*   vc (VectorClock): The vector clock
*   tID (string): The position of the channel operation in the program
*   send (bool): True if the operation is a send
*   buffered (bool): True if the channel is buffered
*   sel (bool): True if the operation is part of a select statement
 */
func CheckForSelectCaseWithoutPartnerChannel(id int, vc clock.VectorClock, tID string,
	send bool, buffered bool) {

	for i, c := range selectCases {
		if c.partner || c.chanID != id || c.send == send || c.vcTID.TID == tID {
			continue
		}

		hb := clock.GetHappensBefore(vc, c.vcTID.Vc)
		found := false
		if send {
			if buffered && (hb == clock.Concurrent || hb == clock.Before) {
				found = true
			} else if !buffered && hb == clock.Concurrent {
				found = true
			}
		} else {
			if buffered && (hb == clock.Concurrent || hb == clock.After) {
				found = true
			} else if !buffered && hb == clock.Concurrent {
				found = true
			}
		}

		if found {
			selectCases[i].partner = true
		}
	}
}

/*
* CheckForSelectCaseWithoutPartnerClose checks for select cases without a valid
* partner. Call whenever a close operation is processed.
* Args:
*   id (int): The id of the channel
*   vc (VectorClock): The vector clock
 */
func CheckForSelectCaseWithoutPartnerClose(id int, vc clock.VectorClock) {
	for i, c := range selectCases {
		if c.partner || c.chanID != id || c.send {
			continue
		}

		hb := clock.GetHappensBefore(vc, c.vcTID.Vc)
		found := false
		if c.buffered && (hb == clock.Concurrent || hb == clock.After) {
			found = true
		} else if !c.buffered && hb == clock.Concurrent {
			found = true
		}

		if found {
			selectCases[i].partner = true
		}
	}
}
