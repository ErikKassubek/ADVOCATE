// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementMutex.go
// Brief: Struct and functions for mutex operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"advocate/analysis/hb/clock"
)

// ElementMutex is a trace element for a mutex
//
// Fields:
//   - traceID: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the mutex
//   - rw bool: Whether the mutex is a read-noWarningrite mutex
//   - op ObjectType: The operation on the mutex
//   - suc bool: Whether the operation was successful (only for trylock else always true)
//   - file string: The file of the mutex operation in the code
//   - line int: The line of the mutex operation in the code
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementMutex struct {
	traceID                  int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	id                       int
	rw                       bool
	op                       ObjectType
	suc                      bool
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementMutex adds a new mutex element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the mutex
//   - rw string: Whether the mutex is a read-noWarningrite mutex
//   - opM string: The operation on the mutex
//   - suc string: Whether the operation was successful (only for trylock else always true)
//   - pos string: The position of the mutex operation in the code
func (t *Trace) AddTraceElementMutex(routine int, tPre string,
	tPost string, id string, rw string, opM string, suc string,
	pos string) error {
	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	rwBool := false
	if rw == "t" {
		rwBool = true
	}

	var opMInt ObjectType
	switch opM {
	case "L":
		opMInt = MutexLock
	case "R":
		opMInt = MutexRLock
	case "T":
		opMInt = MutexTryLock
	case "Y":
		opMInt = MutexTryRLock
	case "U":
		opMInt = MutexUnlock
	case "N":
		opMInt = MutexRUnlock
	default:
		return errors.New("opM is not a valid operation")
	}

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementMutex{
		index:                    t.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		id:                       idInt,
		rw:                       rwBool,
		op:                       opMInt,
		suc:                      sucBool,
		file:                     file,
		line:                     line,
		vc:                       nil,
		wVc:                      nil,
		numberConcurrent:         -1,
		numberConcurrentWeak:     -1,
		numberConcurrentSame:     -1,
		numberConcurrentWeakSame: -1,
	}

	t.AddElement(&elem)
	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (mu *ElementMutex) GetID() int {
	return mu.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (mu *ElementMutex) GetRoutine() int {
	return mu.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (mu *ElementMutex) GetTPre() int {
	return mu.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (mu *ElementMutex) GetTPost() int {
	return mu.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (mu *ElementMutex) GetTSort() int {
	if mu.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return mu.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (mu *ElementMutex) GetPos() string {
	return fmt.Sprintf("%s:%d", mu.file, mu.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (mu *ElementMutex) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", mu.routine, mu.file, mu.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (mu *ElementMutex) GetFile() string {
	return mu.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (mu *ElementMutex) GetLine() int {
	return mu.line
}

// GetTID returns the tID of the element.
// The tID is a string of form "M@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (mu *ElementMutex) GetTID() string {
	return "M@" + mu.GetPos() + "@" + strconv.Itoa(mu.tPre)
}

// IsLock returns if the element is a lock operation
//
// Returns:
//   - bool: If the element is a lock operation
func (mu *ElementMutex) IsLock() bool {
	return mu.op == MutexLock || mu.op == MutexRLock || mu.op == MutexTryLock || mu.op == MutexTryRLock
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (mu *ElementMutex) SetVc(vc *clock.VectorClock) {
	mu.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (mu *ElementMutex) SetWVc(vc *clock.VectorClock) {
	mu.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (mu *ElementMutex) GetVC() *clock.VectorClock {
	return mu.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (mu *ElementMutex) GetWVC() *clock.VectorClock {
	return mu.wVc
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (mu *ElementMutex) GetType(operation bool) ObjectType {
	if !operation {
		return Mutex
	}

	return mu.op
}

// IsSuc returns whether the locking was successful of the element
//
// Returns:
//   - For trylock wether it was successful, otherwise always true
func (mu *ElementMutex) IsSuc() bool {
	return mu.suc
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (mu *ElementMutex) IsEqual(elem Element) bool {
	return mu.routine == elem.GetRoutine() && mu.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same mutex
func (mu *ElementMutex) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Mutex {
		return false
	}

	return mu.id == elem.GetID()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (mu *ElementMutex) GetTraceIndex() (int, int) {
	return mu.routine, mu.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (mu *ElementMutex) SetT(time int) {
	mu.tPre = time
	mu.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (mu *ElementMutex) SetTPre(tPre int) {
	mu.tPre = tPre
	if mu.tPost != 0 && mu.tPost < tPre {
		mu.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (mu *ElementMutex) SetTSort(tSort int) {
	mu.SetTPre(tSort)
	mu.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (mu *ElementMutex) SetTWithoutNotExecuted(tSort int) {
	mu.SetTPre(tSort)
	if mu.tPost != 0 {
		mu.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (mu *ElementMutex) ToString() string {
	res := "M,"
	res += strconv.Itoa(mu.tPre) + "," + strconv.Itoa(mu.tPost) + ","
	res += strconv.Itoa(mu.id) + ","

	if mu.rw {
		res += "R,"
	} else {
		res += "-,"
	}

	res += string(string(mu.op)[1])

	if mu.suc {
		res += ",t"
	} else {
		res += ",f"
	}
	res += "," + mu.GetPos()
	return res
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (mu *ElementMutex) GetTraceID() int {
	return mu.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (mu *ElementMutex) setTraceID(ID int) {
	mu.traceID = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since mutex do not contain reference to other elements and no other
//     elements contain referents to mutex, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (mu *ElementMutex) Copy(_ map[string]Element) Element {
	return &ElementMutex{
		traceID:                  mu.traceID,
		index:                    mu.index,
		routine:                  mu.routine,
		tPre:                     mu.tPre,
		tPost:                    mu.tPost,
		id:                       mu.id,
		rw:                       mu.rw,
		op:                       mu.op,
		suc:                      mu.suc,
		file:                     mu.file,
		line:                     mu.line,
		vc:                       mu.vc.Copy(),
		wVc:                      mu.wVc.Copy(),
		numberConcurrent:         mu.numberConcurrent,
		numberConcurrentWeak:     mu.numberConcurrentWeak,
		numberConcurrentSame:     mu.numberConcurrentSame,
		numberConcurrentWeakSame: mu.numberConcurrentWeakSame,
	}
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Parameter:
//   - weak bool: get number of weak concurrent
//   - sameElem bool: only operation on the same variable
//
// Returns:
//   - number of concurrent element, or -1
func (mu *ElementMutex) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return mu.numberConcurrentWeakSame
		}
		return mu.numberConcurrentWeak
	}
	if sameElem {
		return mu.numberConcurrentSame
	}
	return mu.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (mu *ElementMutex) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			mu.numberConcurrentWeakSame = c
		} else {
			mu.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			mu.numberConcurrentSame = c
		} else {
			mu.numberConcurrent = c
		}
	}
}
