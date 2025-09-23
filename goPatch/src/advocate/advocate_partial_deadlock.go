package advocate

import (
	"runtime"
	"time"
)

// DetectPartialDeadlock runs a partial deadlock detection in the current execution
// Parameter:
//   - loop bool: if true, run a
func DetectPartialDeadlock(interval int) {
	runtime.AdvocatePDDetectionStopped = false

	go func() {
		for {
			if runtime.AdvocatePDDetectionStopped {
				return
			}

			runtime.AdvocateDetectPD()

			if interval <= 0 {
				return
			}

			time.Sleep(1 * time.Millisecond)
		}
	}()
}

// StopPartialDeadlockDetection stops the partial deadlock detection before
// the start of a new run
// If the detection is currently not run in a loop, this has no effect
func StopPartialDeadlockDetection() {
	runtime.AdvocatePDDetectionStopped = true
}
