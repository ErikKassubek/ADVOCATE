package experiments

import (
	"aplas/data"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var defaultSettings = []string{
	"fuzzing",
	"-mode",
	"GoCRHB",
	// "-noInfo",
	"-noProgress",
	"-noWarning",
	"-timeoutFuz",
	"600",
	"-maxFuzzingRuns",
	"20",
	"-stats",
	"-deleteTrace",
}

var maxWorker = 10
var wg = sync.WaitGroup{}
var sem = make(chan struct{}, maxWorker)

func Run() {
	_, err := initialize(true, data.PathOutput)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	base()

	time.Sleep(time.Second)

	oocLength()
	time.Sleep(time.Second)

	sameElement()
	time.Sleep(time.Second)

	overhead()
	time.Sleep(time.Second)

	realWorld()

	time.Sleep(time.Second)
	wg.Wait()
}

func initialize(main bool, name ...string) (string, error) {
	path, err := data.CreateOutput(name...)
	if err != nil {
		return path, err
	}

	if main {
		err = os.Mkdir(data.PathStats, 0755)
		if err != nil {
			return path, err
		}
		err = os.Mkdir(data.PathLog, 0755)
	} else {
		err = data.CopyGobench(path)
	}

	return path, err
}

func start() {
	wg.Add(1)
	sem <- struct{}{}
}

func done() {
	<-sem
	wg.Done()
}

func writeData(name, id string, t time.Duration, out string) error {
	foundBugs := "0"
	numFuzzRun := "0"

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Tests with indicated bugs: ") {
			foundBugs = strings.TrimSuffix(strings.Split(line, ": ")[1], "[0m")
		} else if strings.Contains(line, "Finish fuzzing after") {
			numFuzzRun = strings.Split(line, " ")[5]
		}
	}

	data.OutputLock.Lock()
	defer data.OutputLock.Unlock()

	f_time, err := os.OpenFile(filepath.Join(data.PathStats, "rt_"+name+".log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f_time.Close()

	nfr, err := strconv.Atoi(numFuzzRun)
	if err != nil {
		fmt.Println(err)
	}

	avgTime := t.Milliseconds()
	if nfr > 1 {
		avgTime = int64(float64(t.Milliseconds()) / float64(nfr))
	}

	f_time.WriteString(fmt.Sprintf("%s-%d-%d-%d\n", id, nfr, t.Milliseconds(), avgTime))

	if out == "" {
		return err
	}

	f_prec, err := os.OpenFile(filepath.Join(data.PathStats, "prec_"+name+".log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f_prec.Close()

	f_prec.WriteString(fmt.Sprintf("%s-%s\n", id, foundBugs))

	return nil
}
