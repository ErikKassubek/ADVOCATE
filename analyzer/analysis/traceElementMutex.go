// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementMutex.go
// Brief: Struct and functions for mutex operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package analysis

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"analyzer/clock"
	"analyzer/utils"
)

// enum for opM
type OpMutex int

const (
	LockOp OpMutex = iota
	RLockOp
	TryLockOp
	TryRLockOp
	UnlockOp
	RUnlockOp
)

/*
 * TraceElementMutex is a trace element for a mutex
 * MARK: Struct
 * Fields:
 *   index (int): Index in the routine
 *   routine (int): The routine id
 *   tpre (int): The timestamp at the start of the event
 *   tpost (int): The timestamp at the end of the event
 *   id (int): The id of the mutex
 *   rw (bool): Whether the mutex is a read-noWarningrite mutex
 *   opM (opMutex): The operation on the mutex
 *   suc (bool): Whether the operation was successful (only for trylock else always true)
 *   file (string), line(int): The position of the mutex operation in the code
 */
type TraceElementMutex struct {
	index   int
	routine int
	tPre    int
	tPost   int
	id      int
	rw      bool
	opM     OpMutex
	suc     bool
	file    string
	line    int
	vc      clock.VectorClock
	rel1    []TraceElement
	rel2    []TraceElement
}

/*
 * Create a new mutex trace element
 * MARK: New
 * Args:
 *   routine (int): The routine id
 *   tPre (string): The timestamp at the start of the event
 *   tPost (string): The timestamp at the end of the event
 *   id (string): The id of the mutex
 *   rw (string): Whether the mutex is a read-noWarningrite mutex
 *   opM (string): The operation on the mutex
 *   suc (string): Whether the operation was successful (only for trylock else always true)
 *   pos (string): The position of the mutex operation in the code
 */
func AddTraceElementMutex(routine int, tPre string,
	tPost string, id string, rw string, opM string, suc string,
	pos string) error {
	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	rwBool := false
	if rw == "t" {
		rwBool = true
	}

	var opMInt OpMutex
	switch opM {
	case "L":
		opMInt = LockOp
	case "R":
		opMInt = RLockOp
	case "T":
		opMInt = TryLockOp
	case "Y":
		opMInt = TryRLockOp
	case "U":
		opMInt = UnlockOp
	case "N":
		opMInt = RUnlockOp
	default:
		return errors.New("opM is not a valid operation")
	}

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementMutex{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		rw:      rwBool,
		opM:     opMInt,
		suc:     sucBool,
		file:    file,
		line:    line,
		rel1:    make([]TraceElement, 2),
		rel2:    make([]TraceElement, 0),
	}

	AddElementToTrace(&elem)
	return nil
}

// MARK: Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (mu *TraceElementMutex) GetID() int {
	return mu.id
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (mu *TraceElementMutex) GetRoutine() int {
	return mu.routine
}

/*
 * Get the tpre of the element.
 * Returns:
 *   int: The tpre of the element
 */
func (mu *TraceElementMutex) GetTPre() int {
	return mu.tPre
}

/*
 * Get the tpost of the element.
 * Returns:
 *   int: The tpost of the element
 */
