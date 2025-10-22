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
	"advocate/analysis/baseA"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/timer"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

var leaks = make(map[int]TERLeak, 0)
var deadlocks = make([]results.ResultElem, 0)

func PartialDeadlocks() error {
	log.Info("Check for actual partial deadlocks")
	output := filepath.Join(paths.ProgDir, paths.NameOutput)

	file, err := os.Open(output)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "DEADLOCK@") {
			err := readDeadlock(line)
			if err != nil {
				log.Errorf(err.Error())
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	reportDeadlocks()
	reportNonDeadlockLeaks()

	log.Info("Finish check for actual partial deadlocks")

	return nil
}

func readDeadlock(deadlock string) error {
	fields := strings.Split(deadlock, "@")

	if len(fields) != 4 {
		return fmt.Errorf("Could not process deadlock %s", deadlock)
	}

	routineID, err := strconv.Atoi(fields[1])
	if err != nil {
		return err
	}

	file := "-"
	line := -1

	path := fields[2]
	if path != "-" {
		posFields := strings.Split(path, ":")
		if len(posFields) != 2 {
			return fmt.Errorf("Could not process deadlock position %s", path)
		}

		file = posFields[0]
		line, err = strconv.Atoi(posFields[1])
		if err != nil {
			return err
		}
	}

	var objRes results.ResultElem
	objResSet := false

	if obj, ok := leaks[routineID]; ok {
		objRes = obj.arg1[0]
		objResSet = true
		delete(leaks, routineID)
	}

	if !objResSet {
		objRes = results.TraceElementResult{
			RoutineID: routineID,
			ObjID:     -1,
			TPre:      -1,
			ObjType:   getObjectType(fields[3]),
			File:      file,
			Line:      line,
		}
	}

	deadlocks = append(deadlocks, objRes)

	return nil
}

func getObjectType(val string) trace.OperationType {
	switch val {
	case "chan:recvOnNil", "chan:revc":
		return trace.ChannelRecv
	case "chan:sendOnNil", "chan:send":
		return trace.ChannelSend
	case "select:select", "select:withoutCases":
		return trace.SelectOp
	case "cond:wait":
		return trace.CondWait
	case "mutex:lock", "rwmutex:lock":
		return trace.MutexLock
	case "rwmutex:rlock":
		return trace.MutexRLock
	case "waitGroup:wait":
		return trace.WaitWait
	}
	return trace.None
}

// reportDeadlocks creates a result for all elements that are in a deadlock
func reportDeadlocks() {
	if len(deadlocks) == 0 {
		return
	}

	results.Result(results.CRITICAL, helper.ADeadlock,
		"Blocked", deadlocks, "", []results.ResultElem{})
}

// reportNonDeadlockLeaks creates results for all elements that have a leek
// without being in a deadlock
func reportNonDeadlockLeaks() {
	for _, leak := range leaks {
		results.Result(results.CRITICAL, leak.resultType,
			leak.argType1, leak.arg2, leak.argType1, leak.arg2)
	}
}

// CheckForLeakChannelStuck is run for channel operation without a post event.
// It checks if the operation has a possible communication partner in
// baseA.MostRecentSend, baseA.MostRecentReceive or baseA.ClosebaseA.
// If so, add an the data to leaks
// If not, add to baseA.LeakingChannels, for later check.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc VectorClock: The vector clock of the operation
//   - func CheckForLeakChannelStuck(routineID int, objID int, vc clock.VectorClock, tID string, opType int, buffered bool) {
func CheckForLeakChannelStuck(ch *trace.ElementChannel, vc *clock.VectorClock) {
	buffered := (ch.GetQSize() != 0)
	id := ch.GetID()
	opC := ch.GetType(true)
	routine := ch.GetRoutine()

	if id == -1 {
		if opC == trace.ChannelClose {
			return // close
		}

		arg1 := results.TraceElementResult{
			RoutineID: routine, ObjID: id, TPre: ch.GetTPre(), ObjType: opC, File: ch.GetFile(), Line: ch.GetLine()}

		leaks[routine] = TERLeak{helper.LNilChan,
			"Channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}

		return
	}

	// if !buffered {
	foundPartner := false

	switch opC {
	case trace.ChannelSend:
		for partnerRout, mrr := range baseA.MostRecentReceive {
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

					timer, ctx := isLeakTimerOrCtx(ch)
					if !timer {
						if !ctx {
							leaks[routine] = TERLeak{bugType,
								"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
							foundPartner = true
						} else {
							leaks[routine] = TERLeak{helper.LContext,
								"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}
					} else {
						leaks[routine] = TERLeak{helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					}
				}
			}
		}
	case trace.ChannelRecv: // recv
		for partnerRout, mrs := range baseA.MostRecentSend {
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

					timer, ctx := isLeakTimerOrCtx(ch)
					if !timer {
						if !ctx {
							leaks[routine] = TERLeak{bugType,
								"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
							foundPartner = true
						} else {
							leaks[routine] = TERLeak{helper.LContext,
								"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
						}
					} else {
						leaks[routine] = TERLeak{helper.LUnknown,
							"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					}
				}
			}
		}

	}

	if !foundPartner {
		baseA.LeakingChannels[id] = append(baseA.LeakingChannels[id], baseA.VectorClockTID2{
			Routine:  routine,
			ID:       id,
			Vc:       vc,
			TID:      ch.GetTID(),
			TypeVal:  opC,
			Val:      -1,
			Buffered: buffered,
			Sel:      false,
			SelID:    0,
		})
	}
}

// CheckForLeakChannelRun is run for channel operation with a post event.
// It checks if the operation would be possible communication partner for a
// stuck operation in baseA.LeakingChannels.
// If so, add the if to leaks and remove the stuck operation.
//
// Parameter:
//   - routineID int: The routine id
//   - objID int: The channel id
//   - vc VectorClock: The vector clock of the operation
//   - opType trace.ObjectType: The type of operation
//   - buffered bool: If the channel is buffered
func CheckForLeakChannelRun(routineID int, objID int, elemVc baseA.ElemWithVc, opType trace.OperationType, buffered bool) bool {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	res := false
	if opType == trace.ChannelSend || opType == trace.ChannelClose {
		for i, vcTID2 := range baseA.LeakingChannels[objID] {
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

				arg1 := results.TraceElementResult{
					RoutineID: routineID, ObjID: objID, TPre: tPre1, ObjType: "CR", File: file1, Line: line1}
				arg2 := results.TraceElementResult{
					RoutineID: vcTID2.Routine, ObjID: objID, TPre: elem2.GetTPre(), ObjType: opType, File: elem2.GetFile(), Line: elem2.GetLine()}

				timer, ctx := chanIsTimerOrCtx(objID)
				if timer {
					leaks[routineID] = TERLeak{helper.LUnknown,
						"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					continue
				}
				if ctx {
					leaks[routineID] = TERLeak{helper.LContext,
						"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					continue
				}

				leaks[routineID] = TERLeak{bugType,
					"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}

				res = true

				// remove the stuck operation from the list. If it is a select, remove all operations with the same val
				if vcTID2.Val == -1 {
					baseA.LeakingChannels[objID] = append(baseA.LeakingChannels[objID][:i], baseA.LeakingChannels[objID][i+1:]...)
				} else {
					for j, vcTID3 := range baseA.LeakingChannels[objID] {
						if vcTID3.Val == vcTID2.Val {
							baseA.LeakingChannels[objID] = append(baseA.LeakingChannels[objID][:j], baseA.LeakingChannels[objID][j+1:]...)
						}
					}
				}
			}
		}
	} else if opType == trace.ChannelRecv {
		for i, vcTID2 := range baseA.LeakingChannels[objID] {
			objType := trace.Channel
			switch vcTID2.Val {
			case 0:
				objType = trace.ChannelSend
			case 2:
				objType += trace.ChannelClose
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

				timer, ctx := chanIsTimerOrCtx(objID)
				if timer {
					leaks[routineID] = TERLeak{helper.LUnknown,
						"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					continue
				}
				if ctx {
					leaks[routineID] = TERLeak{helper.LContext,
						"channel", []results.ResultElem{arg1}, "", []results.ResultElem{}}
					continue
				}

				leaks[routineID] = TERLeak{bugType,
					"channel", []results.ResultElem{arg1}, "partner", []results.ResultElem{arg2}}
				res = true

				// remove the stuck operation from the list. If it is a select, remove all operations with the same val
				if vcTID2.Val == -1 {
					baseA.LeakingChannels[objID] = append(baseA.LeakingChannels[objID][:i], baseA.LeakingChannels[objID][i+1:]...)
				} else {
					for j, vcTID3 := range baseA.LeakingChannels[objID] {
						if vcTID3.Val == vcTID2.Val {
							baseA.LeakingChannels[objID] = append(baseA.LeakingChannels[objID][:j], baseA.LeakingChannels[objID][j+1:]...)
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
	for _, vcTIDs := range baseA.LeakingChannels {
		buffered := false
		for _, vcTID := range vcTIDs {
			if vcTID.TID == "" {
				continue
			}

			routineID := vcTID.Routine

			found := false
			var partner baseA.AllSelectCase
			for _, c := range baseA.SelectCases {
				if c.ChanID != vcTID.ID {
					continue
				}

				if (c.Send && vcTID.TypeVal == trace.ChannelSend) || (!c.Send && vcTID.TypeVal == trace.ChannelRecv) {
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
						RoutineID: routineID, ObjID: vcTID.ID, TPre: tPre1, ObjType: "SS", File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.Sel.GetID(), TPre: tPre2, ObjType: "SS", File: file2, Line: line2}

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
						RoutineID: routineID, ObjID: vcTID.ID, TPre: tPre1, ObjType: vcTID.TypeVal, File: file1, Line: line1}

					arg2 := results.TraceElementResult{ // select
						RoutineID: elem2.GetRoutine(), ObjID: partner.Sel.GetID(), TPre: tPre2, ObjType: "SS", File: file2, Line: line2}

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
						RoutineID: vcTID.Routine, ObjID: vcTID.SelID, TPre: tPre, ObjType: "SS", File: file, Line: line}

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
						RoutineID: vcTID.Routine, ObjID: vcTID.ID, TPre: tPre, ObjType: vcTID.TypeVal, File: file, Line: line}

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
func CheckForLeakSelectStuck(se *trace.ElementSelect, ids []int, buffered []bool, vc *clock.VectorClock, opTypes []trace.OperationType) {
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
			for routinePartner, mrr := range baseA.MostRecentReceive {
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
			for routinePartner, mrs := range baseA.MostRecentSend {
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
			if cl, ok := baseA.CloseData[id]; ok {
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
			baseA.LeakingChannels[id] = append(baseA.LeakingChannels[id], baseA.VectorClockTID2{
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

	id := mu.GetID()
	opM := mu.GetType(true)
	routineID := mu.GetRoutine()

	if _, ok := baseA.MostRecentAcquireTotal[id]; !ok {
		return
	}

	elem := baseA.MostRecentAcquireTotal[id].Elem

	file2, line2, tPre2 := elem.GetFile(), elem.GetLine(), elem.GetTPre()

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
		RoutineID: mu.GetRoutine(), ObjID: id, TPre: mu.GetTPre(), ObjType: opM, File: mu.GetFile(), Line: mu.GetLine()}

	arg2 := results.TraceElementResult{
		RoutineID: elem.GetRoutine(), ObjID: id, TPre: tPre2, ObjType: objType2, File: file2, Line: line2}

	leaks[routineID] = TERLeak{helper.LMutex,
		"mutex", []results.ResultElem{arg1}, "last", []results.ResultElem{arg2}}
}

// AddMostRecentAcquireTotal adds the most recent acquire operation for a mutex
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//   - vc VectorClock: The vector clock of the operation
//   - op int: The operation on the mutex
func AddMostRecentAcquireTotal(mu *trace.ElementMutex, vc *clock.VectorClock) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	baseA.MostRecentAcquireTotal[mu.GetID()] = baseA.ElemWithVc{Elem: mu, Vc: vc.Copy()}
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
		RoutineID: routineID, ObjID: wa.GetID(), TPre: tPre, ObjType: "WW", File: file, Line: line}

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
		RoutineID: routineID, ObjID: co.GetID(), TPre: tPre, ObjType: "DW", File: file, Line: line}

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

	for routine, tr := range baseA.MainTrace.GetTraces() {
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
			RoutineID: routine, ObjID: lastElem.GetID(), TPre: lastElem.GetTPre(),
			ObjType: objectType, File: lastElem.GetFile(), Line: lastElem.GetLine(),
		}

		timer, ctx := isLeakTimerOrCtx(lastElem)

		if leakType == helper.LUnknown {
			leaks[routine] = TERLeak{leakType,
				"elem", []results.ResultElem{arg}, "", []results.ResultElem{}}
		} else if timer {
			leaks[routine] = TERLeak{helper.LUnknown,
				"elem", []results.ResultElem{arg}, "", []results.ResultElem{}}
		} else if ctx {
			leaks[routine] = TERLeak{helper.LContext,
				"elem", []results.ResultElem{arg}, "", []results.ResultElem{}}
		} else {
			leaks[routine] = TERLeak{leakType,
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
		return chanIsTimerOrCtx(elem.GetID())
	case *trace.ElementSelect:
		for _, c := range e.GetCases() {
			ti, co := chanIsTimerOrCtx(c.GetID())
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
	pos, ok := baseA.NewChan[id]

	if !ok {
		return false, false
	}

	return strings.Contains(pos, "/src/time/"), strings.Contains(pos, "/src/context/")
}
