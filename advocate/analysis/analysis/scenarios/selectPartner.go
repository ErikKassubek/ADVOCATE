// Copyright (c) 2024 Erik Kassubek
//
// File: analysisSelectPartner.go
// Brief: Trace analysis for detection of select cases without any possible partners
//
// Author: Erik Kassubek
// Created: 2024-03-04
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/trace"
	"advocate/utils/timer"
)

// CheckForSelectCaseWithPartner checks for select cases with a valid
// partner. Call when all elements have been processed.
func CheckForSelectCaseWithPartner() {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	// check if not selected cases could be partners
	for i, c1 := range baseA.SelectCases {
		for j := i + 1; j < len(baseA.SelectCases); j++ {
			c2 := baseA.SelectCases[j]

			// if c1.partnerFound && c2.partnerFound {
			// 	continue
			// }

			if c1.ChanID != c2.ChanID || c1.Elem.Elem.GetTID() == c2.Elem.Elem.GetTID() || c1.Send == c2.Send {
				continue
			}

			if c2.Send { // c1 should be send, c2 should be recv
				c1, c2 = c2, c1
			}

			hbInfo := clock.GetHappensBefore(c1.Elem.Vc, c2.Elem.Vc)
			found := false
			if c1.Buffered && (hbInfo == hb.Concurrent || hbInfo == hb.After) {
				found = true
			} else if !c1.Buffered && hbInfo == hb.Concurrent {
				found = true
			}

			if found {
				baseA.SelectCases[i].PartnerFound = true
				baseA.SelectCases[j].PartnerFound = true
				baseA.SelectCases[i].Partner = append(baseA.SelectCases[i].Partner, baseA.ElemWithVcVal{
					Elem: baseA.SelectCases[j].Sel,
					Vc:   baseA.SelectCases[j].Sel.GetVC(),
					Val:  0,
				})
				baseA.SelectCases[j].Partner = append(baseA.SelectCases[j].Partner, baseA.ElemWithVcVal{
					Elem: baseA.SelectCases[i].Sel,
					Vc:   baseA.SelectCases[i].Sel.GetVC(),
					Val:  0,
				})
			}
		}
	}

	if len(baseA.SelectCases) == 0 {
		return
	}

	// collect all cases with no partner and all not triggered cases with partner

	for _, c := range baseA.SelectCases {
		opjType := "C"
		if c.Send {
			opjType += "S"
		} else {
			opjType += "R"
		}

		if c.PartnerFound {
			c.Sel.AddCasesWithPosPartner(c.Casi)
			baseA.NumberSelectCasesWithPartner++
		}
	}
}

