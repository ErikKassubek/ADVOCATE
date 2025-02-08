// Copyrigth (c) 2024 Erik Kassubek
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
	timemeasurement "analyzer/timeMeasurement"
	"log"
)

/*
 * Check if a send or receive on a closed channel is possible
 * It it is possible, print a warning or error
 * Args:
 *   ch (*TraceElementChannel): The trace element
 */
func checkForCommunicationOnClosedChannel(ch *TraceElementChannel) {
	// check if there is an earlier send, that could happen concurrently to close
	// println("Check for possible send on closed channel ", analysisCases["sendOnClosed"], hasSend[id])
	if analysisCases["sendOnClosed"] && hasSend[ch.id] {
		timemeasurement.Start("panic")
		defer timemeasurement.End("panic")

		for routine, mrs := range mostRecentSend {
			happensBefore := clock.GetHappensBefore(mrs[ch.id].Vc, closeData[ch.id].vc)

			if mrs[ch.id].Elem != nil && mrs[ch.id].Elem.GetTID() != "" && happensBefore != clock.Before {

				file1, line1, tPre1, err := infoFromTID(mrs[ch.id].Elem.GetTID()) // send
				if err != nil {
					log.Print(err.Error())
					return
				}

				file2, line2, tPre2, err := infoFromTID(ch.GetTID()) // close
				if err != nil {
					log.Print(err.Error())
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
		timemeasurement.Start("other")
		defer timemeasurement.End("other")

		for routine, mrr := range mostRecentReceive {
			happensBefore := clock.GetHappensBefore(closeData[ch.id].vc, mrr[ch.id].Vc)
			if mrr[ch.id].Elem != nil && mrr[ch.id].Elem.GetTID() != "" && (happensBefore == clock.Concurrent || happensBefore == clock.Before) {

				file1, line1, tPre1, err := infoFromTID(mrr[ch.id].Elem.GetTID()) // recv
				if err != nil {
					log.Print(err.Error())
					return
				}

				file2, line2, tPre2, err := infoFromTID(ch.GetTID()) // close
				if err != nil {
					log.Print(err.Error())
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
 *  routineID (int): id of the routine where the send happened
 *  id (int): id of the channel
 *  posSend (string): code location of the send
 *  actual (bool): set actual to true it the panic occurred, set to false if it is in an not triggered select case
 */
func foundSendOnClosedChannel(routineID int, id int, posSend string, actual bool) {
	timemeasurement.Start("panic")
	defer timemeasurement.End("panic")

	if _, ok := closeData[id]; !ok {
		return
	}

	posClose := closeData[id].GetTID()
	if posClose == "" || posSend == "" || posClose == "\n" || posSend == "\n" {
		return
	}

	file1, line1, tPre1, err := infoFromTID(posSend)
	if err != nil {
		log.Print(err.Error())
		return
	}

	file2, line2, tPre2, err := infoFromTID(posClose)
	if err != nil {
		log.Print(err.Error())
		return
	}

	arg1 := results.TraceElementResult{ // send
		RoutineID: routineID,
		ObjID:     id,
		TPre:      tPre1,
		ObjType:   "CS",
		File:      file1,
		Line:      line1,
	}

	arg2 := results.TraceElementResult{ // close
		RoutineID: closeData[id].routine,
		ObjID:     id,
		TPre:      tPre2,
		ObjType:   "CC",
		File:      file2,
		Line:      line2,
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
 *  ch (*TraceElementChannel): The trace element
 */
func foundReceiveOnClosedChannel(ch *TraceElementChannel, actual bool) {
	timemeasurement.Start("panic")
	defer timemeasurement.End("panic")

	if _, ok := closeData[ch.id]; !ok {
		return
	}

	posClose := closeData[ch.id].GetTID()
	if posClose == "" || ch.GetTID() == "" || posClose == "\n" || ch.GetTID() == "\n" {
		return
	}

	file1, line1, tPre1, err := infoFromTID(ch.GetTID())
	if err != nil {
		log.Print(err.Error())
		return
	}

	file2, line2, tPre2, err := infoFromTID(posClose)
	if err != nil {
		log.Print(err.Error())
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
 *  ch (*TraceElementChannel): The trace element
 */
func checkForClosedOnClosed(ch *TraceElementChannel) {
	timemeasurement.Start("panic")
	defer timemeasurement.End("panic")

	if oldClose, ok := closeData[ch.id]; ok {
		if oldClose.GetTID() == "" || oldClose.GetTID() == "\n" || ch.GetTID() == "" || ch.GetTID() == "\n" {
			return
		}

		file1, line1, tPre1, err := infoFromTID(oldClose.GetTID())
		if err != nil {
			log.Print(err.Error())
			return
		}

		file2, line2, tPre2, err := infoFromTID(oldClose.GetTID())
		if err != nil {
			log.Print(err.Error())
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

		results.Result(results.CRITICAL, results.ACloseOnClosed,
			"close", []results.ResultElem{arg1}, "close", []results.ResultElem{arg2})
	}
}
