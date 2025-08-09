//
// File: traceElementCond.go
// Brief: Struct and functions for operations of conditional variables in the trace
//
// Created: 2023-12-25
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"goCR/analysis/hb/clock"
	"math"
	"strconv"
)

// OpCond provides an enum for the operation of a conditional variable
type OpCond int

// Values for the OpCount enum
const (
	WaitCondOp OpCond = iota
	SignalOp
	BroadcastOp
)

// ElementCond is a trace element for a condition variable
// Fields:
//   - traceID: id of the element, should never be changed
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the condition variable
//   - opC opCond: The operation on the condition variable
//   - file string, The file of the condition variable operation in the code
//   - line int, The line of the condition variable operation in the code
//   - children []TraceElement: children in partial order graph
//   - parent []TraceElement: parents in partial order graph
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementCond struct {
	traceID                  int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	id                       int
	opC                      OpCond
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementCond adds a new condition variable element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the condition variable
//   - opC string: The operation on the condition variable
//   - pos string: The position of the condition variable operation in the code
func (t *Trace) AddTraceElementCond(routine int, tPre string, tPost string, id string, opN string, pos string) error {
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
	var op OpCond
	switch opN {
	case "W":
		op = WaitCondOp
	case "S":
		op = SignalOp
	case "B":
		op = BroadcastOp
	default:
		return errors.New("op is not a valid operation")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementCond{
		index:                    t.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		id:                       idInt,
		opC:                      op,
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
func (co *ElementCond) GetID() int {
	return co.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine id
func (co *ElementCond) GetRoutine() int {
	return co.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (co *ElementCond) GetTPre() int {
	return co.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (co *ElementCond) GetTPost() int {
	return co.tPost
}

// GetTSort returns the timer, that is used for sorting the trace
//
// Returns:
//   - int: The timer of the element
func (co *ElementCond) GetTSort() int {
	t := co.tPre
	if co.opC == WaitCondOp {
		t = co.tPost
	}
	if t == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return t
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (co *ElementCond) GetPos() string {
	return fmt.Sprintf("%s:%d", co.file, co.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (co *ElementCond) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", co.routine, co.file, co.line)
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (co *ElementCond) GetFile() string {
	return co.file
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (co *ElementCond) GetLine() int {
	return co.line
}

// GetTID returns the tID of the element.
// The tID is a string of form D@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (co *ElementCond) GetTID() string {
	return "D@" + co.GetPos() + "@" + strconv.Itoa(co.tPre)
}

// GetOpC returns the operation of the element
//
// Returns:
//   - OpCond: The operation of the element
func (co *ElementCond) GetOpC() OpCond {
	return co.opC
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (co *ElementCond) SetVc(vc *clock.VectorClock) {
	co.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (co *ElementCond) SetWVc(vc *clock.VectorClock) {
	co.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (co *ElementCond) GetVC() *clock.VectorClock {
	return co.vc
}

// GetWVC returns the vector clock of the element for the weak must happens before relation
//
// Returns:
//   - VectorClock: The vector clock of the element
func (co *ElementCond) GetWVC() *clock.VectorClock {
	return co.wVc
}

// GetObjType returns the string representation of the object type
func (co *ElementCond) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeCond
	}

	switch co.opC {
	case WaitCondOp:
		return ObjectTypeCond + "W"
	case BroadcastOp:
		return ObjectTypeCond + "B"
	case SignalOp:
		return ObjectTypeCond + "S"
	}
	return ObjectTypeCond
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (co *ElementCond) IsEqual(elem Element) bool {
	return co.routine == elem.GetRoutine() && co.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (co *ElementCond) GetTraceIndex() (int, int) {
	return co.routine, co.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (co *ElementCond) SetT(time int) {
	co.tPre = time
	co.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (co *ElementCond) SetTPre(tPre int) {
	co.tPre = tPre
	if co.tPost != 0 && co.tPost < tPre {
		co.tPost = tPre
	}
}

// SetTSort sets the timer that is used for sorting the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (co *ElementCond) SetTSort(tSort int) {
	co.SetTPre(tSort)
	if co.opC == WaitCondOp {
		co.tPost = tSort
	}
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (co *ElementCond) SetTWithoutNotExecuted(tSort int) {
	co.SetTPre(tSort)
	if co.opC == WaitCondOp {
		if co.tPost != 0 {
			co.tPost = tSort
		}
		return
	}
	if co.tPre != 0 {
		co.tPre = tSort
	}
}

// ToString returns the string representation of the element
//
// Returns:
//   - string: The string representation of the element
func (co *ElementCond) ToString() string {
	res := "D,"
	res += strconv.Itoa(co.tPre) + "," + strconv.Itoa(co.tPost) + ","
	res += strconv.Itoa(co.id) + ","
	switch co.opC {
	case WaitCondOp:
		res += "W"
	case SignalOp:
		res += "S"
	case BroadcastOp:
		res += "B"
	}
	res += "," + co.GetPos()
	return res
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (co *ElementCond) GetTraceID() int {
	return co.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (co *ElementCond) setTraceID(ID int) {
	co.traceID = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since conds do not contain reference to other elements and no other
//     elements contain referents to conds, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (co *ElementCond) Copy(_ map[string]Element) Element {
	return &ElementCond{
		traceID:                  co.traceID,
		index:                    co.index,
		routine:                  co.routine,
		tPre:                     co.tPre,
		tPost:                    co.tPost,
		id:                       co.id,
		opC:                      co.opC,
		file:                     co.file,
		line:                     co.line,
		vc:                       co.vc.Copy(),
		wVc:                      co.wVc.Copy(),
		numberConcurrent:         co.numberConcurrent,
		numberConcurrentWeak:     co.numberConcurrentWeak,
		numberConcurrentSame:     co.numberConcurrentSame,
		numberConcurrentWeakSame: co.numberConcurrentWeakSame,
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
func (co *ElementCond) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return co.numberConcurrentWeakSame
		}
		return co.numberConcurrentWeak
	}
	if sameElem {
		return co.numberConcurrentSame
	}
	return co.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (co *ElementCond) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			co.numberConcurrentWeakSame = c
		} else {
			co.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			co.numberConcurrentSame = c
		} else {
			co.numberConcurrent = c
		}
	}
}
