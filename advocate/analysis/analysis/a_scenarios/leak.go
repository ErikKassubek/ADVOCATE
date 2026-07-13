// Copyright (c) 2024 Erik Kassubek
//
// File: leak.go
// Brief: Trace analysis for routine leaks
//
// Author: Erik Kassubek
// Created: 2024-01-28
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/results"
	"advocate/utils/timer"
	"strings"
)

// Struct to store found leaks
type TERLeak struct {
	resultType helper.ResultType
	argType1   string
	arg1       []results.ResultElem
	argType2   string
	arg2       []results.ResultElem
}

var leaks = make(map[int]TERLeak, 0) // based on trace (could be released if program continues)

// CheckForLeakChannelStuck is run for channel operation without a post event.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc VectorClock: The vector clock of the operation
func CheckForLeakChannelStuck(ch *trace.ElementChannel, vc *a_clock.VectorClock) {
	// buffered := (ch.GetQSize() != 0)
	id := ch.GetObjId()
	opC := ch.GetType(true)
	routine := ch.GetRoutine()

	if opC == trace.ChannelClose {
		return // close
	}

	arg1 := results.TraceElementResult{
		RoutineID: routine, ObjID: id, TRequest: ch.GetT(trace.Request), ObjType: opC, File: ch.GetFile(), Line: ch.GetLine()}

	if id == -1 {
		leaks[routine] = TERLeak{helper.LNilChan,
			"Channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
	} else {
		if ch.IsBuffered() {
			leaks[routine] = TERLeak{helper.LBufferedWithout,
				"Channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
		} else {
			leaks[routine] = TERLeak{helper.LUnbufferedWithout,
				"Channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
		}
	}
}

// CheckForLeak is run after all operations have been analyzed, and checks if there are still leaking
// operations without a possible partner.
func CheckForLeak() {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	// channel
	for _, vcTIDs := range a_base.LeakingChannels {
		buffered := false
		for _, vcTID := range vcTIDs {
			if vcTID.TID == "" {
				continue
			}

			routineID := vcTID.Routine

			found := false
			var partner a_base.AllSelectCase
			for _, c := range a_base.SelectCases {
				if c.ChanID != vcTID.ID {
					continue
				}

				if (c.Send && vcTID.TypeVal == trace.ChannelSend) || (!c.Send && vcTID.TypeVal == trace.ChannelRecv) {
					continue
				}

				hbInfo := a_clock.GetHappensBefore(c.Elem.Vc, vcTID.Vc)
				if hbInfo == a_hb.Concurrent {
					found = true
					if c.Buffered {
						buffered = true
					}
					partner = c
					break
				}

				if c.Buffered {
					if (c.Send && hbInfo == a_hb.Before) || (!c.Send && hbInfo == a_hb.After) {
						found = true
						buffered = true
						partner = c
						break
					}
				}
			}

			if found {
				file1, line1, tPre1, err := trace.InfoFromTID(vcTID.TID)
				if err != nil {
					log.Errorf("Error in trace.InfoFromTID(%s)\n", vcTID.TID)
					continue
				}

				elem2 := partner.Elem.Elem
				file2 := elem2.GetFile()
				line2 := elem2.GetLine()
				tPre2 := elem2.GetT(trace.Request)

				if vcTID.Sel {
					arg1 := results.TraceElementResult{ // select
						RoutineID: routineID, ObjID: vcTID.ID, TRequest: tPre1, ObjType: "SS", File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.Sel.GetObjId(), TRequest: tPre2, ObjType: "SS", File: file2, Line: line2}

					timer, ctx := chanIsTimerOrCtx(vcTID.ID)
					if !timer {
						if !ctx {
							leaks[routineID] = TERLeak{helper.LSelectWith,
								"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
						} else {
							leaks[routineID] = TERLeak{helper.LContext,
								"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}
					} else {
						leaks[routineID] = TERLeak{helper.LUnknown,
							"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					}
				} else {
					var bugType helper.ResultType = helper.LUnbufferedWith
					if buffered {
						bugType = helper.LBufferedWith
					}

					arg1 := results.TraceElementResult{ // channel
						RoutineID: routineID, ObjID: vcTID.ID, TRequest: tPre1, ObjType: vcTID.TypeVal, File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.Sel.GetObjId(), TRequest: tPre2, ObjType: "SS", File: file2, Line: line2}

					timer, ctx := chanIsTimerOrCtx(vcTID.ID)
					if !timer {
						if !ctx {
							leaks[routineID] = TERLeak{bugType,
								"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
						} else {
							leaks[routineID] = TERLeak{helper.LContext,
								"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}
					} else {
						leaks[routineID] = TERLeak{helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					}
				}

			} else {
				if vcTID.Sel {
					file, line, tPre, err := trace.InfoFromTID(vcTID.TID)
					if err != nil {
						log.Errorf("Error in trace.InfoFromTID(%s)\n", vcTID.TID)
						continue
					}

					arg1 := results.TraceElementResult{
						RoutineID: vcTID.Routine, ObjID: vcTID.SelID, TRequest: tPre, ObjType: "SS", File: file, Line: line}

					timer, ctx := chanIsTimerOrCtx(vcTID.ID)
					if !timer {
						if !ctx {
							leaks[routineID] = TERLeak{helper.LSelectWithout,
								"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						} else {
							leaks[routineID] = TERLeak{helper.LContext,
								"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}
					} else {
						leaks[routineID] = TERLeak{helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					}

				} else {
					file, line, tPre, err := trace.InfoFromTID(vcTID.TID)
					if err != nil {
						log.Errorf("Error in trace.InfoFromTID(%s)\n", vcTID.TID)
						continue
					}

					arg1 := results.TraceElementResult{
						RoutineID: vcTID.Routine, ObjID: vcTID.ID, TRequest: tPre, ObjType: vcTID.TypeVal, File: file, Line: line}

					var bugType helper.ResultType = helper.LUnbufferedWithout
					if buffered {
						bugType = helper.LBufferedWithout
					}

					timer, ctx := chanIsTimerOrCtx(vcTID.ID)
					if !timer {
						if !ctx {
							leaks[routineID] = TERLeak{bugType,
								"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
							leaks[routineID] = TERLeak{helper.LContext,
								"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}
					} else {
						leaks[routineID] = TERLeak{helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					}
				}
			}
		}
	}
}

// CheckForLeakSelectStuck is run for select operation without a post event.
// It checks if the operation has a possible communication partner in
// baseA.MostRecentSend, baseA.MostRecentReceive or baseA.ClosebaseA.
//
//	add the if to leaks
//
// If not, add all elements to baseA.LeakingChannels, for later check.
//
// Parameter:
//   - se *TraceElementSelect: The trace element
//   - ids int: The channel ids
//   - buffered []bool: If the channels are buffered
//   - vc *VectorClock: The vector clock of the operation
//   - opTypes []int: An identifier for the type of the operations (send = 0, recv = 1)
func CheckForLeakSelectStuck(se *trace.ElementSelect, ids []int, buffered []bool, vc *a_clock.VectorClock, opTypes []trace.OperationType) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	foundPartner := false

	routine := se.GetRoutine()
	id := se.GetObjId()
	tPre := se.GetT(trace.Request)

	if len(ids) == 0 {
		file, line, _, err := trace.InfoFromTID(se.GetTID())
		if err != nil {
			log.Errorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
			return
		}

		arg1 := results.TraceElementResult{
			RoutineID: routine, ObjID: id, TRequest: tPre, ObjType: "SS", File: file, Line: line}

		timer, ctx := isLeakTimerOrCtx(se)
		if !timer && !se.GetContainsDefault() {
			if !ctx {
				leaks[routine] = TERLeak{helper.LSelectWithout,
					"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
			} else {
				leaks[routine] = TERLeak{helper.LContext,
					"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
			}
		} else {
			leaks[routine] = TERLeak{helper.LUnknown,
				"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
		}

		return
	}

	for i, id := range ids {
		switch opTypes[i] {
		case trace.ChannelSend:
			for routinePartner, mrr := range a_base.MostRecentReceive {
				if recv, ok := mrr[id]; ok {
					if a_clock.GetHappensBefore(vc, mrr[id].Vc) == a_hb.Concurrent {
						file1, line1, _, err1 := trace.InfoFromTID(se.GetTID()) // select
						if err1 != nil {
							log.Errorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
							return
						}
						file2, line2, tPre2, err2 := trace.InfoFromTID(recv.Elem.GetTID()) // partner
						if err2 != nil {
							log.Errorf("Error in trace.InfoFromTID(%s)\n", recv.Elem.GetTID())
							return
						}

						arg1 := results.TraceElementResult{
							RoutineID: routine, ObjID: id, TRequest: tPre, ObjType: "SS", File: file1, Line: line1}
						arg2 := results.TraceElementResult{
							RoutineID: routinePartner, ObjID: id, TRequest: tPre2, ObjType: "CR", File: file2, Line: line2}

						timer, ctx := isLeakTimerOrCtx(se)
						if !timer && !se.GetContainsDefault() {
							if !ctx {
								leaks[routine] = TERLeak{helper.LSelectWith,
									"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
								foundPartner = true
							} else {
								leaks[routine] = TERLeak{helper.LContext,
									"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
							}
						} else {
							leaks[routine] = TERLeak{helper.LUnknown,
								"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}
					}
				}
			}
		case trace.ChannelRecv:
			for routinePartner, mrs := range a_base.MostRecentSend {
				if send, ok := mrs[id]; ok {
					if a_clock.GetHappensBefore(vc, mrs[id].Vc) == a_hb.Concurrent {
						file1, line1, _, err1 := trace.InfoFromTID(se.GetTID()) // select
						if err1 != nil {
							log.Errorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
							return
						}
						file2, line2, tPre2, err2 := trace.InfoFromTID(send.Elem.GetTID()) // partner
						if err2 != nil {
							log.Errorf("Error in trace.InfoFromTID(%s)\n", send.Elem.GetTID())
							return
						}

						arg1 := results.TraceElementResult{
							RoutineID: routine, ObjID: id, TRequest: tPre, ObjType: "SS", File: file1, Line: line1}
						arg2 := results.TraceElementResult{
							RoutineID: routinePartner, ObjID: id, TRequest: tPre2, ObjType: "CS", File: file2, Line: line2}

						timer, ctx := isLeakTimerOrCtx(se)
						if !timer && !se.GetContainsDefault() {
							if !ctx {
								leaks[routine] = TERLeak{helper.LSelectWith,
									"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
							} else {
								leaks[routine] = TERLeak{helper.LContext,
									"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
							}
						} else {
							leaks[routine] = TERLeak{helper.LUnknown,
								"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}

						foundPartner = true
					}
				}
			}
			if cl, ok := a_base.CloseData[id]; ok {
				file1, line1, _, err1 := trace.InfoFromTID(se.GetTID()) // select
				if err1 != nil {
					log.Errorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
					return
				}
				file2, line2, tPre2, err2 := trace.InfoFromTID(cl.GetTID()) // partner
				if err2 != nil {
					log.Errorf("Error in trace.InfoFromTID(%s)\n", cl.GetTID())
					return
				}

				arg1 := results.TraceElementResult{
					RoutineID: routine, ObjID: id, TRequest: tPre, ObjType: "SS", File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: cl.GetRoutine(), ObjID: id, TRequest: tPre2, ObjType: "CS", File: file2, Line: line2}

				timer, ctx := isLeakTimerOrCtx(se)
				if !timer && !se.GetContainsDefault() {
					if !ctx {
						leaks[routine] = TERLeak{helper.LSelectWith,
							"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
					} else {
						leaks[routine] = TERLeak{helper.LContext,
							"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					}
				} else {
					leaks[routine] = TERLeak{helper.LUnknown,
						"select", []results.ResultElem{arg1}, "", []results.ResultElem{}}
				}

				foundPartner = true
			}
		}
	}

	if !foundPartner {
		for i, id := range ids {
			// add all select operations to leaking Channels,
			a_base.LeakingChannels[id] = append(a_base.LeakingChannels[id], a_base.VectorClockTID2{
				Routine:  routine,
				ID:       id,
				Vc:       vc,
				TID:      se.GetTID(),
				TypeVal:  opTypes[i],
				Val:      tPre,
				Buffered: buffered[i],
				Sel:      true,
				SelID:    id,
			})
		}
	}
}

// CheckForLeakMutex is run for mutex operation without a post event.
// It add the leak to leaks
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func CheckForLeakMutex(mu *trace.ElementMutex) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	id := mu.GetObjId()
	opM := mu.GetType(true)
	routineID := mu.GetRoutine()

	if _, ok := a_base.MostRecentAcquireTotal[id]; !ok {
		return
	}

	elem := a_base.MostRecentAcquireTotal[id].Elem

	file2, line2, tPre2 := elem.GetFile(), elem.GetLine(), elem.GetT(trace.Request)

	switch opM {
	case trace.MutexLock, trace.MutexRLock:
	default: // only lock and rlock can lead to leak
		return
	}

	objType2 := elem.GetType(true)
	switch objType2 {
	case trace.MutexLock, trace.MutexRLock:
	default: // only lock and rlock can lead to leak
		return
	}

	arg1 := results.TraceElementResult{
		RoutineID: mu.GetRoutine(), ObjID: id, TRequest: mu.GetT(trace.Request), ObjType: opM, File: mu.GetFile(), Line: mu.GetLine()}

	arg2 := results.TraceElementResult{
		RoutineID: elem.GetRoutine(), ObjID: id, TRequest: tPre2, ObjType: objType2, File: file2, Line: line2}

	leaks[routineID] = TERLeak{helper.LMutex,
		"mutex", []results.ResultElem{arg1}, "last", []results.ResultElem{arg2}}
}

// AddMostRecentAcquireTotal adds the most recent acquire operation for a mutex
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//   - vc VectorClock: The vector clock of the operation
//   - op int: The operation on the mutex
func AddMostRecentAcquireTotal(mu *trace.ElementMutex, vc *a_clock.VectorClock) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	a_base.MostRecentAcquireTotal[mu.GetObjId()] = a_base.ElemWithVc{Elem: mu, Vc: vc.Copy()}
}

// CheckForLeakWait is run for wait group operation without a post event.
// It add the leak to leaks
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func CheckForLeakWait(wa *trace.ElementWait) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	file, line, tPre, err := trace.InfoFromTID(wa.GetTID())
	if err != nil {
		log.Errorf("Error in trace.InfoFromTID(%s)\n", wa.GetTID())
		return
	}

	routineID := wa.GetRoutine()

	arg := results.TraceElementResult{
		RoutineID: routineID, ObjID: wa.GetObjId(), TRequest: tPre, ObjType: "WW", File: file, Line: line}

	leaks[routineID] = TERLeak{helper.LWaitGroup,
		"wait", []results.ResultElem{arg}, "", []results.ResultElem{}}
}

// CheckForLeakCond is run for conditional variable operation without a post
// event. It adds a leak to the results
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CheckForLeakCond(co *trace.ElementCond) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	file, line, tPre, err := trace.InfoFromTID(co.GetTID())
	if err != nil {
		log.Errorf("Error in trace.InfoFromTID(%s)\n", co.GetTID())
		return
	}

	routineID := co.GetRoutine()

	arg := results.TraceElementResult{
		RoutineID: routineID, ObjID: co.GetObjId(), TRequest: tPre, ObjType: "DW", File: file, Line: line}

	leaks[routineID] = TERLeak{helper.LCond,
		"cond", []results.ResultElem{arg}, "", []results.ResultElem{}}
}

// CheckForStuckRoutine iterates over all routines and checks if the routines finished.
// Only record leaking routines, that don't have a leaking element (tPost = 0)
// as its last element, since they are recorded separately
//
// Parameter
//   - simple bool: set to true, if only simple analysis is run
//
// Returns
//   - bool: true if a stuck routine was found
func CheckForStuckRoutine(simple bool) bool {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	res := false

	for routId, routine := range a_base.MainTrace.GetTraces() {
		if routine.Empty() {
			continue
		}

		lastElem := routine.Last()
		switch lastElem.(type) {
		case *trace.ElementRoutineEnd:
			continue
		}

		lastTPost := lastElem.GetT(trace.Commit)

		leakType := helper.LUnknown
		objectType := trace.None
		// do not record extra if a leak with a blocked operation is present
		// if simple, find the type of blocking
		if lastTPost == 0 {
			if simple {
				ot := lastElem.GetType(true)
				objectType = ot
				switch ot {
				case trace.ChannelSend, trace.ChannelRecv:
					c := lastElem.(*trace.ElementChannel)
					if c.GetObjId() == -1 {
						leakType = helper.LNilChan
					} else if lastElem.(*trace.ElementChannel).IsBuffered() {
						leakType = helper.LBufferedWithout
					} else {
						leakType = helper.LUnbufferedWithout
					}
				case "DW":
					leakType = helper.LCond
				case "ML", "MR":
					leakType = helper.LMutex
				case "WW":
					leakType = helper.LWaitGroup
				case "SS":
					if lastElem.(*trace.ElementSelect).GetContainsDefault() {
						leakType = helper.LUnknown
						objectType = trace.None
					} else {
						leakType = helper.LSelectWithout
					}
				default:
					objectType = trace.None
				}
			} else {
				continue
			}
		}

		arg := results.TraceElementResult{
			RoutineID: routId, ObjID: lastElem.GetObjId(), TRequest: lastElem.GetT(trace.Request),
			ObjType: objectType, File: lastElem.GetFile(), Line: lastElem.GetLine(),
		}

		timer, ctx := isLeakTimerOrCtx(lastElem)

		if leakType == helper.LUnknown {
			leaks[routId] = TERLeak{leakType,
				"elem", []results.ResultElem{arg}, "", []results.ResultElem{}}
		} else if timer {
			leaks[routId] = TERLeak{helper.LUnknown,
				"elem", []results.ResultElem{arg}, "", []results.ResultElem{}}
		} else if ctx {
			leaks[routId] = TERLeak{helper.LContext,
				"elem", []results.ResultElem{arg}, "", []results.ResultElem{}}
		} else {
			leaks[routId] = TERLeak{leakType,
				"elem", []results.ResultElem{arg}, "", []results.ResultElem{}}
		}

		res = true
	}

	return res
}

func isLeakTimerOrCtx(elem trace.Element) (bool, bool) {
	isTimer, isContext := false, false
	switch e := elem.(type) {
	case *trace.ElementChannel:
		return chanIsTimerOrCtx(elem.GetObjId())
	case *trace.ElementSelect:
		for _, c := range e.GetCases() {
			ti, co := chanIsTimerOrCtx(c.GetObjId())
			if ti {
				isTimer = true
			}
			if co {
				isContext = true
			}
		}
	default:
		return false, false
	}

	return isTimer, isContext
}

func chanIsTimerOrCtx(id int) (bool, bool) {
	pos, ok := a_base.NewChan[id]

	if !ok {
		return false, false
	}

	return strings.Contains(pos, paths.Join(true, true, "src", "time")), strings.Contains(pos, paths.Join(true, true, "src", "context"))
}
