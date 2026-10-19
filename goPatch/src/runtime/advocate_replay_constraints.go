// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay_constrains.go
// Brief: Functionality to handle the constraints
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

type ConstraintElem struct {
	time  string
	key   string
	index int
}

func NewConstraintElem(routine int, file string, line int, index int) *ConstraintElem {
	return &ConstraintElem{
		key:   intToString(routine) + "$" + file + "#" + intToString(line),
		index: index,
	}
}

func AddConstraintElemFromString(info string) {
	fields := split(info, ':')
	if len(fields) != 3 {
		return
	}

	newConst := ConstraintElem{
		key:   fields[0] + "$" + fields[1],
		index: stringToInt(fields[2]),
	}

	constratins.Add(newConst)

}

type Constraint struct {
	order       []ConstraintElem
	elems       map[string]map[int]bool // key -> index
	elemIndex   map[string]int          // key -> currentIndex
	replayIndex int

	mu mutex
}

func NewConstraint() {
	constratins = Constraint{
		order:       make([]ConstraintElem, 0),
		elems:       make(map[string]map[int]bool),
		elemIndex:   make(map[string]int),
		replayIndex: 0,
	}
}

func (this *Constraint) Add(elem ConstraintElem) {
	this.order = append(this.order, elem)
	println(elem.key, " -> ", elem.index)
	if _, ok := this.elems[elem.key]; !ok {
		this.elems[elem.key] = make(map[int]bool)
	}
	this.elems[elem.key][elem.index] = true
	this.elemIndex[elem.key] = 1
}

// func (this *Constraint) isNext() *ConstraintElem {
// 	lock(mu)
// 	elem := this.elems[ReplayIndex]
// 	unlock(mu)
// 	return elem
// }

func (this *Constraint) isNext(key string) bool {
	lock(&this.mu)
	next := this.order[this.replayIndex]
	unlock(&this.mu)
	return next.key == key
}

func (this *Constraint) getNext() *ConstraintElem {
	lock(&this.mu)
	next := &this.order[this.replayIndex]
	unlock(&this.mu)
	return next
}

func (this *Constraint) isDone() bool {
	return len(this.order) == this.replayIndex
}

/*
// Info determines if the operation of the key should be blocked or released
// Returns:
//	   bool: true if it should block, false if it can be released
*/
func (this *Constraint) isInConstraint(key string) bool {
	print("IS IN CONSTRAINT: ", key, " -> ")
	lock(&this.mu)
	info, ok := this.elems[key]
	unlock(&this.mu)

	// is never in constratins
	if !ok {
		println("FALSE 1")
		return false
	}

	// is in constraints with this index
	lock(&this.mu)
	index := this.elemIndex[key]
	print(index, " -> ")
	if info[index] {
		unlock(&this.mu)
		println("TRUE")
		return true
	}
	unlock(&this.mu)

	// is in constraints but not with this index
	println("FALSE 2")
	return false
}

func (this *Constraint) update(key string) {
	lock(&this.mu)
	this.elemIndex[key]++
	this.replayIndex++
	unlock(&this.mu)

	this.releaseBlocked()
}

func (this *Constraint) releaseBlocked() {
	next := this.getNext()

	if c := waiting.contains(next); c != nil {
		c <- 0 // TODO: select index
		waiting.remove(next)
		this.releaseBlocked()
	}
}

func (this *Constraint) releaseAllBlocked() {
	lock(&this.mu)

	for _, ch := range waiting.waiting {
		// TODO: send select id or chosen case
		ch <- -1
	}
	waiting.waiting = make(map[string]chan int)

	unlock(&this.mu)
}

func ReleaseAllBlocked() {
	constratins.releaseAllBlocked()
}
