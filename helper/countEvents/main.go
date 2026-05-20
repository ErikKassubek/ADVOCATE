package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var runtimeData = map[string]data{}

// TODO: run, record, replay times

type data struct {
	run        float64
	record     float64
	replay     float64
	replSuc    int
	replTotal  int
	numberTest int
	maxNumRout int
	avgNumRout float64
	maxNumEv   int
	avgNumEv   float64
	maxNumCom  int
	avgNumCom  float64
}

type fuzz struct {
	version string
	goleak  int
	gfuzz   int
	goPie   int
}

var fuzzData = map[string]fuzz{
	"argo-cd":     {"v3.1.0", 0, 2, 4},
	"bleve":       {"v2.5.3", 2, 3, 5},
	"bosun":       {"v0.8.0", 0, 0, 0},
	"caddy":       {"v2.10.0", 0, 1, 1},
	"dns":         {"v1.1.50", 0, 0, 0},
	"etcd":        {"v3.6.11", 0, 0, 0},
	"flannel":     {"v0.20.2", 0, 0, 0},
	"frp":         {"v0.63.0", 0, 0, 0},
	"gin":         {"v1.10.1", 0, 0, 0},
	"go-ethereum": {"v1.17.3", 0, 0, 0},
	"gofiber":     {"v2.40.1", 0, 0, 1},
	"gorums":      {"v0.7.0", 1, 1, 1},
	"grpc":        {"v1.51.0", 1, 1, 1},
	"hugo":        {"v0.148.2", 0, 0, 0},
	"kubernetes":  {"v1.25.5", 8, 8, 8},
	"nsq":         {"v1.3.0", 0, 0, 3},
	"octant":      {"v0.25.1", 0, 0, 0},
	"ollama":      {"v0.11.4", 2, 2, 2},
	"pholcus":     {"v1.3.4", 0, 0, 0},
	"prometheus":  {"v3.11.3", 0, 0, 0},
	"syncthing":   {"v1.22.1", 2, 6, 6},
	"terraform":   {"v1.12.2", 1, 2, 2},
	"zinx":        {"v1.2.7", 0, 0, 0},
}

func main() {
	rootPath := "/home/advocate/Advocate/Experiments/Progs/GoCDR/"

	progs, _ := os.ReadDir(rootPath)
	for _, prog := range progs {

		totalNumRout := 0
		totalNumEv := 0
		totalNumCom := 0

		maxNumRout := 0
		maxNumEv := 0
		maxNumCom := 0

		numberTests := 0

		replayTotal := 0
		replaySuc := 0

		resPath := filepath.Join(rootPath, prog.Name(), "gocdrResult")

		tests, _ := os.ReadDir(resPath)
		for _, test := range tests {
			if strings.HasSuffix(test.Name(), "_replay") {
				outPath := filepath.Join(resPath, test.Name(), "output", "output.log")

				replayTotal += 1
				replaySuc += readReplay(outPath)

				continue
			}

			tracesPath := filepath.Join(resPath, test.Name(), "traces")

			traces, _ := os.ReadDir(tracesPath)

			if len(traces) == 0 {
				continue
			}

			tracePath := filepath.Join(tracesPath, traces[0].Name())

			numRout, numEvents, numCom := readTrace(tracePath)

			totalNumRout += numRout
			totalNumEv += numEvents
			totalNumCom += numCom

			maxNumRout = max(maxNumRout, numRout)
			maxNumEv = max(maxNumEv, numEvents)
			maxNumCom = max(maxNumCom, numCom)

			numberTests += 1
		}

		if numberTests == 0 {
			continue
		}

		avgNumRout := float64(totalNumRout) / float64(numberTests)
		avgNumEv := float64(totalNumEv) / float64(numberTests)
		avgNumCom := float64(totalNumCom) / float64(numberTests)

		rd := data{}
		if d, ok := runtimeData[prog.Name()]; !ok {
			rd = d
		}

		rd.numberTest = numberTests
		rd.maxNumRout = maxNumRout
		rd.avgNumRout = avgNumRout
		rd.maxNumEv = maxNumEv
		rd.avgNumEv = avgNumEv
		rd.maxNumCom = maxNumCom
		rd.avgNumCom = avgNumCom
		rd.replSuc = replaySuc
		rd.replTotal = replayTotal

		runtimeData[prog.Name()] = rd
	}

	printResults()

}

