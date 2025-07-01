// Copyright (c) 2024 Erik Kassubek
//
// File: analysisClose.go
// Brief: Trace analysis for send, receive and close on closed channel
//
// Author: Erik Kassubek
// Created: 2024-01-04
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/clock"
	"advocate/analysis/data"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// checkForCommunicationOnClosedChannel checks if a send or receive on a
// closed channel is possible.
// It it is possible, print a warning or error.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func checkForCommunicationOnClosedChannel(ch *trace.TraceElementChannel) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := ch.GetID()

	// check if there is an earlier send, that could happen concurrently to close
	if data.AnalysisCases["sendOnClosed"] && data.HasSend[id] {
		for routine, mrs := range data.MostRecentSend {
			happensBefore := clock.GetHappensBefore(mrs[id].Vc, data.CloseData[id].GetVC())

			if mrs[id].Elem != nil && mrs[id].Elem.GetTID() != "" && happensBefore != clock.Before {

				file1, line1, tPre1, err := trace.InfoFromTID(mrs[id].Elem.GetTID()) // send
				if err != nil {
					log.Error(err.Error())
					return
				}

				file2, line2, tPre2, err := trace.InfoFromTID(ch.GetTID()) // close
				if err != nil {
					log.Error(err.Error())
					return
				}

				arg1 := results.TraceElementResult{ // send
					RoutineID: routine,
					ObjID:     id,
					TPre:      tPre1,
					ObjType:   "CS",
					File:      file1,
					Line:      line1,
				}

				arg2 := results.TraceElementResult{ // close
					RoutineID: data.CloseData[ch.GetID()].GetRoutine(),
					ObjID:     id,
					TPre:      tPre2,
					ObjType:   "CC",
					File:      file2,
					Line:      line2,
				}

				results.Result(results.CRITICAL, helper.PSendOnClosed,
					"send", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
			}
		}
	}
	// check if there is an earlier receive, that could happen concurrently to close
	if data.AnalysisCases["receiveOnClosed"] && data.HasReceived[id] {
		for routine, mrr := range data.MostRecentReceive {
			happensBefore := clock.GetHappensBefore(data.CloseData[id].GetVC(), mrr[id].Vc)
			if mrr[id].Elem != nil && mrr[id].Elem.GetTID() != "" && (happensBefore == clock.Concurrent || happensBefore == clock.Before) {

				file1, line1, tPre1, err := trace.InfoFromTID(mrr[id].Elem.GetTID()) // recv
				if err != nil {
					log.Error(err.Error())
					return
				}

				file2, line2, tPre2, err := trace.InfoFromTID(ch.GetTID()) // close
				if err != nil {
					log.Error(err.Error())
					return
				}

				arg1 := results.TraceElementResult{ // recv
					RoutineID: routine,
					ObjID:     id,
					TPre:      tPre1,
					ObjType:   "CR",
					File:      file1,
					Line:      line1,
				}

				arg2 := results.TraceElementResult{ // close
					RoutineID: data.CloseData[id].GetRoutine(),
					ObjID:     id,
					TPre:      tPre2,
					ObjType:   "CC",
					File:      file2,
					Line:      line2,
				}

				results.Result(results.WARNING, helper.PRecvOnClosed,
					"recv", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
			}
		}

	}

}

// foundSendOnClosedChannel is called, id an actual send on closed was found.
//
// Parameter:
//   - elem TraceElement: the send/select elem
//   - id int: id of the channel
//   - actual bool: set actual to true it the panic occurred, set to false if it is in an not triggered select case
func foundSendOnClosedChannel(elem trace.TraceElement, actual bool) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := elem.GetID()

	if _, ok := data.CloseData[id]; !ok {
		return
	}

	closeElem := data.CloseData[id]
	fileSend := elem.GetFile()

	if fileSend == "" || fileSend == "\n" {
		return
	}

	arg1 := results.TraceElementResult{ // send
		RoutineID: elem.GetRoutine(),
		ObjID:     id,
		TPre:      elem.GetTPre(),
		ObjType:   "CS",
		File:      fileSend,
		Line:      elem.GetLine(),
	}

	arg2 := results.TraceElementResult{ // close
		RoutineID: data.CloseData[id].GetRoutine(),
		ObjID:     id,
		TPre:      closeElem.GetTPre(),
		ObjType:   "CC",
		File:      closeElem.GetFile(),
		Line:      closeElem.GetLine(),
	}

	if actual {
		results.Result(results.CRITICAL, helper.ASendOnClosed,
			"send", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	} else {
		results.Result(results.CRITICAL, helper.PSendOnClosed,
			"send", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	}

}

// foundReceiveOnClosedChannel log the detection of an actual receive on a closed channel
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func foundReceiveOnClosedChannel(ch *trace.TraceElementChannel, actual bool) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := ch.GetID()

	if _, ok := data.CloseData[id]; !ok {
		return
	}

	posClose := data.CloseData[id].GetTID()
	if posClose == "" || ch.GetTID() == "" || posClose == "\n" || ch.GetTID() == "\n" {
		return
	}

	file1, line1, tPre1, err := trace.InfoFromTID(ch.GetTID())
	if err != nil {
		log.Error(err.Error())
		return
	}

	file2, line2, tPre2, err := trace.InfoFromTID(posClose)
	if err != nil {
		log.Error(err.Error())
		return
	}

	arg1 := results.TraceElementResult{ // recv
		RoutineID: ch.GetRoutine(),
		ObjID:     id,
		TPre:      tPre1,
		ObjType:   "CR",
		File:      file1,
		Line:      line1,
	}

	arg2 := results.TraceElementResult{ // close
		RoutineID: data.CloseData[id].GetRoutine(),
		ObjID:     id,
		TPre:      tPre2,
		ObjType:   "CC",
		File:      file2,
		Line:      line2,
	}

	if actual {
		results.Result(results.WARNING, helper.ARecvOnClosed,
			"recv", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	} else {
		results.Result(results.WARNING, helper.PRecvOnClosed,
			"recv", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	}
}

// checkForClosedOnClosed checks for a close on a closed channel.
// Must be called, before the current close operation is added to closePos
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func checkForClosedOnClosed(ch *trace.TraceElementChannel) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := ch.GetID()

	if oldClose, ok := data.CloseData[id]; ok {
		if oldClose.GetTID() == "" || oldClose.GetTID() == "\n" || ch.GetTID() == "" || ch.GetTID() == "\n" {
			return
		}

		file1, line1, tPre1, err := trace.InfoFromTID(oldClose.GetTID())
		if err != nil {
			log.Error(err.Error())
			return
		}

		file2, line2, tPre2, err := trace.InfoFromTID(oldClose.GetTID())
		if err != nil {
			log.Error(err.Error())
			return
		}

		arg1 := results.TraceElementResult{
			RoutineID: ch.GetRoutine(),
			ObjID:     id,
			TPre:      tPre1,
			ObjType:   "CC",
			File:      file1,
			Line:      line1,
		}

		arg2 := results.TraceElementResult{
			RoutineID: oldClose.GetRoutine(),
			ObjID:     id,
			TPre:      tPre2,
			ObjType:   "CC",
			File:      file2,
			Line:      line2,
		}

		log.Error("Found Close on Close: ", ch.ToString())

		results.Result(results.CRITICAL, helper.ACloseOnClosed,
			"close", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	}
}
