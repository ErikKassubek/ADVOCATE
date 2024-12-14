package advocate

import (
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

/*
 * Write the trace of the program to a file.
 * The trace is written in the file named file_name.
 * The trace is written in the format of advocate.
 */
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

	runtime.DisableTrace()

	writeToTraceFiles(tracePathRecorded)
}

/*
 * Write the trace to a set of files. The traces are written into a folder
 * with name trace. For each routine, a file is created. The file is named
 * trace_routineId.log. The trace of the routine is written into the file.
 */
func writeToTraceFiles(tracePath string) {
	numRout := runtime.GetNumberOfRoutines()
	var wg sync.WaitGroup
	for i := 1; i <= numRout; i++ {
		// write the trace to the file
		wg.Add(1)
		go writeToTraceFile(i, &wg, tracePath)
	}

	wg.Wait()
}

/*
 * Write the trace of a routine to a file.
 * The trace is written in the file named trace_routineId.log.
 * The trace is written in the format of advocate.
 * Args:
 * 	- routine: The id of the routine
 */
func writeToTraceFile(routine int, wg *sync.WaitGroup, tracePath string) {
	// create the file if it does not exist and open it
	defer wg.Done()

	// if runtime.TraceIsEmptyByRoutine(routine) {
	// 	return
	// }

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
 * Delete empty files in the trace folder.
 * The function deletes all files in the trace folder that are empty.
 */
func deleteEmptyFiles() {
	files, err := os.ReadDir(tracePathRecorded)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		stat, err := os.Stat(tracePathRecorded + "/" + file.Name())
		if err != nil {
			continue
		}
		if stat.Size() == 0 {
			err := os.Remove(tracePathRecorded + "/" + file.Name())
			if err != nil {
				panic(err)
			}
		}
	}

}

/*
 * InitTracing initializes the tracing.
 * The function creates the trace folder and starts the background memory test.
 * Args:
 */
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
		if !runtime.GetAdvocateDisabled() {
			FinishTracing()
		}
		os.Exit(1)
	}()

	// go writeTraceIfFull()
	// go removeAtomicsIfFull()
	runtime.InitAdvocate()
}