func (mu *TraceElementMutex) GetTPost() int {
	return mu.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (mu *TraceElementMutex) GetTSort() int {
	if mu.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return mu.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The position of the element
 */
func (mu *TraceElementMutex) GetPos() string {
	return fmt.Sprintf("%s:%d", mu.file, mu.line)
}

func (mu *TraceElementMutex) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", mu.routine, mu.file, mu.line)
}

func (mu *TraceElementMutex) GetFile() string {
	return mu.file
}

func (mu *TraceElementMutex) GetLine() int {
	return mu.line
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (mu *TraceElementMutex) GetTID() string {
	return mu.GetPos() + "@" + strconv.Itoa(mu.tPre)
}

/*
 * Get the operation of the element
 * Returns:
 *   OpMutex: The operation of the element
 */
func (mu *TraceElementMutex) GetOperation() OpMutex {
	return mu.opM
}

func (mu TraceElementMutex) GetSuccess() bool {
	return mu.suc
}

/*
 * Get if the element is a lock operation
 * Returns:
 *   bool: If the element is a lock operation
 */
func (mu *TraceElementMutex) IsLock() bool {
	return mu.opM == LockOp || mu.opM == RLockOp || mu.opM == TryLockOp || mu.opM == TryRLockOp
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (mu *TraceElementMutex) GetVC() clock.VectorClock {
	return mu.vc
}

/*
 * Get the string representation of the object type
 */
func (mu *TraceElementMutex) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeMutex
	}

	switch mu.opM {
	case LockOp:
		return ObjectTypeMutex + "L"
	case RLockOp:
		return ObjectTypeMutex + "R"
	case TryLockOp:
		return ObjectTypeMutex + "T"
	case TryRLockOp:
		return ObjectTypeMutex + "Y"
	case UnlockOp:
		return ObjectTypeMutex + "U"
	case RUnlockOp:
		return ObjectTypeMutex + "N"
	}
	return ObjectTypeMutex
}

func (mu *TraceElementMutex) IsSuc() bool {
	return mu.suc
}

func (mu *TraceElementMutex) IsEqual(elem TraceElement) bool {
	return mu.routine == elem.GetRoutine() && mu.ToString() == elem.ToString()
}

func (mu *TraceElementMutex) GetTraceIndex() (int, int) {
	return mu.routine, mu.index
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (mu *TraceElementMutex) SetT(time int) {
	mu.tPre = time
	mu.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (mu *TraceElementMutex) SetTPre(tPre int) {
	mu.tPre = tPre
	if mu.tPost != 0 && mu.tPost < tPre {
		mu.tPost = tPre
	}
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (mu *TraceElementMutex) SetTSort(tSort int) {
	mu.SetTPre(tSort)
	mu.tPost = tSort
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (mu *TraceElementMutex) SetTWithoutNotExecuted(tSort int) {
	mu.SetTPre(tSort)
	if mu.tPost != 0 {
		mu.tPost = tSort
	}
}

/*
 * Get the simple string representation of the element
 * MARK: ToString
 * Returns:
 *   string: The simple string representation of the element
 */
func (mu *TraceElementMutex) ToString() string {
	res := "M,"
	res += strconv.Itoa(mu.tPre) + "," + strconv.Itoa(mu.tPost) + ","
	res += strconv.Itoa(mu.id) + ","

	if mu.rw {
		res += "R,"
	} else {
		res += "-,"
	}

	switch mu.opM {
	case LockOp:
		res += "L"
	case RLockOp:
		res += "R"
	case TryLockOp:
		res += "T"
	case TryRLockOp:
		res += "Y"
	case UnlockOp:
		res += "U"
	case RUnlockOp:
		res += "N"
	}

	if mu.suc {
		res += ",t"
	} else {
		res += ",f"
	}
	res += "," + mu.GetPos()
	return res
}

/*
* Update the vector clock of the trace and element
* MARK: VectorClock
 */
func (mu *TraceElementMutex) updateVectorClock() {
	mu.vc = currentVCHb[mu.routine].Copy()

	switch mu.opM {
	case LockOp:
		Lock(mu, currentVCHb, currentVCWmhb)
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockLock(mu)
		}
		if analysisCases["cyclicDeadlock"] {
			CyclicDeadlockMutexLock(mu, false, currentVCWmhb[mu.routine])
		}
	case RLockOp:
		RLock(mu, currentVCHb, currentVCWmhb)
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockLock(mu)
		}
		if analysisCases["cyclicDeadlock"] {
			CyclicDeadlockMutexLock(mu, true, currentVCWmhb[mu.routine])
		}
	case TryLockOp:
		if mu.suc {
			if analysisCases["unlockBeforeLock"] {
				checkForUnlockBeforeLockLock(mu)
			}
			Lock(mu, currentVCHb, currentVCWmhb)
			if analysisCases["cyclicDeadlock"] {
				CyclicDeadlockMutexLock(mu, false, currentVCWmhb[mu.routine])
			}
		}
	case TryRLockOp:
		if mu.suc {
			RLock(mu, currentVCHb, currentVCWmhb)
			if analysisCases["unlockBeforeLock"] {
				checkForUnlockBeforeLockLock(mu)
			}
			if analysisCases["cyclicDeadlock"] {
				CyclicDeadlockMutexLock(mu, true, currentVCWmhb[mu.routine])
			}
		}
	case UnlockOp:
		Unlock(mu, currentVCHb)
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockUnlock(mu)
		}
		if analysisCases["cyclicDeadlock"] {
			CyclicDeadlockMutexUnLock(mu)
		}
	case RUnlockOp:
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockUnlock(mu)
		}
		RUnlock(mu, currentVCHb)
		if analysisCases["cyclicDeadlock"] {
			CyclicDeadlockMutexUnLock(mu)
		}
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		utils.LogError(err)
	}
}

func (mu *TraceElementMutex) updateVectorClockAlt() {
	mu.vc = currentVCHb[mu.routine].Copy()

	currentVCHb[mu.routine] = currentVCHb[mu.routine].Inc(mu.routine)
}

/*
 * Copy the element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (mu *TraceElementMutex) Copy() TraceElement {
	return &TraceElementMutex{
		index:   mu.index,
		routine: mu.routine,
		tPre:    mu.tPre,
		tPost:   mu.tPost,
		id:      mu.id,
		rw:      mu.rw,
		opM:     mu.opM,
		suc:     mu.suc,
		file:    mu.file,
		line:    mu.line,
		vc:      mu.vc.Copy(),
	}
}

// MARK: GoPie
func (mu *TraceElementMutex) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}
	mu.rel1[pos] = elem
}

func (mu *TraceElementMutex) AddRel2(elem TraceElement) {
	mu.rel2 = append(mu.rel2, elem)
}

func (mu *TraceElementMutex) GetRel1() []TraceElement {
	return mu.rel1
}

func (mu *TraceElementMutex) GetRel2() []TraceElement {
	return mu.rel2
}
