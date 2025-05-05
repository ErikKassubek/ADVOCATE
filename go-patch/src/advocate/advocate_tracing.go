package advocate

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var traceFileCounter = 0
var tracePathRecorded = "advocateTrace"

var hasFinished = false

var timerStarted = false
var startTime time.Time
var duration time.Duration

// InitTracing initializes the tracing.
// The function creates the trace folder and starts the background memory test.
func InitTracing() {
	// if the program panics, but is not in the main routine, no trace is written
	// to prevent this, the following is done. The corresponding send/recv are in the panic definition
	blocked := make(chan struct{})
	writingDone := make(chan struct{})
	runtime.GetAdvocatePanicChannels(blocked, writingDone)
	go func() {
		<-blocked
		FinishTracing()
		writingDone <- struct{}{}
	}()

	// if the program is terminated by the user, the defer in the header
	// is not executed. Therefore capture the signal and write the trace.
	interuptSignal := make(chan os.Signal, 1)
	signal.Notify(interuptSignal, os.Interrupt)
	go func() {
		<-interuptSignal
		println("\nCancel Run. Write trace. Cancel again to force exit.")
		go func() {
			<-interuptSignal
			os.Exit(1)
		}()
		if runtime.IsTracingEnabled() {
			FinishTracing()
		}
		os.Exit(1)
	}()

	startTime = time.Now()
	timerStarted = true

	// go writeTraceIfFull()
	// go removeAtomicsIfFull()
	runtime.InitTracing(FinishTracing)
}

// Write the trace of the program to a file.
// The trace is written in the file named file_name.
// The trace is written in the format of advocate.
func FinishTracing() {
	if hasFinished {
		return
	}
	hasFinished = true

	// remove the trace folder if it exists
	err := os.RemoveAll(tracePathRecorded)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}

	// create the trace folder
	err = os.Mkdir(tracePathRecorded, 0755)
	if err != nil {
		if !os.IsExist(err) {
			panic(err)
		}
	}

	runtime.AdvocatRoutineExit()

	time.Sleep(100 * time.Millisecond)

	runtime.DisableTracing()

	if timerStarted {
		duration = time.Since(startTime)
	}

	writeToTraceFiles(tracePathRecorded)
}

// Write the trace to a set of files. The traces are written into a folder
// with name trace. For each routine, a file is created. The file is named
// trace_routineId.log. The trace of the routine is written into the file.
//
// Parameter:
//
//	tracePath string: path to where the trace should be written
func writeToTraceFiles(tracePath string) {
	numRout := runtime.GetNumberOfRoutines()
	var wg sync.WaitGroup
	for i := 1; i <= numRout; i++ {
		// write the trace to the file
		wg.Add(1)
		go writeToTraceFile(i, &wg, tracePath)
	}

	writeToTraceFileInfo(tracePath, numRout)

	wg.Wait()
}

// Write the trace of a routine to a file.
// The trace is written in the file named trace_routineId.log.
// The trace is written in the format of advocate.
//
// Parameter:
//   - routine: The id of the routine
//   - wg *sync.WaitGroup: wait group used to make writing of different routines concurrent
//   - tracePath string: path to where the trace should be written
func writeToTraceFile(routine int, wg *sync.WaitGroup, tracePath string) {
	// create the file if it does not exist and open it
	defer wg.Done()

	fileName := filepath.Join(tracePath, "trace_"+strconv.Itoa(routine)+".log")

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// get the runtime to send the trace
	advocateChan := make(chan string)
	go func() {
		runtime.TraceToStringByIDChannel(routine, advocateChan)
		close(advocateChan)
	}()

	// receive the trace and write it to the file
	for trace := range advocateChan {
		if _, err := file.WriteString(trace); err != nil {
			panic(err)
		}
	}
}

/*
 * Write a trace info file
 */
func writeToTraceFileInfo(tracePath string, numberRoutines int) {
	fileName := filepath.Join(tracePath, "trace_info.log")

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	exitCode, exitPos := runtime.GetExitCode()
	replayOldest, replayDisabled, replayAck := runtime.GetReplayStatus()
	file.WriteString(fmt.Sprintf("ExitCode!%d\n", exitCode))
	file.WriteString(fmt.Sprintf("ExitPosition!%s\n", exitPos))
	file.WriteString(fmt.Sprintf("ReplayTimeout!%d\n", replayOldest))
	file.WriteString(fmt.Sprintf("ReplayDisabled!%d\n", replayDisabled))
	file.WriteString(fmt.Sprintf("ReplayAck!%d\n", replayAck))
	file.WriteString(fmt.Sprintf("NumberRoutines!%d\n", numberRoutines))
	if timerStarted {
		file.WriteString(fmt.Sprintf("Runtime!%d", int(duration.Seconds())))
	} else {
		file.WriteString("Runtime:0")
	}

}

// Delete empty files in the trace folder.
// The function deletes all files in the trace folder that are empty.
// func deleteEmptyFiles() {
// 	files, err := os.ReadDir(tracePathRecorded)
// 	if err != nil {
// 		panic(err)
// 	}

// 	for _, file := range files {
// 		if file.IsDir() {
// 			continue
// 		}

// 		stat, err := os.Stat(tracePathRecorded + "/" + file.Name())
// 		if err != nil {
// 			continue
// 		}
// 		if stat.Size() == 0 {
// 			err := os.Remove(tracePathRecorded + "/" + file.Name())
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	}

// }
