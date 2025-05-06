// Copyright (c) 2024 Erik Kassubek
//
// File: analysisLeak.go
// Brief: Trace analysis for routine leaks
//
// Author: Erik Kassubek
// Created: 2024-01-28
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/clock"
	"advocate/results"
	"advocate/timer"
	"advocate/trace"
	"advocate/utils"
	"strconv"
	"strings"
)

// CheckForLeakChannelStuck is run for channel operation without a post event.
// It checks if the operation has a possible communication partner in
// mostRecentSend, mostRecentReceive or closeData.
// If so, add an error or warning to the result.
// If not, add to leakingChannels, for later check.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc VectorClock: The vector clock of the operation
//   - func CheckForLeakChannelStuck(routineID int, objID int, vc clock.VectorClock, tID string, opType int, buffered bool) {
func CheckForLeakChannelStuck(ch *trace.TraceElementChannel, vc *clock.VectorClock) {
	buffered := (ch.GetQSize() != 0)
	id := ch.GetID()
	opC := ch.GetOpC()
	routine := ch.GetRoutine()

	if id == -1 {
		objType := "C"
		if opC == trace.SendOp {
			objType += "S"
		} else if opC == trace.RecvOp {
			objType += "R"
		} else {
			return // close
		}

		file, line, tPre, err := trace.InfoFromTID(ch.GetTID())
		if err != nil {
			utils.LogError("Error in infoFromTID: ", err.Error())
			return
		}

		arg1 := results.TraceElementResult{
			RoutineID: routine, ObjID: id, TPre: tPre, ObjType: objType, File: file, Line: line}

		results.Result(results.CRITICAL, utils.LNilChan,
			"Channel", []results.ResultElem{arg1}, "", []results.ResultElem{})

		return
	}

	// if !buffered {
	foundPartner := false

	if opC == trace.SendOp { // send
		for partnerRout, mrr := range mostRecentReceive {
			if _, ok := mrr[id]; ok {
				if clock.GetHappensBefore(mrr[id].Vc, vc) == clock.Concurrent {

					var bugType utils.ResultType = utils.LUnbufferedWith
					if buffered {
						bugType = utils.LBufferedWith
					}

					file1, line1, tPre1, err := trace.InfoFromTID(ch.GetTID())
					if err != nil {
						utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", ch.GetTID())
						return
					}
					file2, line2, tPre2, err := trace.InfoFromTID(mrr[id].Elem.GetTID())
					if err != nil {
						utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", mrr[id].Elem.GetTID())
						return
					}

					arg1 := results.TraceElementResult{
						RoutineID: routine, ObjID: id, TPre: tPre1, ObjType: "CS", File: file1, Line: line1}
					arg2 := results.TraceElementResult{
						RoutineID: partnerRout, ObjID: id, TPre: tPre2, ObjType: "CR", File: file2, Line: line2}

					results.Result(results.CRITICAL, bugType,
						"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})

					foundPartner = true
				}
			}
		}
	} else if opC == trace.RecvOp { // recv
		for partnerRout, mrs := range mostRecentSend {
			if _, ok := mrs[id]; ok {
				if clock.GetHappensBefore(mrs[id].Vc, vc) == clock.Concurrent {

					var bugType = utils.LUnbufferedWith
					if buffered {
						bugType = utils.LBufferedWith
					}

					file1, line1, tPre1, err1 := trace.InfoFromTID(ch.GetTID())
					if err1 != nil {
						utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", ch.GetTID())
						return
					}
					file2, line2, tPre2, err2 := trace.InfoFromTID(mrs[id].Elem.GetTID())
					if err2 != nil {
						utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", mrs[id].Elem.GetTID())
						return
					}

					arg1 := results.TraceElementResult{
						RoutineID: routine, ObjID: id, TPre: tPre1, ObjType: "CR", File: file1, Line: line1}
					arg2 := results.TraceElementResult{
						RoutineID: partnerRout, ObjID: id, TPre: tPre2, ObjType: "CS", File: file2, Line: line2}

					results.Result(results.CRITICAL, bugType,
						"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})

					foundPartner = true
				}
			}
		}

	}

	if !foundPartner {
		leakingChannels[id] = append(leakingChannels[id], VectorClockTID2{routine, id, vc, ch.GetTID(), int(opC), -1, buffered, false, 0})
	}
}

// CheckForLeakChannelRun is run for channel operation with a post event.
// It checks if the operation would be possible communication partner for a
// stuck operation in leakingChannels.
// If so, add an error or warning to the result and remove the stuck operation.
//
// Parameter:
//   - routineID int: The routine id
//   - objID int: The channel id
//   - vc VectorClock: The vector clock of the operation
//   - opType int: An identifier for the type of the operation (send = 0, recv = 1, close = 2)
//   - buffered bool: If the channel is buffered
func CheckForLeakChannelRun(routineID int, objID int, elemVc elemWithVc, opType int, buffered bool) bool {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	res := false
	if opType == 0 || opType == 2 { // send or close
		for i, vcTID2 := range leakingChannels[objID] {
			if vcTID2.val != 1 {
				continue
			}

			if clock.GetHappensBefore(vcTID2.vc, elemVc.vc) == clock.Concurrent {
				var bugType = utils.LUnbufferedWith
				if buffered {
					bugType = utils.LBufferedWith
				}

				file1, line1, tPre1, err1 := trace.InfoFromTID(vcTID2.tID) // leaking
				if err1 != nil {
					utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", vcTID2.tID)
					return res
				}

				elem2 := elemVc.elem

				objType := "C"
				if opType == 0 {
					objType += "S"
				} else {
					objType += "C"
				}

				arg1 := results.TraceElementResult{
					RoutineID: routineID, ObjID: objID, TPre: tPre1, ObjType: "CR", File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: vcTID2.routine, ObjID: objID, TPre: elem2.GetTPre(), ObjType: objType, File: elem2.GetFile(), Line: elem2.GetLine()}

				results.Result(results.CRITICAL, bugType,
					"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})

				res = true

				// remove the stuck operation from the list. If it is a select, remove all operations with the same val
				if vcTID2.val == -1 {
					leakingChannels[objID] = append(leakingChannels[objID][:i], leakingChannels[objID][i+1:]...)
				} else {
					for j, vcTID3 := range leakingChannels[objID] {
						if vcTID3.val == vcTID2.val {
							leakingChannels[objID] = append(leakingChannels[objID][:j], leakingChannels[objID][j+1:]...)
						}
					}
				}
			}
		}
	} else if opType == 1 { // recv
		for i, vcTID2 := range leakingChannels[objID] {
			objType := "C"
			if vcTID2.val == 0 {
				objType += "S"
			} else if vcTID2.val == 2 {
				objType += "C"
			} else {
				continue
			}

			if clock.GetHappensBefore(vcTID2.vc, elemVc.vc) == clock.Concurrent {

				var bugType = utils.LUnbufferedWith
				if buffered {
					bugType = utils.LBufferedWith
				}

				file1, line1, tPre1, err1 := trace.InfoFromTID(vcTID2.tID) // leaking
				if err1 != nil {
					utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", vcTID2.tID)
					return res
				}

				elem2 := elemVc.elem

				arg1 := results.TraceElementResult{
					RoutineID: routineID, ObjID: objID, TPre: tPre1, ObjType: objType, File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: vcTID2.routine, ObjID: objID, TPre: elem2.GetTPre(), ObjType: "CR", File: elem2.GetFile(), Line: elem2.GetLine()}

				results.Result(results.CRITICAL, bugType,
					"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})

				res = true

				// remove the stuck operation from the list. If it is a select, remove all operations with the same val
				if vcTID2.val == -1 {
					leakingChannels[objID] = append(leakingChannels[objID][:i], leakingChannels[objID][i+1:]...)
				} else {
					for j, vcTID3 := range leakingChannels[objID] {
						if vcTID3.val == vcTID2.val {
							leakingChannels[objID] = append(leakingChannels[objID][:j], leakingChannels[objID][j+1:]...)
						}
					}
				}
			}
		}
	}
	return res
}

// After all operations have been analyzed, check if there are still leaking
// operations without a possible partner.
func checkForLeak() {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	// channel
	for _, vcTIDs := range leakingChannels {
		buffered := false
		for _, vcTID := range vcTIDs {
			if vcTID.tID == "" {
				continue
			}

			found := false
			var partner allSelectCase
			for _, c := range selectCases {
				if c.chanID != vcTID.id {
					continue
				}

				if (c.send && vcTID.typeVal == 0) || (!c.send && vcTID.typeVal == 1) {
					continue
				}

				hb := clock.GetHappensBefore(c.elem.vc, vcTID.vc)
				if hb == clock.Concurrent {
					found = true
					if c.buffered {
						buffered = true
					}
					partner = c
					break
				}

				if c.buffered {
					if (c.send && hb == clock.Before) || (!c.send && hb == clock.After) {
						found = true
						buffered = true
						partner = c
						break
					}
				}
			}

			if found {
				file1, line1, tPre1, err := trace.InfoFromTID(vcTID.tID)
				if err != nil {
					utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", vcTID.tID)
					continue
				}

				elem2 := partner.elem.elem
				file2 := elem2.GetFile()
				line2 := elem2.GetLine()
				tPre2 := elem2.GetTPre()

				if vcTID.sel {

					arg1 := results.TraceElementResult{ // select
						RoutineID: vcTID.routine, ObjID: vcTID.id, TPre: tPre1, ObjType: "SS", File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.sel.GetID(), TPre: tPre2, ObjType: "SS", File: file2, Line: line2}

					results.Result(results.CRITICAL, utils.LSelectWith,
						"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
				} else {
					obType := "C"
					if vcTID.typeVal == 0 {
						obType += "S"
					} else {
						obType += "R"
					}

					var bugType utils.ResultType = utils.LUnbufferedWith
					if buffered {
						bugType = utils.LBufferedWith
					}

					arg1 := results.TraceElementResult{ // channel
						RoutineID: vcTID.routine, ObjID: vcTID.id, TPre: tPre1, ObjType: obType, File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.sel.GetID(), TPre: tPre2, ObjType: "SS", File: file2, Line: line2}

					results.Result(results.CRITICAL, bugType,
						"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
				}

			} else {
				if vcTID.sel {
					file, line, tPre, err := trace.InfoFromTID(vcTID.tID)
					if err != nil {
						utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", vcTID.tID)
						continue
					}

					arg1 := results.TraceElementResult{
						RoutineID: vcTID.routine, ObjID: vcTID.selID, TPre: tPre, ObjType: "SS", File: file, Line: line}

					results.Result(results.CRITICAL, utils.LSelectWithout,
						"select", []results.ResultElem{arg1}, "", []results.ResultElem{})

				} else {
					objType := "C"
					if vcTID.typeVal == 0 {
						objType += "S"
					} else {
						objType += "R"
					}

					file, line, tPre, err := trace.InfoFromTID(vcTID.tID)
					if err != nil {
						utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", vcTID.tID)
						continue
					}

					arg1 := results.TraceElementResult{
						RoutineID: vcTID.routine, ObjID: vcTID.id, TPre: tPre, ObjType: objType, File: file, Line: line}

					var bugType utils.ResultType = utils.LUnbufferedWithout
					if buffered {
						bugType = utils.LBufferedWithout
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
// mostRecentSend, mostRecentReceive or closeData.
// If so, add an error or warning to the result.
// If not, add all elements to leakingChannels, for later check.
//
// Parameter:
//   - se *TraceElementSelect: The trace element
//   - ids int: The channel ids
//   - buffered []bool: If the channels are buffered
//   - vc *VectorClock: The vector clock of the operation
//   - opTypes []int: An identifier for the type of the operations (send = 0, recv = 1)
func CheckForLeakSelectStuck(se *trace.TraceElementSelect, ids []int, buffered []bool, vc *clock.VectorClock, opTypes []int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	foundPartner := false

	routine := se.GetRoutine()
	id := se.GetID()
	tPre := se.GetTPre()

	if len(ids) == 0 {
		file, line, _, err := trace.InfoFromTID(se.GetTID())
		if err != nil {
			utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
			return
		}

		arg1 := results.TraceElementResult{
			RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file, Line: line}

		results.Result(results.CRITICAL, utils.LSelectWithout,
			"select", []results.ResultElem{arg1}, "", []results.ResultElem{})

		return
	}

	for i, id := range ids {
		if opTypes[i] == 0 { // send
			for routinePartner, mrr := range mostRecentReceive {
				if recv, ok := mrr[id]; ok {
					if clock.GetHappensBefore(vc, mrr[id].Vc) == clock.Concurrent {
						file1, line1, _, err1 := trace.InfoFromTID(se.GetTID()) // select
						if err1 != nil {
							utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
							return
						}
						file2, line2, tPre2, err2 := trace.InfoFromTID(recv.Elem.GetTID()) // partner
						if err2 != nil {
							utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", recv.Elem.GetTID())
							return
						}

						arg1 := results.TraceElementResult{
							RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file1, Line: line1}
						arg2 := results.TraceElementResult{
							RoutineID: routinePartner, ObjID: id, TPre: tPre2, ObjType: "CR", File: file2, Line: line2}

						results.Result(results.CRITICAL, utils.LSelectWith,
							"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})
						foundPartner = true
					}
				}
			}
		} else if opTypes[i] == 1 { // recv
			for routinePartner, mrs := range mostRecentSend {
				if send, ok := mrs[id]; ok {
					if clock.GetHappensBefore(vc, mrs[id].Vc) == clock.Concurrent {
						file1, line1, _, err1 := trace.InfoFromTID(se.GetTID()) // select
						if err1 != nil {
							utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
							return
						}
						file2, line2, tPre2, err2 := trace.InfoFromTID(send.Elem.GetTID()) // partner
						if err2 != nil {
							utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", send.Elem.GetTID())
							return
						}

						arg1 := results.TraceElementResult{
							RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file1, Line: line1}
						arg2 := results.TraceElementResult{
							RoutineID: routinePartner, ObjID: id, TPre: tPre2, ObjType: "CS", File: file2, Line: line2}

						results.Result(results.CRITICAL, utils.LSelectWith,
							"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})

						foundPartner = true
					}
				}
			}
			if cl, ok := closeData[id]; ok {
				file1, line1, _, err1 := trace.InfoFromTID(se.GetTID()) // select
				if err1 != nil {
					utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", se.GetTID())
					return
				}
				file2, line2, tPre2, err2 := trace.InfoFromTID(cl.GetTID()) // partner
				if err2 != nil {
					utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", cl.GetTID())
					return
				}

				arg1 := results.TraceElementResult{
					RoutineID: routine, ObjID: id, TPre: tPre, ObjType: "SS", File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: cl.GetRoutine(), ObjID: id, TPre: tPre2, ObjType: "CS", File: file2, Line: line2}

				results.Result(results.CRITICAL, utils.LSelectWith,
					"select", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2})

				foundPartner = true
			}
		}
	}

	if !foundPartner {
		for i, id := range ids {
			// add all select operations to leaking Channels,
			leakingChannels[id] = append(leakingChannels[id], VectorClockTID2{routine, id, vc, se.GetTID(), opTypes[i], tPre, buffered[i], true, id})
		}
	}
}

// CheckForLeakMutex is run for mutex operation without a post event.
// It adds a found leak to the results
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func CheckForLeakMutex(mu *trace.TraceElementMutex) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	id := mu.GetID()
	opM := mu.GetOpM()

	if _, ok := mostRecentAcquireTotal[id]; !ok {
		return
	}

	elem := mostRecentAcquireTotal[id].Elem

	file2, line2, tPre2 := elem.GetFile(), elem.GetLine(), elem.GetTPre()

	objType1 := "M"
	if opM == trace.LockOp { // lock
		objType1 += "L"
	} else if opM == trace.RLockOp { // rlock
		objType1 += "R"
	} else { // only lock and rlock can lead to leak
		return
	}

	objType2 := "M"
	if mostRecentAcquireTotal[id].Val == int(trace.LockOp) { // lock
		objType2 += "L"
	} else if mostRecentAcquireTotal[id].Val == int(trace.RLockOp) { // rlock
		objType2 += "R"
	} else if mostRecentAcquireTotal[id].Val == int(trace.TryLockOp) { // TryLock
		objType2 += "T"
	} else if mostRecentAcquireTotal[id].Val == int(trace.TryRLockOp) { // TryRLock
		objType2 += "Y"
	} else { // only lock and rlock can lead to leak
		return
	}

	arg1 := results.TraceElementResult{
		RoutineID: mu.GetRoutine(), ObjID: id, TPre: mu.GetTPre(), ObjType: objType1, File: mu.GetFile(), Line: mu.GetLine()}

	arg2 := results.TraceElementResult{
		RoutineID: mostRecentAcquireTotal[id].Elem.GetRoutine(), ObjID: id, TPre: tPre2, ObjType: objType2, File: file2, Line: line2}

	results.Result(results.CRITICAL, utils.LMutex,
		"mutex", []results.ResultElem{arg1}, "last", []results.ResultElem{arg2})
}

// Add the most recent acquire operation for a mutex
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//   - vc VectorClock: The vector clock of the operation
//   - op int: The operation on the mutex
func addMostRecentAcquireTotal(mu *trace.TraceElementMutex, vc *clock.VectorClock, op int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	mostRecentAcquireTotal[mu.GetID()] = ElemWithVcVal{Elem: mu, Vc: vc.Copy(), Val: op}
}

// CheckForLeakWait is run for wait group operation without a post event.
// It adds an error to the results
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func CheckForLeakWait(wa *trace.TraceElementWait) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	file, line, tPre, err := trace.InfoFromTID(wa.GetTID())
	if err != nil {
		utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", wa.GetTID())
		return
	}

	arg := results.TraceElementResult{
		RoutineID: wa.GetRoutine(), ObjID: wa.GetID(), TPre: tPre, ObjType: "WW", File: file, Line: line}

	results.Result(results.CRITICAL, utils.LWaitGroup,
		"wait", []results.ResultElem{arg}, "", []results.ResultElem{})
}

// CheckForLeakCond is run for conditional variable operation without a post
// event. It adds a leak to the results
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CheckForLeakCond(co *trace.TraceElementCond) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	file, line, tPre, err := trace.InfoFromTID(co.GetTID())
	if err != nil {
		utils.LogErrorf("Error in trace.InfoFromTID(%s)\n", co.GetTID())
		return
	}

	arg := results.TraceElementResult{
		RoutineID: co.GetRoutine(), ObjID: co.GetID(), TPre: tPre, ObjType: "DW", File: file, Line: line}

	results.Result(results.CRITICAL, utils.LCond,
		"cond", []results.ResultElem{arg}, "", []results.ResultElem{})
}

// Iterate over all routines and check if the routines finished.
// Only record leaking routines, that don't have a leaking element (tPost = 0)
// as its last element, since they are recorded separately
func checkForStuckRoutine() {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	for routine, tr := range MainTrace.GetTraces() {
		if len(tr) < 1 {
			continue
		}

		lastElem := tr[len(tr)-1]
		switch lastElem.(type) {
		case *trace.TraceElementRoutineEnd:
			continue
		}

		// do not record extra if a leak with a blocked operation is present
		if len(tr) > 0 && tr[len(tr)-1].GetTPost() == 0 {
			continue
		}

		file := ""
		line := -1
		if p, ok := allForks[routine]; ok {
			pos := p.GetPos()
			posSplit := strings.Split(pos, ":")
			if len(posSplit) == 2 {
				file = posSplit[0]
				line, _ = strconv.Atoi(posSplit[1])
			}
		}

		arg := results.TraceElementResult{
			RoutineID: routine, ObjID: -1, TPre: lastElem.GetTPre(),
			ObjType: "RE", File: file, Line: line,
		}

		results.Result(results.CRITICAL, utils.LWithoutBlock,
			"fork", []results.ResultElem{arg}, "", []results.ResultElem{})
	}
}
