package trace

import (
	"analyzer/analysis"
	"analyzer/clock"
	"analyzer/logging"
	"errors"
	"strconv"
)

// enum for operation
type opAtomic int

const (
	LoadOp opAtomic = iota
	StoreOp
	AddOp
	SwapOp
	CompSwapOp
)

/*
 * Struct to save an atomic event in the trace
 * MARK: Struct
 * Fields:
 *   routine (int): The routine id
 *   tpost (int): The timestamp of the event
 *   id (int): The id of the atomic variable
 *   operation (int, enum): The operation on the atomic variable
 */
type TraceElementAtomic struct {
	routine int
	tPost   int
	id      int
	opA     opAtomic
	vc      clock.VectorClock
}

/*
 * Create a new atomic trace element
 * MARK: New
 * Args:
 *   routine (int): The routine id
 *   tpost (string): The timestamp of the event
 *   id (string): The id of the atomic variable
 *   operation (string): The operation on the atomic variable
 */
func AddTraceElementAtomic(routine int, tpost string,
	id string, operation string) error {
	tPostInt, err := strconv.Atoi(tpost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	var opAInt opAtomic
	switch operation {
	case "L":
		opAInt = LoadOp
	case "S":
		opAInt = StoreOp
	case "A":
		opAInt = AddOp
	case "W":
		opAInt = SwapOp
	case "C":
		opAInt = CompSwapOp
	default:
		return errors.New("operation is not a valid operation")
	}

	elem := TraceElementAtomic{
		routine: routine,
		tPost:   tPostInt,
		id:      idInt,
		opA:     opAInt,
	}

	return AddElementToTrace(&elem)
}

// MARK: Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (at *TraceElementAtomic) GetID() int {
	return at.id
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (at *TraceElementAtomic) GetRoutine() int {
	return at.routine
}

/*
 * Get the tpost of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpost of the element
 */
func (at *TraceElementAtomic) GetTPre() int {
	return at.tPost
}

/*
 * Get the tpost of the element. For atomic elements, tpre and tpost are the same
 * Returns:
 *   int: The tpost of the element
 */
func (at *TraceElementAtomic) getTpost() int {
	return at.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (at *TraceElementAtomic) GetTSort() int {
	return at.tPost
}

/*
 * Get the position of the operation. For atomic elements, the position is always empty
 * Returns:
 *   string: The file of the element
 */
func (at *TraceElementAtomic) GetPos() string {
	return ""
}

/*
 * Get the tID of the element. For atomic elements, the tID is always empty
 * Returns:
 *   string: The tID of the element
 */
func (at *TraceElementAtomic) GetTID() string {
	return "A@" + strconv.Itoa(at.tPost)
}

/*
 * Dummy function to implement the interface
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (at *TraceElementAtomic) GetVC() clock.VectorClock {
	return at.vc
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (at *TraceElementAtomic) SetT(time int) {
	at.tPost = time
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpost of the element
 */
func (at *TraceElementAtomic) SetTPre(tPre int) {
	at.tPost = tPre
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (at *TraceElementAtomic) SetTSort(tSort int) {
	at.SetTPre(tSort)
	at.tPost = tSort
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (at *TraceElementAtomic) SetTWithoutNotExecuted(tSort int) {
	at.SetTPre(tSort)
	if at.tPost != 0 {
		at.tPost = tSort
	}
}

// MARK: ToString

/*
 * Get the simple string representation of the element.
 * Returns:
 *   string: The simple string representation of the element
 */
func (at *TraceElementAtomic) ToString() string {
	res := "A," + strconv.Itoa(at.tPost) + "," +
		strconv.Itoa(at.id) + ","

	switch at.opA {
	case LoadOp:
		res += "L"
	case StoreOp:
		res += "S"
	case AddOp:
		res += "A"
	case SwapOp:
		res += "W"
	case CompSwapOp:
		res += "C"
	default:
		res += "U"
	}

	return res
}

// MARK: Vector Clock

/*
 * Update and calculate the vector clock of the element
 */
func (at *TraceElementAtomic) updateVectorClock() {
	at.vc = currentVCHb[at.routine].Copy()

	switch at.opA {
	case LoadOp:
		analysis.Read(at.routine, at.id, currentVCHb, true)
	case StoreOp, AddOp:
		analysis.Write(at.routine, at.id, currentVCHb)
	case SwapOp, CompSwapOp:
		analysis.Swap(at.routine, at.id, currentVCHb, true)
	default:
		err := "Unknown operation: " + at.ToString()
		logging.Debug(err, logging.ERROR)
	}
}

/*
 * Update and calculate the vector clock of the element
 */
func (at *TraceElementAtomic) updateVectorClockAlt() {
	at.vc = currentVCHb[at.routine].Copy()

	switch at.opA {
	case LoadOp:
		analysis.Read(at.routine, at.id, currentVCHb, false)
	case StoreOp, AddOp:
		analysis.Write(at.routine, at.id, currentVCHb)
	case SwapOp, CompSwapOp:
		analysis.Swap(at.routine, at.id, currentVCHb, false)
	default:
		err := "Unknown operation: " + at.ToString()
		logging.Debug(err, logging.ERROR)
	}
}

// MARK: Copy

/*
 * Copy the atomic element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (at *TraceElementAtomic) Copy() TraceElement {
	return &TraceElementAtomic{
		routine: at.routine,
		tPost:   at.tPost,
		id:      at.id,
		opA:     at.opA,
		vc:      at.vc.Copy(),
	}
}