func readReplay(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Exit Replay with code  0 ") {
			return 1
		}
	}
	return 0
}

func readTrace(path string) (int, int, int) {
	numRouts := 0
	numEv := 0
	numCom := 0

	routs, _ := os.ReadDir(path)

	for _, rout := range routs {
		if rout.Name() == "trace_info.log" {
			continue
		}

		routPath := filepath.Join(path, rout.Name())

		ne, nc := readTraceRout(routPath)
		numEv += ne
		numCom += nc
		numRouts += 1
	}

	return numRouts, numEv, numCom
}

func readTraceRout(path string) (int, int) {
	numEv := 0
	numCom := 0

	file, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		switch line[0] {
		case 'G', 'E', 'O', 'A':
			numEv += 1
			numCom += 1
		case 'C':
			fields := strings.Split(line, ",")
			if fields[2] == "0" { // no commit
				numEv += 1
			} else {
				if fields[4] == "C" { // close
					numEv += 1
				} else {
					numEv += 2
				}
				numCom += 1
			}
		case 'S':
			fields := strings.Split(line, ",")
			if fields[2] == "0" { // no commit
				numEv += 1
			} else {
				numEv += 2
				numCom += 1
			}
		case 'M':
			fields := strings.Split(line, ",")
			if fields[2] == "0" { // no commit
				numEv += 1
			} else {
				if fields[4] == "U" || fields[4] == "N" { // unlock cannot block
					numEv += 1
				} else {
					numEv += 2
				}

				numCom += 1
			}
		case 'W', 'D':
			fields := strings.Split(line, ",")
			if fields[2] == "0" { // no commit
				numEv += 1
			} else {
				if fields[4] == "W" { // only wait blocks
					numEv += 2
				} else {
					numEv += 1
				}

				numCom += 1
			}
		}
	}

	return numEv, numCom
}

func readRuntime() {
	path := "/home/advocate/Advocate/Advocate/helper/runnerGoCDR/runner_2026-05-20-13-45-58.log"

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "FINISH:") {
			continue
		}

		fields := strings.Split(line, ", ")
		name := strings.Split(fields[0], ": ")[1]
		run := toFloat(strings.Split(fields[1], ": ")[1])
		record := toFloat(strings.Split(fields[2], ": ")[1])
		replay := toFloat(strings.Split(fields[3], ": ")[1])

		rd := data{}
		if d, ok := runtimeData[name]; !ok {
			rd = d
		}

		rd.run = run
		rd.record = record
		rd.replay = replay

		runtimeData[name] = rd

	}
}

func toFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return 0
	}
	return f
}

func printResults() {
	names := make([]string, 0, len(runtimeData))
	for k := range runtimeData {
		names = append(names, k)
	}

	// Sort keys alphabetically
	sort.Strings(names)

	for _, name := range names {

		d := runtimeData[name]
		f := fuzzData[name]

		sucRate := float64(d.replSuc) / float64(d.replTotal)

		recordOverhead := (d.record - d.run) / d.run * 100
		replayOverhead := (d.record - d.run) / d.run * 100

		fmt.Printf("%s & %s & %d & %d & %d & %.1f & %.1f & %.1f & %.3f & %d & %d & %d \\\\\n", name, f.version, d.numberTest, d.maxNumRout, d.maxNumEv, d.run, d.record, d.replay, sucRate, f.goleak, f.gfuzz, f.goPie)
		fmt.Printf(" &  &  & %.1f & %.1f & & %.1f & %.1f & & &  & \\\\ \\hline\n", d.avgNumRout, d.avgNumEv, recordOverhead, replayOverhead)
	}
}
