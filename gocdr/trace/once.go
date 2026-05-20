// Copyright (c) 2024 Erik Kassubek
//
// File: /gocdr/trace/once.go
// Brief: Struct and functions for once operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-09-25
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"gocdr/utils/consts"
)

// ElementOnce is a trace element for a once
// Fields:
//   - id: id of the element, should never be changed
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - objId int: The id of the mutex
//   - suc bool: Whether the operation was successful
//   - file (string), line int: The position of the mutex operation in the code
type ElementOnce struct {
	id      int
	index   int
	routine int
	tPre    int
	tPost   int
	objId   int
	suc     bool
	file    string
	line    int
}

// AddTraceElementOnce adds a new mutex trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the mutex
//   - suc string: Whether the operation was successful (only for trylock else always true)
//   - pos string: The position of the mutex operation in the code
func (this *Trace) AddTraceElementOnce(routine int, tPre string,
	tPost string, id string, suc string, pos string) error {
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

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementOnce{
		index:   this.NumberElemInRoutine(routine),
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		objId:   idInt,
		suc:     sucBool,
		file:    file,
		line:    line,
	}

	this.AddElement(&elem)

	return nil
}

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementOnce) GetObjId() int {
	return this.objId
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementOnce) GetRoutine() int {
	return this.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (this *ElementOnce) GetTPre() int {
	return this.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (this *ElementOnce) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementOnce) GetTSort() int {
	if this.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return this.tPre
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementOnce) GetPos() string {
	return fmt.Sprintf("%s%s%d", this.file, consts.PosSep, this.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementOnce) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementOnce) GetFile() string {
	return this.file
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementOnce) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementOnce) GetTID() string {
	return "O@" + this.GetPos() + "@" + strconv.Itoa(this.tPre)
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementOnce) GetType(operation bool) OperationType {
	if !operation {
		return Once
	}

	if this.suc {
		return OnceSuc
	}
	return OnceFail
}

// GetSuc returns whether the once do was executed (successful)
//
// Returns:
//   - bool: true if function in Do was executed, false otherwise
func (this *ElementOnce) GetSuc() bool {
	return this.suc
}

// SetSuc sets whether the once do was executed successful
//
// Parameter:
//   - bool: true if function in Do was executed, false otherwise
func (this *ElementOnce) SetSuc(s bool) {
	this.suc = s
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementOnce) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same once
func (this *ElementOnce) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Once {
		return false
	}

	return this.objId == elem.GetObjId()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementOnce) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementOnce) SetT(time int) {
	this.tPre = time
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementOnce) SetTPre(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementOnce) SetTSort(tSort int) {
	this.SetTPre(tSort)
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementOnce) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementOnce) ToString() string {
	res := "O,"
	res += strconv.Itoa(this.tPre) + ","
	res += strconv.Itoa(this.tPost) + ","
	res += strconv.Itoa(this.objId) + ","
	if this.suc {
		res += "t"
	} else {
		res += "f"
	}
	res += "," + this.GetPos()
	return res
}

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementOnce) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementOnce) setID(ID int) {
	this.id = ID
}

// Copy the element
//
// Parameter:
//   - mapping map[string]Element: map containing all already copied elements.
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementOnce) Copy(_ map[string]Element, keep bool) Element {
	if !keep {
		return &ElementOnce{
			id:      this.id,
			index:   0,
			routine: this.routine,
			tPre:    0,
			tPost:   0,
			objId:   this.objId,
			suc:     false,
			file:    this.file,
			line:    this.line,
		}
	}

	return &ElementOnce{
		id:      this.id,
		index:   this.index,
		routine: this.routine,
		tPre:    this.tPre,
		tPost:   this.tPost,
		objId:   this.objId,
		suc:     this.suc,
		file:    this.file,
		line:    this.line,
	}
}

func (this *ElementOnce) IsValid() bool {
	return this != nil
}
