package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const goBenchPath = "/drives/D/AdvocateResults/Experiments/ablationStudyLong/withReplay/advocateResult"

func main() {

	entries, err := os.ReadDir(goBenchPath)
	if err != nil {
		log.Fatal(err)
	}

	exec, total, leak := 0, 0, 0

	for _, entry := range entries {
		if entry.IsDir() {
			path := filepath.Join(goBenchPath, entry.Name())
			e, t, l := CheckTest(path)
			exec += e
			total += t
			leak += l
		}
	}

	if total == 0 {
		println("0")
	} else {
		perc := float64(exec) / float64(total) * 100
		perc2 := float64(leak) / (float64(total) + float64(leak)) * 100
		fmt.Println(perc, perc2)
	}
}

func CheckTest(path string) (int, int, int) {
	if !containsBugs(path) {
		return 0, 0, 0
	}

	bugPath := filepath.Join(path, "bugs")

	traceFiles, err := os.ReadDir(path)

	leak := 0
	total := 0
	exec := 0

	if err != nil {
		log.Fatal(err)
	}
	traces := make([]int, 0)

	for _, file := range traceFiles {
		if !file.IsDir() {
			continue
		}
		if !strings.HasPrefix(file.Name(), "advocateTrace_") {
			continue
		}
		number, err := strconv.Atoi(strings.Split(file.Name(), "_")[1])
		if err != nil {
			panic(err)
		}
		traces = append(traces, number)
	}

	sort.Ints(traces)

	// fmt.Println(traces)

	bugFiles, err := os.ReadDir(bugPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range bugFiles {
		if strings.HasPrefix(file.Name(), "leak_0_") {
			leak++
			continue
		}
		if lingering(filepath.Join(path, "bugs", file.Name())) {
			continue
		}
		number, err := strconv.Atoi(strings.Split(file.Name(), "_")[1])
		if err != nil {
			panic(err)
		}

		tracePathBug := filepath.Join(path, fmt.Sprintf("advocateTrace_%d", traces[number]))
		if isSCExec(tracePathBug) {
			exec++
		}
		total++
	}

	return exec, total, leak

}

func isSCExec(path string) bool {
	infoPath := filepath.Join(path, "trace_info.log")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "ActiveReached") {
			if strings.Split(line, "!")[1] == "1" {
				return true
			}
		}
	}

	return false
}

func containsBugs(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() == "bugs" {
			return true
		}
	}

	return false
}

func lingering(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(data), "\n")

	return lines[0] == "# Leak: Leak"
}
