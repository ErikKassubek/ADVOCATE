package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const gocdrPath = "/home/advocate/Advocate/Advocate/gocdr/"
const srcPath = "/home/advocate/Advocate/Experiments/Progs/GoCDR/"

var progs = []string{
	"zinx",
	"argo-cd",
	"bleve",
	"bosun",
	"caddy",
	"dns",
	"etcd",
	"frp",
	"gin",
	// "fabiolb",
	"go-ethereum",
	"gorums",
	"grpc",
	"kubernetes",
	"moby",
	"nsq",
	"ollama",
	"pholcus",
	"prometheus",
	"syncthing",
	"terraform",
}

var now = time.Now()
var id = now.Format("2006-01-02-15-04-05")
var logName = "runner_" + id + ".log"

var modes = []string{
	// "GoPie",
	// "GFuzz",
	// "GoCR",
}

var settings = "-timeoutRec 30 -timeoutRep 30 -noInfo"

var maxWorker = 10

var fileMutex sync.Mutex

func main() {
	readFlag()
	log("START RUNNING")
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, maxWorker)
	// for _, mode := range modes {
	for i, p := range progs {
		wg.Add(1)
		// m := mode
		path := filepath.Join(srcPath, p)
		go worker(path, p, &wg, sem, i == len(progs)-1)
		time.Sleep(time.Second)
	}
	// }
	wg.Wait()
	log("END RUNNING")
}

func readFlag() {
	valuesFlag := flag.String("p", "", "Space-separated list of progs")
	threadsFlags := flag.String("w", "", "Number of progs run parallel")

	// Parse the flags
	flag.Parse()

	// Split the input string by spaces
	if *valuesFlag != "" {
		progs = strings.Fields(*valuesFlag)
	}

	if *threadsFlags != "" {
		var err error
		maxWorker, err = strconv.Atoi(*threadsFlags)
		if err != nil {
			fmt.Println("Error:", err)
			panic("Invalid value t")
		}
	}

	log("Run: " + id + " " + *valuesFlag + " with " + *threadsFlags + " workers")
}

func worker(path, name string, wg *sync.WaitGroup, sem chan struct{}, last bool) {
	defer wg.Done()
	sem <- struct{}{}

	log(fmt.Sprintf("START : %s", name))

	start := time.Now()
	run(path, name)
	timeRun := time.Since(start)

	start = time.Now()
	record(path, name)
	timeRecord := time.Since(start)

	start = time.Now()
	nrTests := replay(path, name)
	timeReplay := time.Since(start)

	log(fmt.Sprintf("FINISH: %s, RUN: %f, RECORD: %f, REPLAY: %f, NUM: %d", name, timeRun.Seconds(), timeRecord.Seconds(), timeReplay.Seconds(), nrTests))

	if !last {
		time.Sleep(30 * time.Minute)
	}

	<-sem
}

func run(path, name string) {
	cmdStr := fmt.Sprintf("run -path %s %s", path, settings)
	cmd := exec.Command("./gocdr", strings.Split(cmdStr, " ")...)

	// Set the working directory
	cmd.Dir = gocdrPath

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log(fmt.Sprintf("ERROR RUN: %s %s", name, err.Error()))
	}
}

func record(path, name string) {
	cmdStr := fmt.Sprintf("record -path %s %s", path, settings)
	cmd := exec.Command("./gocdr", strings.Split(cmdStr, " ")...)

	// Set the working directory
	cmd.Dir = gocdrPath

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log(fmt.Sprintf("ERROR RECORD: %s %s", name, err.Error()))
	}
}

func replay(path, name string) int {
	outpath := filepath.Join(path, "gocdrResult")

	entries, _ := os.ReadDir(outpath)

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(outpath, entry.Name(), "traces")

			traces, _ := os.ReadDir(fullPath)
			if len(traces) == 0 {
				continue
			}

			fullPath = filepath.Join(fullPath, traces[0].Name())

			testNameSl := strings.Split(entry.Name(), "-")
			if len(testNameSl) == 0 {
				continue
			}

			testName := testNameSl[len(testNameSl)-1]

			cmdStr := fmt.Sprintf("replay -path %s -exec %s -trace %s %s -keepResultDir", path, testName, fullPath, settings)
			cmd := exec.Command("./gocdr", strings.Split(cmdStr, " ")...)

			// Set the working directory
			cmd.Dir = gocdrPath

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				log(fmt.Sprintf("ERROR REPLAY: %s %s", name, err.Error()))
			}
		}
	}

	return len(entries)

}

func log(c string) {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	now := time.Now().Format("2006/01/02 15:04:05")

	_, err = file.WriteString(now + " " + c + "\n")

	fmt.Println(now + " " + c)
}
