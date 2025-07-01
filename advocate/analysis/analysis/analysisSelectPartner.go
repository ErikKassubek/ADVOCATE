// Copyright (c) 2024 Erik Kassubek
//
// File: analysisSelectPartner.go
// Brief: Trace analysis for detection of select cases without any possible partners
//
// Author: Erik Kassubek
// Created: 2024-03-04
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/clock"
	"advocate/trace"
	"advocate/utils/timer"
)

// CheckForSelectCaseWithPartner checks for select cases with a valid
// partner. Call when all elements have been processed.
func CheckForSelectCaseWithPartner() {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	// check if not selected cases could be partners
	for i, c1 := range selectCases {
		for j := i + 1; j < len(selectCases); j++ {
			c2 := selectCases[j]

			// if c1.partnerFound && c2.partnerFound {
			// 	continue
			// }

			if c1.chanID != c2.chanID || c1.elem.elem.GetTID() == c2.elem.elem.GetTID() || c1.send == c2.send {
				continue
			}

			if c2.send { // c1 should be send, c2 should be recv
				c1, c2 = c2, c1
			}

			hb := clock.GetHappensBefore(c1.elem.vc, c2.elem.vc)
			found := false
			if c1.buffered && (hb == clock.Concurrent || hb == clock.After) {
				found = true
			} else if !c1.buffered && hb == clock.Concurrent {
				found = true
			}

			if found {
				selectCases[i].partnerFound = true
				selectCases[j].partnerFound = true
				selectCases[i].partner = append(selectCases[i].partner, ElemWithVcVal{selectCases[j].sel, selectCases[j].sel.GetVC(), 0})
				selectCases[j].partner = append(selectCases[j].partner, ElemWithVcVal{selectCases[i].sel, selectCases[i].sel.GetVC(), 0})
			}
		}
	}

	if len(selectCases) == 0 {
		return
	}

	// collect all cases with no partner and all not triggered cases with partner

	for _, c := range selectCases {
		opjType := "C"
		if c.send {
			opjType += "S"
		} else {
			opjType += "R"
		}

		if c.partnerFound {
			c.sel.AddCasesWithPosPartner(c.casi)
			numberSelectCasesWithPartner++
		}
	}
}

// CheckForSelectCaseWithPartnerSelect checks for select cases with a valid
// partner. Call whenever a select is processed.
//
// Parameter:
//   - se *TraceElementSelect: The trace elemen
//   - vc *VectorClock: The vector clock
func CheckForSelectCaseWithPartnerSelect(se *trace.TraceElementSelect, vc *clock.VectorClock) {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for casi, c := range se.GetCases() {

		id := c.GetID()

		buffered := (c.GetQSize() > 0)
		send := (c.GetOpC() == trace.SendOp)

		found := false
		executed := false
		var partner = make([]ElemWithVcVal, 0)

		if casi == se.GetChosenIndex() && se.GetTPost() != 0 {
			// no need to check if the channel is the chosen case
			executed = true
			p := se.GetPartner()
			if p != nil {
				found = true
				vcTID := ElemWithVcVal{
					p, p.GetVC().Copy(), 0,
				}
				partner = append(partner, vcTID)
			}
		} else {
			// not select cases
			if send {
				for _, mrr := range mostRecentReceive {
					if possiblePartner, ok := mrr[id]; ok {
						hb := clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hb == clock.Concurrent || hb == clock.Before) {
							found = true
							partner = append(partner, possiblePartner)
						} else if !buffered && hb == clock.Concurrent {
							found = true
							partner = append(partner, possiblePartner)
						}
					}
				}
			} else { // recv
				for _, mrs := range mostRecentSend {
					if possiblePartner, ok := mrs[id]; ok {
						hb := clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hb == clock.Concurrent || hb == clock.After) {
							found = true
							partner = append(partner, possiblePartner)
						} else if !buffered && hb == clock.Concurrent {
							found = true
							partner = append(partner, possiblePartner)
						}
					}
				}
			}
		}

		selectCases = append(selectCases,
			allSelectCase{se, id, elemWithVc{vc, se}, send, buffered, found, partner, executed, casi})

	}
}

// CheckForSelectCaseWithPartnerChannel checks for select cases with a valid
// partner. Call whenever a channel operation is processed.
//
// Parameter:
//   - id int: The id of the channel
//   - vc VectorClock: The vector clock
//   - send bool: True if the operation is a send
//   - buffered bool: True if the channel is buffered
func CheckForSelectCaseWithPartnerChannel(ch trace.TraceElement, vc *clock.VectorClock,
	send bool, buffered bool) {

	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for i, c := range selectCases {
		if c.partnerFound || c.chanID != ch.GetID() || c.send == send || c.elem.elem.GetTID() == ch.GetTID() {
			continue
		}

		hb := clock.GetHappensBefore(vc, c.elem.vc)
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
			selectCases[i].partnerFound = true
			selectCases[i].partner = append(selectCases[i].partner, ElemWithVcVal{ch, vc, 0})
		}
	}
}

// CheckForSelectCaseWithPartnerClose checks for select cases without a valid
// partner. Call whenever a close operation is processed.
//
// Parameter:
//   - id int: The id of the channel
//   - vc VectorClock: The vector clock
func CheckForSelectCaseWithPartnerClose(cl *trace.TraceElementChannel, vc *clock.VectorClock) {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for i, c := range selectCases {
		if c.partnerFound || c.chanID != cl.GetID() || c.send {
			continue
		}

		hb := clock.GetHappensBefore(vc, c.elem.vc)
		found := false
		if c.buffered && (hb == clock.Concurrent || hb == clock.After) {
			found = true
		} else if !c.buffered && hb == clock.Concurrent {
			found = true
		}

		if found {
			selectCases[i].partnerFound = true
			selectCases[i].partner = append(selectCases[i].partner, ElemWithVcVal{cl, vc, 0})
		}
	}
}

// GetNumberSelectCasesWithPartner returns the number of cases with possible partner
//
// Returns:
//   - int: the total number of select cases with possible partner over all selects
func GetNumberSelectCasesWithPartner() int {
	return numberSelectCasesWithPartner
}

// Rerun the CheckForSelectCaseWithPartnerChannel for all channel. This
// is needed to find potential communication partners for not executed
// select cases, if the select was executed after the channel
func rerunCheckForSelectCaseWithPartnerChannel() {
	for _, tr := range MainTrace.GetTraces() {
		for _, elem := range tr {
			if e, ok := elem.(*trace.TraceElementChannel); ok {
				CheckForSelectCaseWithPartnerChannel(e, e.GetVC(),
					e.Operation() == trace.SendOp, e.IsBuffered())
			}
		}
	}
}
