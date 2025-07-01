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
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/timer"
)

// CheckForSelectCaseWithPartner checks for select cases with a valid
// partner. Call when all elements have been processed.
func CheckForSelectCaseWithPartner() {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	// check if not selected cases could be partners
	for i, c1 := range data.SelectCases {
		for j := i + 1; j < len(data.SelectCases); j++ {
			c2 := data.SelectCases[j]

			// if c1.partnerFound && c2.partnerFound {
			// 	continue
			// }

			if c1.ChanID != c2.ChanID || c1.Elem.Elem.GetTID() == c2.Elem.Elem.GetTID() || c1.Send == c2.Send {
				continue
			}

			if c2.Send { // c1 should be send, c2 should be recv
				c1, c2 = c2, c1
			}

			hb := clock.GetHappensBefore(c1.Elem.Vc, c2.Elem.Vc)
			found := false
			if c1.Buffered && (hb == clock.Concurrent || hb == clock.After) {
				found = true
			} else if !c1.Buffered && hb == clock.Concurrent {
				found = true
			}

			if found {
				data.SelectCases[i].PartnerFound = true
				data.SelectCases[j].PartnerFound = true
				data.SelectCases[i].Partner = append(data.SelectCases[i].Partner, data.ElemWithVcVal{data.SelectCases[j].Sel, data.SelectCases[j].Sel.GetVC(), 0})
				data.SelectCases[j].Partner = append(data.SelectCases[j].Partner, data.ElemWithVcVal{data.SelectCases[i].Sel, data.SelectCases[i].Sel.GetVC(), 0})
			}
		}
	}

	if len(data.SelectCases) == 0 {
		return
	}

	// collect all cases with no partner and all not triggered cases with partner

	for _, c := range data.SelectCases {
		opjType := "C"
		if c.Send {
			opjType += "S"
		} else {
			opjType += "R"
		}

		if c.PartnerFound {
			c.Sel.AddCasesWithPosPartner(c.Casi)
			data.NumberSelectCasesWithPartner++
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
		var partner = make([]data.ElemWithVcVal, 0)

		if casi == se.GetChosenIndex() && se.GetTPost() != 0 {
			// no need to check if the channel is the chosen case
			executed = true
			p := se.GetPartner()
			if p != nil {
				found = true
				vcTID := data.ElemWithVcVal{
					p, p.GetVC().Copy(), 0,
				}
				partner = append(partner, vcTID)
			}
		} else {
			// not select cases
			if send {
				for _, mrr := range data.MostRecentReceive {
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
				for _, mrs := range data.MostRecentSend {
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

		data.SelectCases = append(data.SelectCases,
			data.AllSelectCase{se, id, data.ElemWithVc{vc, se}, send, buffered, found, partner, executed, casi})

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

	for i, c := range data.SelectCases {
		if c.PartnerFound || c.ChanID != ch.GetID() || c.Send == send || c.Elem.Elem.GetTID() == ch.GetTID() {
			continue
		}

		hb := clock.GetHappensBefore(vc, c.Elem.Vc)
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
			data.SelectCases[i].PartnerFound = true
			data.SelectCases[i].Partner = append(data.SelectCases[i].Partner, data.ElemWithVcVal{ch, vc, 0})
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

	for i, c := range data.SelectCases {
		if c.PartnerFound || c.ChanID != cl.GetID() || c.Send {
			continue
		}

		hb := clock.GetHappensBefore(vc, c.Elem.Vc)
		found := false
		if c.Buffered && (hb == clock.Concurrent || hb == clock.After) {
			found = true
		} else if !c.Buffered && hb == clock.Concurrent {
			found = true
		}

		if found {
			data.SelectCases[i].PartnerFound = true
			data.SelectCases[i].Partner = append(data.SelectCases[i].Partner, data.ElemWithVcVal{cl, vc, 0})
		}
	}
}

// Rerun the CheckForSelectCaseWithPartnerChannel for all channel. This
// is needed to find potential communication partners for not executed
// select cases, if the select was executed after the channel
func rerunCheckForSelectCaseWithPartnerChannel() {
	for _, tr := range data.MainTrace.GetTraces() {
		for _, elem := range tr {
			if e, ok := elem.(*trace.TraceElementChannel); ok {
				CheckForSelectCaseWithPartnerChannel(e, e.GetVC(),
					e.Operation() == trace.SendOp, e.IsBuffered())
			}
		}
	}
}
