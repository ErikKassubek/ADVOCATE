package runtime

/*
 * AdvocateWaitGroupAdd adds a waitgroup add or done to the trace
 * MARK: Add
 * Args:
 * 	id: id of the waitgroup
 *  delta: delta of the waitgroup
 * 	val: value of the waitgroup after the operation
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateWaitGroupAdd(id uint64, delta int, val int32) int {
	timer := GetNextTimeStep()

	var file string
	var line int
	if delta > 0 {
		_, file, line, _ = Caller(2)
	} else {
		_, file, line, _ = Caller(3)
	}

	elem := "W," + uint64ToString(timer) + "," + uint64ToString(timer) + "," +
		uint64ToString(id) + ",A," +
		intToString(delta) + "," + int32ToString(val) + "," + file + ":" +
		intToString(line)

	return insertIntoTrace(elem, false)

}

/*
 * AdvocateWaitGroupWaitPre adds a waitgroup wait to the trace
 * MARK: Wait Pre
 * Args:
 * 	id: id of the waitgroup
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateWaitGroupWaitPre(id uint64) int {
	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	elem := "W," + uint64ToString(timer) + ",0," + uint64ToString(id) +
		",W,0,0," + file + ":" + intToString(line)

	return insertIntoTrace(elem, false)
}

/*
 * AdvocateWaitGroupWaitPost adds the end counter to an operation of the trace
 * MARKL: Wait Post
 * Args:
 * 	index: index of the operation in the trace
 */
func AdvocateWaitGroupPost(index int) {
	timer := GetNextTimeStep()

	// internal elements are not in the trace
	if index == -1 {
		return
	}

	// only needed to fix tests

	if currentGoRoutine() == nil {
		return
	}

	elem := currentGoRoutine().getElement(index)
	split := splitStringAtCommas(elem, []int{2, 3})
	split[1] = uint64ToString(timer)
	elem = mergeString(split)

	currentGoRoutine().updateElement(index, elem)
}
