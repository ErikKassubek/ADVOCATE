package runtime

/*
 * AdvocateSpawnCaller adds a routine spawn to the trace
 * Args:
 * 	callerRoutine: routine that created the new routine
 * 	newID: id of the new routine
 * 	file: file where the routine was created
 * 	line: line where the routine was created
 */
func AdvocateSpawnCaller(callerRoutine *AdvocateRoutine, newID uint64, file string, line int32) {
	timer := GetNextTimeStep()

	if AdvocateIgnore(file) {
		return
	}

	elem := "G," + uint64ToString(timer) + "," + uint64ToString(newID) + "," + file + ":" + int32ToString(line)

	callerRoutine.addToTrace(elem)
}

/*
 * Record the finish of a routine
 */
func AdvocatRoutineExit() {
	timer := GetNextTimeStep()
	elem := "E," + uint64ToString(timer)
	insertIntoTrace(elem)
}