// CheckForSelectCaseWithPartnerSelect checks for select cases with a valid
// partner. Call whenever a select is processed.
//
// Parameter:
//   - se *TraceElementSelect: The trace elem
//   - vc *VectorClock: The vector clock
func CheckForSelectCaseWithPartnerSelect(se *trace.ElementSelect, vc *clock.VectorClock) {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for casi, c := range se.GetCases() {

		id := c.GetObjId()

		buffered := (c.GetQSize() > 0)
		send := (c.GetType(true) == trace.ChannelSend)

		found := false
		executed := false
		var partner = make([]baseA.ElemWithVcVal, 0)

		if casi == se.GetChosenIndex() && se.GetTPost() != 0 {
			// no need to check if the channel is the chosen case
			executed = true
			p := se.GetPartner()
			if p != nil {
				found = true
				vcTID := baseA.ElemWithVcVal{
					Elem: p,
					Vc:   p.GetVC().Copy(),
					Val:  0,
				}
				partner = append(partner, vcTID)
			}
		} else {
			// not select cases
			if send {
				for _, mrr := range baseA.MostRecentReceive {
					if possiblePartner, ok := mrr[id]; ok {
						hbInfo := clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hbInfo == hb.Concurrent || hbInfo == hb.Before) {
							found = true
							partner = append(partner, possiblePartner)
						} else if !buffered && hbInfo == hb.Concurrent {
							found = true
							partner = append(partner, possiblePartner)
						}
					}
				}
			} else { // recv
				for _, mrs := range baseA.MostRecentSend {
					if possiblePartner, ok := mrs[id]; ok {
						hbInfo := clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hbInfo == hb.Concurrent || hbInfo == hb.After) {
							found = true
							partner = append(partner, possiblePartner)
						} else if !buffered && hbInfo == hb.Concurrent {
							found = true
							partner = append(partner, possiblePartner)
						}
					}
				}
			}
		}

		baseA.SelectCases = append(baseA.SelectCases,
			baseA.AllSelectCase{Sel: se,
				ChanID: id,
				Elem: baseA.ElemWithVc{
					Vc:   vc,
					Elem: se,
				},
				Send:         send,
				Buffered:     buffered,
				PartnerFound: found,
				Partner:      partner,
				Exec:         executed,
				Casi:         casi})

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
func CheckForSelectCaseWithPartnerChannel(ch trace.Element, vc *clock.VectorClock,
	send bool, buffered bool) {

	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for i, c := range baseA.SelectCases {
		if c.PartnerFound || c.ChanID != ch.GetObjId() || c.Send == send || c.Elem.Elem.GetTID() == ch.GetTID() {
			continue
		}

		hbInfo := clock.GetHappensBefore(vc, c.Elem.Vc)
		found := false
		if send {
			if buffered && (hbInfo == hb.Concurrent || hbInfo == hb.Before) {
				found = true
			} else if !buffered && hbInfo == hb.Concurrent {
				found = true
			}
		} else {
			if buffered && (hbInfo == hb.Concurrent || hbInfo == hb.After) {
				found = true
			} else if !buffered && hbInfo == hb.Concurrent {
				found = true
			}
		}

		if found {
			baseA.SelectCases[i].PartnerFound = true
			baseA.SelectCases[i].Partner = append(baseA.SelectCases[i].Partner, baseA.ElemWithVcVal{
				Elem: ch,
				Vc:   vc,
				Val:  0,
			})
		}
	}
}

// CheckForSelectCaseWithPartnerClose checks for select cases without a valid
// partner. Call whenever a close operation is processed.
//
// Parameter:
//   - id int: The id of the channel
//   - vc VectorClock: The vector clock
func CheckForSelectCaseWithPartnerClose(cl *trace.ElementChannel, vc *clock.VectorClock) {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for i, c := range baseA.SelectCases {
		if c.PartnerFound || c.ChanID != cl.GetObjId() || c.Send {
			continue
		}

		hbInfo := clock.GetHappensBefore(vc, c.Elem.Vc)
		found := false
		if c.Buffered && (hbInfo == hb.Concurrent || hbInfo == hb.After) {
			found = true
		} else if !c.Buffered && hbInfo == hb.Concurrent {
			found = true
		}

		if found {
			baseA.SelectCases[i].PartnerFound = true
			baseA.SelectCases[i].Partner = append(baseA.SelectCases[i].Partner, baseA.ElemWithVcVal{
				Elem: cl,
				Vc:   vc,
				Val:  0,
			})
		}
	}
}

// RerunCheckForSelectCaseWithPartnerChannel reruns the
// CheckForSelectCaseWithPartnerChannel for all channel. This
// is needed to find potential communication partners for not executed
// select cases, if the select was executed after the channel
func RerunCheckForSelectCaseWithPartnerChannel() {
	for _, tr := range baseA.MainTrace.GetTraces() {
		for _, elem := range tr {
			if e, ok := elem.(*trace.ElementChannel); ok {
				CheckForSelectCaseWithPartnerChannel(e, e.GetVC(),
					e.GetType(true) == trace.ChannelSend, e.IsBuffered())
			}
		}
	}
}
