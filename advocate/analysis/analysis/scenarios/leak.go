// Copyright (c) 2024 Erik Kassubek
//
// File: analysisLeak.go
// Brief: Trace analysis for routine leaks
//
// Author: Erik Kassubek
// Created: 2024-01-28
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/data"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
	"strings"
)

// CheckForLeakChannelStuck is run for channel operation without a post event.
// It checks if the operation has a possible communication partner in
// data.MostRecentSend, data.MostRecentReceive or data.CloseData.
// If so, add an error or warning to the result.
// If not, add to data.LeakingChannels, for later check.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc VectorClock: The vector clock of the operation
//   - func CheckForLeakChannelStuck(routineID int, objID int, vc clock.VectorClock, tID string, opType int, buffered bool) {
func CheckForLeakChannelStuck(ch *trace.ElementChannel, vc *clock.VectorClock) {
	buffered := (ch.GetQSize() != 0)
	id := ch.GetID()
	opC := ch.GetOpC()
	routine := ch.GetRoutine()

	if id == -1 {
		objType := "C"
		switch opC {
		case trace.SendOp:
			objType += "S"
		case trace.RecvOp:
			objType += "R"
		default:
			return // close
		}

		arg1 := results.TraceElementResult{
			RoutineID: routine, ObjID: id, TPre: ch.GetTPre(), ObjType: objType, File: ch.GetFile(), Line: ch.GetLine()}

		results.Result(results.CRITICAL, helper.LNilChan,
			"Channel", []results.ResultElem{arg1}, "", []results.ResultElem{})

		return
	}

	// if !buffered {
	foundPartner := false

	switch opC {
	case trace.SendOp: // send
		for partnerRout, mrr := range data.MostRecentReceive {
			if _, ok := mrr[id]; ok {
				if clock.GetHappensBefore(mrr[id].Vc, vc) == hb.Concurrent {

					var bugType helper.ResultType = helper.LUnbufferedWith
					if buffered {
						bugType = helper.LBufferedWith
					}

					file1, line1, tPre1, err := trace.InfoFromTID(ch.GetTID())
					if err != nil {
						log.Errorf("Error in trace.InfoFromTID(%s)\n", ch.GetTID())
						return
					}
					file2, line2, tPre2, err := trace.InfoFromTID(mrr[id].Elem.GetTID())
					if err != nil {
						log.Errorf("Error in trace.InfoFromTID(%s)\n", mrr[id].Elem.GetTID())
						return
					}

					arg1 := results.TraceElementResult{
						RoutineID: routine, ObjID: id, TPre: tPre1, ObjType: "CS", File: file1, Line: line1}
					arg2 := results.TraceElementResult{
						RoutineID: partnerRout, ObjID: id, TPre: tPre2, ObjType: "CR", File: file2, Line: line2}

					if !isLeakTimerOrCtx(ch) {
						results.Result(results.CRITICAL, bugType,
							"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
						foundPartner = true
					} else {
						results.Result(results.CRITICAL, helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
					}
				}
			}
		}
	case trace.RecvOp: // recv
		for partnerRout, mrs := range data.MostRecentSend {
			if _, ok := mrs[id]; ok {
				if clock.GetHappensBefore(mrs[id].Vc, vc) == hb.Concurrent {

					var bugType = helper.LUnbufferedWith
					if buffered {
						bugType = helper.LBufferedWith
					}

					arg1 := results.TraceElementResult{
						RoutineID: routine, ObjID: id, TPre: ch.GetTPre(), ObjType: "CR", File: ch.GetFile(), Line: ch.GetLine()}
					arg2 := results.TraceElementResult{
						RoutineID: partnerRout, ObjID: id, TPre: mrs[id].Elem.GetTPre(), ObjType: "CS", File: mrs[id].Elem.GetFile(), Line: mrs[id].Elem.GetLine()}

					if !isLeakTimerOrCtx(ch) {
						results.Result(results.CRITICAL, bugType,
							"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
						foundPartner = true
					} else {
						results.Result(results.CRITICAL, helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
					}
				}
			}
		}

	}

	if !foundPartner {
		data.LeakingChannels[id] = append(data.LeakingChannels[id], data.VectorClockTID2{
			Routine: routine,
			ID:      id,
			Vc:      vc,
			TID:     ch.GetTID(), TypeVal: int(opC),
			Val:      -1,
			Buffered: buffered,
			Sel:      false,
			SelID:    0,
		})
	}
}

// CheckForLeakChannelRun is run for channel operation with a post event.
// It checks if the operation would be possible communication partner for a
// stuck operation in data.LeakingChannels.
// If so, add an error or warning to the result and remove the stuck operation.
//
// Parameter:
//   - routineID int: The routine id
//   - objID int: The channel id
//   - vc VectorClock: The vector clock of the operation
//   - opType int: An identifier for the type of the operation (send = 0, recv = 1, close = 2)
//   - buffered bool: If the channel is buffered
func CheckForLeakChannelRun(routineID int, objID int, elemVc data.ElemWithVc, opType int, buffered bool) bool {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	res := false
	if opType == 0 || opType == 2 { // send or close
		for i, vcTID2 := range data.LeakingChannels[objID] {
			if vcTID2.Val != 1 {
				continue
			}

			if clock.GetHappensBefore(vcTID2.Vc, elemVc.Vc) == hb.Concurrent {
				var bugType = helper.LUnbufferedWith
				if buffered {
					bugType = helper.LBufferedWith
				}

				file1, line1, tPre1, err1 := trace.InfoFromTID(vcTID2.TID) // leaking
				if err1 != nil {
					log.Errorf("Error in trace.InfoFromTID(%s)\n", vcTID2.TID)
					return res
				}

				elem2 := elemVc.Elem

				objType := "C"
				if opType == 0 {
					objType += "S"
				} else {
					objType += "C"
				}

				arg1 := results.TraceElementResult{
					RoutineID: routineID, ObjID: objID, TPre: tPre1, ObjType: "CR", File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: vcTID2.Routine, ObjID: objID, TPre: elem2.GetTPre(), ObjType: objType, File: elem2.GetFile(), Line: elem2.GetLine()}

				if chanIsTimerOrCtx(objID) {
					results.Result(results.CRITICAL, helper.LUnknown,
						"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
					continue
				}

				results.Result(results.CRITICAL, bugType,
					"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})

				res = true

				// remove the stuck operation from the list. If it is a select, remove all operations with the same val
				if vcTID2.Val == -1 {
					data.LeakingChannels[objID] = append(data.LeakingChannels[objID][:i], data.LeakingChannels[objID][i+1:]...)
				} else {
					for j, vcTID3 := range data.LeakingChannels[objID] {
						if vcTID3.Val == vcTID2.Val {
							data.LeakingChannels[objID] = append(data.LeakingChannels[objID][:j], data.LeakingChannels[objID][j+1:]...)
						}
					}
				}
			}
		}
	} else if opType == 1 { // recv
		for i, vcTID2 := range data.LeakingChannels[objID] {
			objType := "C"
			switch vcTID2.Val {
			case 0:
				objType += "S"
			case 2:
				objType += "C"
			default:
				continue
			}

			if clock.GetHappensBefore(vcTID2.Vc, elemVc.Vc) == hb.Concurrent {

				var bugType = helper.LUnbufferedWith
				if buffered {
					bugType = helper.LBufferedWith
				}

				file1, line1, tPre1, err1 := trace.InfoFromTID(vcTID2.TID) // leaking
				if err1 != nil {
					log.Errorf("Error in trace.InfoFromTID(%s)\n", vcTID2.TID)
					return res
				}

				elem2 := elemVc.Elem

				arg1 := results.TraceElementResult{
					RoutineID: routineID, ObjID: objID, TPre: tPre1, ObjType: objType, File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: vcTID2.Routine, ObjID: objID, TPre: elem2.GetTPre(), ObjType: "CR", File: elem2.GetFile(), Line: elem2.GetLine()}

				if chanIsTimerOrCtx(objID) {
					results.Result(results.CRITICAL, helper.LUnknown,
						"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
					continue
				}

				results.Result(results.CRITICAL, bugType,
					"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
				res = true

				// remove the stuck operation from the list. If it is a select, remove all operations with the same val
				if vcTID2.Val == -1 {
					data.LeakingChannels[objID] = append(data.LeakingChannels[objID][:i], data.LeakingChannels[objID][i+1:]...)
				} else {
					for j, vcTID3 := range data.LeakingChannels[objID] {
						if vcTID3.Val == vcTID2.Val {
							data.LeakingChannels[objID] = append(data.LeakingChannels[objID][:j], data.LeakingChannels[objID][j+1:]...)
						}
					}
				}
			}
		}
	}
	return res
}

// CheckForLeak is run after all operations have been analyzed, and checks if there are still leaking
// operations without a possible partner.
func CheckForLeak() {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	// channel
	for _, vcTIDs := range data.LeakingChannels {
		buffered := false
		for _, vcTID := range vcTIDs {
			if vcTID.TID == "" {
				continue
			}

			found := false
			var partner data.AllSelectCase
			for _, c := range data.SelectCases {
				if c.ChanID != vcTID.ID {
					continue
				}

				if (c.Send && vcTID.TypeVal == 0) || (!c.Send && vcTID.TypeVal == 1) {
					continue
				}

				hbInfo := clock.GetHappensBefore(c.Elem.Vc, vcTID.Vc)
				if hbInfo == hb.Concurrent {
					found = true
					if c.Buffered {
						buffered = true
					}
					partner = c
					break
				}

				if c.Buffered {
					if (c.Send && hbInfo == hb.Before) || (!c.Send && hbInfo == hb.After) {
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
				tPre2 := elem2.GetTPre()

				if vcTID.Sel {

					arg1 := results.TraceElementResult{ // select
						RoutineID: vcTID.Routine, ObjID: vcTID.ID, TPre: tPre1, ObjType: "SS", File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.Sel.GetID(), TPre: tPre2, ObjType: "SS", File: file2, Line: line2}

					if !chanIsTimerOrCtx(vcTID.ID) {
						results.Result(results.CRITICAL, helper.LSelectWith,
							"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
					} else {
						results.Result(results.CRITICAL, helper.LUnknown,
							"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
					}
				} else {
					obType := "C"
					if vcTID.TypeVal == 0 {
						obType += "S"
					} else {
						obType += "R"
					}

					var bugType helper.ResultType = helper.LUnbufferedWith
					if buffered {
						bugType = helper.LBufferedWith
					}

					arg1 := results.TraceElementResult{ // channel
						RoutineID: vcTID.Routine, ObjID: vcTID.ID, TPre: tPre1, ObjType: obType, File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.Sel.GetID(), TPre: tPre2, ObjType: "SS", File: file2, Line: line2}

					if !chanIsTimerOrCtx(vcTID.ID) {
						results.Result(results.CRITICAL, bugType,
							"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
					} else {
						results.Result(results.CRITICAL, helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
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
						RoutineID: vcTID.Routine, ObjID: vcTID.SelID, TPre: tPre, ObjType: "SS", File: file, Line: line}

					results.Result(results.CRITICAL, helper.LSelectWithout,
						"select", []results.ResultElem{arg1}, "", []results.ResultElem{})

				} else {
					objType := "C"
					if vcTID.TypeVal == 0 {
						objType += "S"
					} else {
						objType += "R"
					}

					file, line, tPre, err := trace.InfoFromTID(vcTID.TID)
					if err != nil {
						log.Errorf("Error in trace.InfoFromTID(%s)\n", vcTID.TID)
						continue
					}

					arg1 := results.TraceElementResult{
						RoutineID: vcTID.Routine, ObjID: vcTID.ID, TPre: tPre, ObjType: objType, File: file, Line: line}

					var bugType helper.ResultType = helper.LUnbufferedWithout
					if buffered {
						bugType = helper.LBufferedWithout
					}

					results.Result(results.CRITICAL, bugType,
						"channel", []results.ResultElem{arg1}, "", []results.ResultElem{})
				}
			}
		}
	}
}

// CheckForLeakSelectStuck is run for select operation without a post event.
// It checks if the operation has a possible communication partner in
// data.MostRecentSend, data.MostRecentReceive or data.CloseData.
// If so, add an error or warning to the result.
// If not, add all elements to data.LeakingChannels, for later check.
//
// Parameter:
//   - se *TraceElementSelect: The trace element
//   - ids int: The channel ids
//   - buffered []bool: If the channels are buffered
//   - vc *VectorClock: The vector clock of the operation
//   - opTypes []int: An identifier for the type of the operations (send = 0, recv = 1)
func CheckForLeakSelectStuck(se *trace.ElementSelect, ids []int, buffered []bool, vc *clock.VectorClock, opTypes []int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	foundPartner := false

	routine := se.GetRoutine()
	id := se.GetID()
	tPre := se.GetTPre()

	if len(ids) == 0 {
		file, line, _, err := trace.InfoFromTID(se.GetTID())
		if err != nil {
			log.Errorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
			return
		}

		arg1 := results.TraceElementResult{
			RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file, Line: line}

		if !isLeakTimerOrCtx(se) && !se.GetContainsDefault() {
			results.Result(results.CRITICAL, helper.LSelectWithout,
				"select", []results.ResultElem{arg1}, "", []results.ResultElem{})
		} else {
			results.Result(results.CRITICAL, helper.LUnknown,
				"select", []results.ResultElem{arg1}, "", []results.ResultElem{})
		}

		return
	}

	for i, id := range ids {
		switch opTypes[i] {
		case 0: // send
			for routinePartner, mrr := range data.MostRecentReceive {
				if recv, ok := mrr[id]; ok {
					if clock.GetHappensBefore(vc, mrr[id].Vc) == hb.Concurrent {
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
							RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file1, Line: line1}
						arg2 := results.TraceElementResult{
							RoutineID: routinePartner, ObjID: id, TPre: tPre2, ObjType: "CR", File: file2, Line: line2}

						if !isLeakTimerOrCtx(se) && !se.GetContainsDefault() {
							results.Result(results.CRITICAL, helper.LSelectWith,
								"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
							foundPartner = true
						}
					}
				}
			}
		case 1: // recv
			for routinePartner, mrs := range data.MostRecentSend {
				if send, ok := mrs[id]; ok {
					if clock.GetHappensBefore(vc, mrs[id].Vc) == hb.Concurrent {
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
							RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file1, Line: line1}
						arg2 := results.TraceElementResult{
							RoutineID: routinePartner, ObjID: id, TPre: tPre2, ObjType: "CS", File: file2, Line: line2}

						if !isLeakTimerOrCtx(se) && !se.GetContainsDefault() {
							results.Result(results.CRITICAL, helper.LSelectWith,
								"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
						} else {
							results.Result(results.CRITICAL, helper.LUnknown,
								"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
						}

						foundPartner = true
					}
				}
			}
			if cl, ok := data.CloseData[id]; ok {
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
					RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: cl.GetRoutine(), ObjID: id, TPre: tPre2, ObjType: "CS", File: file2, Line: line2}

				if !isLeakTimerOrCtx(se) && !se.GetContainsDefault() {
					results.Result(results.CRITICAL, helper.LSelectWith,
						"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
				} else {
					results.Result(results.CRITICAL, helper.LUnknown,
						"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
				}

				foundPartner = true
			}
		}
	}

	if !foundPartner {
		for i, id := range ids {
			// add all select operations to leaking Channels,
			data.LeakingChannels[id] = append(data.LeakingChannels[id], data.VectorClockTID2{
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
// It adds a found leak to the results
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func CheckForLeakMutex(mu *trace.ElementMutex) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	id := mu.GetID()
	opM := mu.GetOpM()

	if _, ok := data.MostRecentAcquireTotal[id]; !ok {
		return
	}

	elem := data.MostRecentAcquireTotal[id].Elem

	file2, line2, tPre2 := elem.GetFile(), elem.GetLine(), elem.GetTPre()

	objType1 := "M"
	switch opM {
	case trace.LockOp: // lock
		objType1 += "L"
	case trace.RLockOp: // rlock
		objType1 += "R"
	default: // only lock and rlock can lead to leak
		return
	}

	objType2 := "M"
	if data.MostRecentAcquireTotal[id].Val == int(trace.LockOp) { // lock
		objType2 += "L"
	} else if data.MostRecentAcquireTotal[id].Val == int(trace.RLockOp) { // rlock
		objType2 += "R"
	} else if data.MostRecentAcquireTotal[id].Val == int(trace.TryLockOp) { // TryLock
		objType2 += "T"
	} else if data.MostRecentAcquireTotal[id].Val == int(trace.TryRLockOp) { // TryRLock
		objType2 += "Y"
	} else { // only lock and rlock can lead to leak
		return
	}

	arg1 := results.TraceElementResult{
		RoutineID: mu.GetRoutine(), ObjID: id, TPre: mu.GetTPre(), ObjType: objType1, File: mu.GetFile(), Line: mu.GetLine()}

	arg2 := results.TraceElementResult{
		RoutineID: data.MostRecentAcquireTotal[id].Elem.GetRoutine(), ObjID: id, TPre: tPre2, ObjType: objType2, File: file2, Line: line2}

	results.Result(results.CRITICAL, helper.LMutex,
		"mutex", []results.ResultElem{arg1}, "last", []results.ResultElem{arg2})
}

// AddMostRecentAcquireTotal adds the most recent acquire operation for a mutex
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//   - vc VectorClock: The vector clock of the operation
//   - op int: The operation on the mutex
func AddMostRecentAcquireTotal(mu *trace.ElementMutex, vc *clock.VectorClock, op int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	data.MostRecentAcquireTotal[mu.GetID()] = data.ElemWithVcVal{Elem: mu, Vc: vc.Copy(), Val: op}
}

// CheckForLeakWait is run for wait group operation without a post event.
// It adds an error to the results
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

	arg := results.TraceElementResult{
		RoutineID: wa.GetRoutine(), ObjID: wa.GetID(), TPre: tPre, ObjType: "WW", File: file, Line: line}

	results.Result(results.CRITICAL, helper.LWaitGroup,
		"wait", []results.ResultElem{arg}, "", []results.ResultElem{})
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

	arg := results.TraceElementResult{
		RoutineID: co.GetRoutine(), ObjID: co.GetID(), TPre: tPre, ObjType: "DW", File: file, Line: line}

	results.Result(results.CRITICAL, helper.LCond,
		"cond", []results.ResultElem{arg}, "", []results.ResultElem{})
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

	for routine, tr := range data.MainTrace.GetTraces() {
		if len(tr) == 0 {
			continue
		}

		lastElem := tr[len(tr)-1]
		switch lastElem.(type) {
		case *trace.ElementRoutineEnd:
			continue
		}

		lastTPost := lastElem.GetTPost()

		leakType := helper.LUnknown
		objectType := "XX"
		// do not record extra if a leak with a blocked operation is present
		// if simple, find the type of blocking
		if lastTPost == 0 {
			if simple {
				ot := lastElem.GetObjType(true)
				objectType = ot
				switch ot {
				case "CS", "CR":
					c := lastElem.(*trace.ElementChannel)
					if c.GetID() == -1 {
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
						objectType = "XX"
					} else {
						leakType = helper.LSelectWithout
					}
				default:
					objectType = "XX"
				}
			} else {
				continue
			}
		}

		arg := results.TraceElementResult{
			RoutineID: routine, ObjID: -1, TPre: lastElem.GetTPre(),
			ObjType: objectType, File: lastElem.GetFile(), Line: lastElem.GetLine(),
		}

		results.Result(results.CRITICAL, leakType,
			"elem", []results.ResultElem{arg}, "", []results.ResultElem{})

		res = true
	}

	return res
}

func isLeakTimerOrCtx(elem trace.Element) bool {
	switch e := elem.(type) {
	case *trace.ElementChannel:
		return chanIsTimerOrCtx(elem.GetID())
	case *trace.ElementSelect:
		for _, c := range e.GetCases() {
			if chanIsTimerOrCtx(c.GetID()) {
				return true
			}
		}
	default:
		return false
	}

	return false
}

func chanIsTimerOrCtx(id int) bool {
	pos, ok := data.NewChan[id]

	if !ok {
		return false
	}

	return strings.Contains(pos, "/src/time/") || strings.Contains(pos, "/src/context/")
}
