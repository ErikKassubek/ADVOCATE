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
	"analyzer/clock"
	"analyzer/results"
	"analyzer/timer"
	"analyzer/utils"
)

/*
 * Check if a send or receive on a closed channel is possible
 * It it is possible, print a warning or error
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 */
func checkForCommunicationOnClosedChannel(ch *TraceElementChannel) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	// check if there is an earlier send, that could happen concurrently to close
	if analysisCases["sendOnClosed"] && hasSend[ch.id] {
		for routine, mrs := range mostRecentSend {
			happensBefore := clock.GetHappensBefore(mrs[ch.id].Vc, closeData[ch.id].vc)

			if mrs[ch.id].Elem != nil && mrs[ch.id].Elem.GetTID() != "" && happensBefore != clock.Before {

				file1, line1, tPre1, err := infoFromTID(mrs[ch.id].Elem.GetTID()) // send
				if err != nil {
					utils.LogError(err.Error())
					return
				}

				file2, line2, tPre2, err := infoFromTID(ch.GetTID()) // close
				if err != nil {
					utils.LogError(err.Error())
					return
				}

				arg1 := results.TraceElementResult{ // send
					RoutineID: routine,
					ObjID:     ch.id,
					TPre:      tPre1,
					ObjType:   "CS",
					File:      file1,
					Line:      line1,
				}

				arg2 := results.TraceElementResult{ // close
					RoutineID: closeData[ch.id].routine,
					ObjID:     ch.id,
					TPre:      tPre2,
					ObjType:   "CC",
					File:      file2,
					Line:      line2,
				}

				results.Result(results.CRITICAL, results.PSendOnClosed,
					"send", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
			}
		}
	}
	// check if there is an earlier receive, that could happen concurrently to close
	if analysisCases["receiveOnClosed"] && hasReceived[ch.id] {
		for routine, mrr := range mostRecentReceive {
			happensBefore := clock.GetHappensBefore(closeData[ch.id].vc, mrr[ch.id].Vc)
			if mrr[ch.id].Elem != nil && mrr[ch.id].Elem.GetTID() != "" && (happensBefore == clock.Concurrent || happensBefore == clock.Before) {

				file1, line1, tPre1, err := infoFromTID(mrr[ch.id].Elem.GetTID()) // recv
				if err != nil {
					utils.LogError(err.Error())
					return
				}

				file2, line2, tPre2, err := infoFromTID(ch.GetTID()) // close
				if err != nil {
					utils.LogError(err.Error())
					return
				}

				arg1 := results.TraceElementResult{ // recv
					RoutineID: routine,
					ObjID:     ch.id,
					TPre:      tPre1,
					ObjType:   "CR",
					File:      file1,
					Line:      line1,
				}

				arg2 := results.TraceElementResult{ // close
					RoutineID: closeData[ch.id].routine,
					ObjID:     ch.id,
					TPre:      tPre2,
					ObjType:   "CC",
					File:      file2,
					Line:      line2,
				}

				results.Result(results.WARNING, results.PRecvOnClosed,
					"recv", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
			}
		}

	}

}

/*
 * Sound actual send on closed
 * Args:
 * 	elem (TraceElement): the send/select elem
 * 	id (int): id of the channel
 * 	actual (bool): set actual to true it the panic occurred, set to false if it is in an not triggered select case
 */
func foundSendOnClosedChannel(elem TraceElement, actual bool) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := elem.GetID()

	if _, ok := closeData[id]; !ok {
		return
	}

	closeElem := closeData[id]
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
		RoutineID: closeData[id].routine,
		ObjID:     id,
		TPre:      closeElem.tPre,
		ObjType:   "CC",
		File:      closeElem.GetFile(),
		Line:      closeElem.GetLine(),
	}

	if actual {
		results.Result(results.CRITICAL, results.ASendOnClosed,
			"send", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	} else {
		results.Result(results.CRITICAL, results.PSendOnClosed,
			"send", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	}

}

/*
 * Log the detection of an actual receive on a closed channel
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 */
func foundReceiveOnClosedChannel(ch *TraceElementChannel, actual bool) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	if _, ok := closeData[ch.id]; !ok {
		return
	}

	posClose := closeData[ch.id].GetTID()
	if posClose == "" || ch.GetTID() == "" || posClose == "\n" || ch.GetTID() == "\n" {
		return
	}

	file1, line1, tPre1, err := infoFromTID(ch.GetTID())
	if err != nil {
		utils.LogError(err.Error())
		return
	}

	file2, line2, tPre2, err := infoFromTID(posClose)
	if err != nil {
		utils.LogError(err.Error())
		return
	}

	arg1 := results.TraceElementResult{ // recv
		RoutineID: ch.routine,
		ObjID:     ch.id,
		TPre:      tPre1,
		ObjType:   "CR",
		File:      file1,
		Line:      line1,
	}

	arg2 := results.TraceElementResult{ // close
		RoutineID: closeData[ch.id].routine,
		ObjID:     ch.id,
		TPre:      tPre2,
		ObjType:   "CC",
		File:      file2,
		Line:      line2,
	}

	if actual {
		results.Result(results.WARNING, results.ARecvOnClosed,
			"recv", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	} else {
		results.Result(results.WARNING, results.PRecvOnClosed,
			"recv", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	}
}

/*
 * Check for a close on a closed channel.
 * Must be called, before the current close operation is added to closePos
 * Args:
 * 	ch (*TraceElementChannel): The trace element
 */
func checkForClosedOnClosed(ch *TraceElementChannel) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	if oldClose, ok := closeData[ch.id]; ok {
		if oldClose.GetTID() == "" || oldClose.GetTID() == "\n" || ch.GetTID() == "" || ch.GetTID() == "\n" {
			return
		}

		file1, line1, tPre1, err := infoFromTID(oldClose.GetTID())
		if err != nil {
			utils.LogError(err.Error())
			return
		}

		file2, line2, tPre2, err := infoFromTID(oldClose.GetTID())
		if err != nil {
			utils.LogError(err.Error())
			return
		}

		arg1 := results.TraceElementResult{
			RoutineID: ch.routine,
			ObjID:     ch.id,
			TPre:      tPre1,
			ObjType:   "CC",
			File:      file1,
			Line:      line1,
		}

		arg2 := results.TraceElementResult{
			RoutineID: oldClose.routine,
			ObjID:     ch.id,
			TPre:      tPre2,
			ObjType:   "CC",
			File:      file2,
			Line:      line2,
		}

		utils.LogError("Found Close on Close: ", ch.ToString())

		results.Result(results.CRITICAL, results.ACloseOnClosed,
			"close", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	}
}
