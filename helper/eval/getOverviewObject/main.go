package main

import (
	"bufio"
	"errors"
	"flag"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// small program to print the trace of one of the objects in a trace

func getElemFromFiles(filePath string, objectID int, start, end int, disableAtomics bool) (map[int]string, error) {
	maxTokenSize := 4

	// traverse all files in the folder
	files, err := os.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	elements := make(map[int]string)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		routine, err := getRoutineFromFileName(file.Name())
		if err != nil {
			continue
		}

		file, err := createTraceFromFile(filePath+"/"+file.Name(), routine, maxTokenSize, objectID, start, end, disableAtomics)
		if err != nil {
			return nil, err
		}

		for k, v := range file {
			elements[k] = v
		}
	}

	return elements, nil
}

func getRoutineFromFileName(fileName string) (int, error) {
	if !strings.HasSuffix(fileName, ".log") {
		return 0, errors.New("File name does not end with .log")
	}

	if !strings.HasPrefix(fileName, "trace_") {
		return 0, errors.New("File name does not start with trace_")
	}

	routineStr := strings.TrimSuffix(strings.TrimPrefix(fileName, "trace_"), ".log")

	routine, err := strconv.Atoi(routineStr)
	if err != nil {
		return 0, err
	}

	return routine, nil
}

func createTraceFromFile(filePath string, routine int, maxTokenSize int, objectID int, start, end int, disableAtomics bool) (map[int]string, error) {
	mb := 1048576 // 1 MB

	elements := make(map[int]string)
	routineStr := strconv.Itoa(routine)
	if len(routineStr) == 1 {
		routineStr = "0" + routineStr
	}

	for {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, maxTokenSize*mb), maxTokenSize*mb)

		for scanner.Scan() {
			line := scanner.Text()
			res := processLine(line, objectID, start, end, disableAtomics, routineStr)
			for k, v := range res {
				elements[k] = routineStr + " -> " + v
			}
		}

		file.Close()

		if err := scanner.Err(); err != nil {
			if err == bufio.ErrTooLong {
				maxTokenSize *= 2 // max buffer was to short, restart
				// println("Increase max file size to " + strconv.Itoa(maxTokenSize) + "MB")
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	return elements, nil
}

func processLine(line string, objectID, start, end int, disableAtomics bool, routineStr string) map[int]string {
	elements := strings.Split(line, ";")
	result := make(map[int]string)
	for _, element := range elements {
		res, tPost := processElement(element, strconv.Itoa(objectID), start, end, disableAtomics)
		if res {
			if tPost == 0 {
				println(routineStr + " -> " + element)
			} else {
				result[tPost] = element
			}
		}
	}
	return result
}

func processElement(element string, objectID string, startTime int, endTime int, disableAtomics bool) (bool, int) {
	if element == "" {
		return false, 0
	}

	fields := strings.Split(element, ",")
	switch fields[0] {
	case "A":
		if disableAtomics {
			return false, 0
		}
		time, _ := strconv.Atoi(fields[1])
		if !isIdValid(fields[2], objectID) || !isValidTime(time, startTime, endTime) {
			return false, 0
		}
		return true, time
	case "C", "M", "W", "O", "N":
		time, _ := strconv.Atoi(fields[2])
		if !isIdValid(fields[3], objectID) || !isValidTime(time, startTime, endTime) {
			return false, 0
		}
		return true, time
	case "G":
		time, _ := strconv.Atoi(fields[1])
		if !(objectID == "-1") || !isValidTime(time, startTime, endTime) {
			return false, 0
		}
		return true, time
	case "S":
		time, _ := strconv.Atoi(fields[2])
		if fields[3] == objectID {
			return true, time
		}
		if isValidTime(time, startTime, endTime) {
			return true, time
		}

		return false, time
	case "X":
		time, _ := strconv.Atoi(fields[1])
		if objectID == "-1" || !isValidTime(time, startTime, endTime) {
			return true, time
		}
	}

	return false, 0
}

func isValidTime(tPost, start, end int) bool {
	return tPost == 0 || (tPost >= start && tPost <= end)
}

func isIdValid(id, objectID string) bool {
	return objectID == "-1" || id == objectID
}

func main() {
	pathTrace := flag.String("t", "", "Path to the trace folder to analyze or rewrite")
	objectID := flag.Int("o", -1, "Object ID to print the trace for")
	start := flag.Int("s", 1, "Start time")
	end := flag.Int("e", math.MaxInt, "End time")
	disableAtomics := flag.Bool("a", false, "Disable atomic operations")
	flag.Parse()

	if pathTrace == nil || *pathTrace == "" {
		println("Please provide a path to the trace folder")
		return
	}

	res, err := getElemFromFiles(*pathTrace, *objectID, *start, *end, *disableAtomics)
	if err != nil {
		println(err.Error())
		return
	}

	keys := make([]int, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	for _, k := range keys {
		println(res[k])
	}
}
