package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	red = "\033[31m"
	org = "\033[33m"
	grn = "\033[32m"
	end = "\033[0m"
)

var (
	home, _        = os.UserHomeDir()
	pathToAdvocate = home + "/Uni/HiWi/ADVOCATE"
	pathToExamples = pathToAdvocate + "/examples"
	analyzerPath   = pathToAdvocate + "/analyzer"
	runEval        = pathToAdvocate + "/eval"
	overviewPath   = pathToAdvocate + "/eval/createOverview"

	// name
	// pathToExec
	// pathToProg
	// basename of executable
	// params (if any)
	programs = [][]string{}

	resPath = runEval

	resSucc      = []string{}
	resFailure   = []string{}
	numberErrors = 0
)

func addConstructed() {
	for i := 1; i <= 44; i++ {
		programs = append(programs, []string{
			"constructed " + strconv.Itoa(i),
			pathToExamples + "/constructed/",
			pathToExamples + "/constructed/",
			"constructed", "-c", strconv.Itoa(i), "-t", "5"})
	}
}

func addGoBench() {
	const n = 18
	names := [n]string{
		"cockroach1055",
		"cockroach1462",
		"etcd6873",
		"etcd7443",
		"etcd7492",
		"etcd7902",
		"grpc1353",
		"grpc1460",
		"grpc1687",
		"istio16224",
		"Kubernetes1321",
		"Kubernetes6632",
		"Kubernetes10182",
		"Kubernetes26980",
		"Moby28462",
		"serving2137",
		"serving3068",
		"serving5865",
	}

	for i := 0; i < n; i++ {
		programs = append(programs,
			[]string{
				"Gobench " + strconv.Itoa(i+1) + ": " + names[i],
				pathToExamples + "/GoBench/",
				pathToExamples + "/GoBench/" + names[i] + "/",
				"gobench",
				"-c", strconv.Itoa(i + 1), "-t", "10"})
	}
}

func addMediumPrograms() {
	// bbolt
	programs = append(programs,
		[]string{
			"bbolt",
			pathToExamples + "/bbolt/cmd/bbolt/",
			pathToExamples + "/bbolt/",
			"bbolt",
			"bench"})

	// gocrawl
	programs = append(programs,
		[]string{
			"gocrawl",
			pathToExamples + "/gocrawl/",
			pathToExamples + "/gocrawl/",
			"gocrawl"})

	// htcat
	programs = append(programs,
		[]string{
			"htcat",
			pathToExamples + "/htcat/cmd/htcat/",
			pathToExamples + "/htcat/",
			"htcat"})

	// pgzip
	programs = append(programs,
		[]string{
			"pgzip",
			pathToExamples + "/pgzip/",
			pathToExamples + "/pgzip/",
			"pgzip"})

	// sorty
	programs = append(programs,
		[]string{
			"sorty",
			pathToExamples + "/sorty/",
			pathToExamples + "/sorty/",
			"sorty"})
}

func runExecs(pathToExec string, execName string, execArgs []string, progName string, replay bool) error {
	success := 0
	var err error
	max := 4
	for i := 1; i <= max; i++ {
		err = nil // reset error

		err = runExec(pathToExec, execName, execArgs, progName, "original", 0)
		if err != nil {
			if i != max {
				resetResFolder(progName)
				continue
			}
		}
		success = 1

		err = runExec(pathToExec, execName, execArgs, progName, "advocate", 0)
		if err != nil {
			if i != max {
				resetResFolder(progName)
				continue
			}
		}
		success = 2

		if replay {
			err = runExec(pathToExec, execName, execArgs, progName, "replay", 0)
			if err != nil {
				if i != max {
					resetResFolder(progName)
					continue
				}
			}

		}
		success = 3

		break
	}

	switch success {
	case 0:
		return errors.New("Failed to run original")
	case 1:
		return errors.New("Failed to run advocate")
	case 2:
		return errors.New("Failed to run replay")
	}

	return nil
}

func runExec(pathToExec string, execName string, execArgs []string, progName string, variant string, repeat int) error {
	commandStr := "./" + execName + "_" + variant
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, commandStr, execArgs...)
	cmd.Dir = pathToExec

	log(EXEC, cmd.String()+" in "+pathToExec)

	out, runtimeOriginal, err := runCommand(cmd)

	if err != nil {
		log(ERROR, err.Error()+":\n"+out)
		return errors.New("Error")
	}

	if out == "Timeout" {
		log(TIMEOUT, variant+" of "+progName+" timed out. Trying again...")
		runExec(pathToExec, execName, execArgs, progName, variant, repeat+1)
		return errors.New("Timeout")
	}

	err = writeTime(runtimeOriginal, progName)
	if err != nil {
		log(ERROR, err.Error())
		return errors.New("Write time failed")
	}

	return nil
}

/*
 * Run the command
 * Args:
 *   cmd (*exec.Cmd): The command to run
 * Returns:
 *   (string): stderr output
 *   (time.Duration): runtime
 *   (error): error
 */
func runCommand(cmd *exec.Cmd) (string, time.Duration, error) {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	start := time.Now()

	err := cmd.Run()

	duration := time.Since(start)

	if err != nil {
		return stderr.String(), duration, err
	}

	return "", duration, nil
}

func runAnalyzer(progName string, pathToTrace string) (string, error) {
	cmdStr := "./analyzer"
	resOut := resPath + "/" + progName + "/"
	cmdArgs := []string{"-t", pathToTrace, "-x", "-p", "-r", resOut}
	cmd := exec.Command(cmdStr, cmdArgs...)
	cmd.Dir = analyzerPath

	log(ANALYZE, cmd.String()+" in "+analyzerPath)

	out, time, err := runCommand(cmd)

	writeTime(time, progName)

	return out, err
}

func writeTime(time time.Duration, programName string) error {
	timeStr := fmt.Sprintf("%f", time.Seconds())
	file, err := os.OpenFile(resPath+"/"+programName+"/times.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(timeStr + ",")

	return nil
}

func createOverview(progName string, progPath string, execPath string) error {
	cmdStr := "./overview"
	resOut := resPath + "/" + progName
	cmdArgs := []string{
		"-n", "overview",
		"-p", progPath,
		"-s", resOut,
		"-t", execPath + "trace/",
		"-r", resOut + "/results_readable.log",
		"-d", resOut + "/times.log"}
	cmd := exec.Command(cmdStr, cmdArgs...)
	cmd.Dir = overviewPath

	log(OVERVIEW, cmd.String()+" in "+overviewPath)
	err := cmd.Run()
	return err
}

func createResFolder(progName string) error {
	progFolder := resPath + "/" + progName

	log(CREATE, progName+" at "+progFolder)

	err := os.MkdirAll(progFolder, 0755)
	if err != nil {
		return err
	}

	return nil
}

func resetResFolder(progName string) error {
	progFolder := resPath + "/" + progName
	err := os.RemoveAll(progFolder)
	if err != nil {
		return err
	}

	err = os.MkdirAll(progFolder, 0755)
	if err != nil {
		return err
	}
	return nil
}

func finish() error {
	log(CLEANUP, "Create overview at "+resPath+"/overview.log")

	overheadsAdvocate := []float64{}
	overheadsReplay := []float64{}
	fractionAnalyses := []float64{}

	// read all times.log files and save the values to a variable
	err := filepath.Walk(resPath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if info.Name() != "times.log" {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			buf := make([]byte, 1024)
			n, err := file.Read(buf)
			if err != nil {
				return err
			}

			timesStr := string(buf[:n-1])
			times := strings.Split(timesStr, ",")

			original := 0.0
			advocate := 0.0
			replay := 0.0
			analysis := 0.0

			if len(times) == 4 {
				original, _ = strconv.ParseFloat(times[0], 64)
				advocate, _ = strconv.ParseFloat(times[1], 64)
				replay, _ = strconv.ParseFloat(times[2], 64)
				analysis, _ = strconv.ParseFloat(times[3], 64)
			} else if len(times) == 3 {
				original, _ = strconv.ParseFloat(times[0], 64)
				advocate, _ = strconv.ParseFloat(times[1], 64)
				analysis, _ = strconv.ParseFloat(times[2], 64)
			} else {
				return nil
			}

			overheadAdvocate := max(0, (advocate-original)/original)
			overheadsAdvocate = append(overheadsAdvocate, overheadAdvocate)

			if len(times) == 4 {
				overheadReplay := max(0, (replay-original)/original)
				overheadsReplay = append(overheadsReplay, overheadReplay)
			}

			fractionAnalysis := analysis / advocate
			fractionAnalyses = append(fractionAnalyses, fractionAnalysis)

			return nil
		})

	averageOverheadAdvocate, errorOverheadAdvocate := average(overheadsAdvocate)
	averageOverheadReplay, errorOverheadReplay := average(overheadsReplay)
	averageFractionAnalysis, errorFractionAnalysis := average(fractionAnalyses)

	file, err := os.OpenFile(resPath+"/overview.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("# Times")
	file.WriteString("\n\n")
	file.WriteString("Average over all programs\n\n")
	file.WriteString("| Type | Avg | StdErr |\n| - | - | - |\n")
	file.WriteString("| Overhead Advocate | " + strconv.FormatFloat(averageOverheadAdvocate*100, 'f', 6, 64) + "% | " + strconv.FormatFloat(errorOverheadAdvocate*100, 'f', 6, 64) + "% |\n")
	file.WriteString("| Overhead Replay | " + strconv.FormatFloat(averageOverheadReplay*100, 'f', 6, 64) + "% | " + strconv.FormatFloat(errorOverheadReplay*100, 'f', 6, 64) + "% |\n")
	file.WriteString("| Fraction Analysis | " + strconv.FormatFloat(averageFractionAnalysis*100, 'f', 6, 64) + "% | " + strconv.FormatFloat(errorFractionAnalysis*100, 'f', 6, 64) + "% |\n")
	file.WriteString("\n\nInfo: Fraction analysis is runtime of analysis divided by runtime of advocate\n\n")

	return err
}

func average(arr []float64) (float64, float64) {
	sum := 0.0
	for _, val := range arr {
		sum += val
	}
	avg := sum / float64(len(arr))
	stdErr := standardError(arr, avg)
	return avg, stdErr
}

func standardError(arr []float64, avg float64) float64 {
	sum := 0.0
	for _, val := range arr {
		sum += (val - avg) * (val - avg)
	}
	return math.Sqrt(sum / float64(len(arr)))
}

func setup(constructed, gobench bool, medium bool) error {
	// create the res folder if needed
	name := resPath + "/results"
	counter := 1

	for {
		nameStr := name + "_" + strconv.Itoa(counter)
		if _, err := os.Stat(nameStr); os.IsNotExist(err) {
			errDir := os.MkdirAll(nameStr, 0755)
			if errDir != nil {
				return errDir
			}
			resPath = nameStr
			break
		}
		counter++
	}

	log(SETUP, "Create res folder "+resPath)

	if constructed {
		addConstructed()
	}
	if gobench {
		addGoBench()
	}
	if medium {
		addMediumPrograms()
	}
	return nil
}

func main() {
	runConstructed := flag.Bool("c", false, "Run constructed programs")
	runGoBench := flag.Bool("g", false, "Run go benchmarks")
	runMedium := flag.Bool("m", false, "Run medium programs")
	pathAdvocate := flag.String("a", "", "Path to advocate")
	noReplay := flag.Bool("r", false, "Disable replay")
	flag.Parse()

	if *pathAdvocate != "" {
		pathToAdvocate = home + *pathAdvocate
		pathToExamples = pathToAdvocate + "/examples"
		analyzerPath = pathToAdvocate + "/analyzer"
		runEval = pathToAdvocate + "/eval"
		overviewPath = pathToAdvocate + "/eval/createOverview"
		log(INFO, "Using path to advocate: "+pathToAdvocate)
	} else {
		log(INFO, "Using default path to advocate: "+pathToAdvocate)
	}

	if !*runConstructed && !*runGoBench && !*runMedium {
		log(ERROR, "No programs to run")
		log(ERROR, "Select at least one from -c, -g, -m or -i")
		return
	}

	setup(*runConstructed, *runGoBench, *runMedium)

	for _, program := range programs {
		log(START, program[0])

		name := program[0]
		pathToTrace := program[1] + "trace/"
		execPath := program[1]
		progPath := program[2]
		execName := program[3]
		execArgs := program[4:]

		err := createResFolder(name)
		if err != nil {
			log(ERROR, err.Error())
			log(FAILED, name)
			continue
		}

		err = runExecs(execPath, execName, execArgs, name, !(*noReplay))
		runAna := true
		if err != nil {
			log(ERROR, err.Error())
			log(FAILED, name)
			if err.Error() != "Failed to run original" || err.Error() != "Failed to run advocate" {
				runAna = false
			}
			continue
		}

		if runAna {
			out, err := runAnalyzer(name, pathToTrace)
			if err != nil {
				log(ERROR, err.Error()+":\n"+out)
				log(FAILED, name)
				continue
			}
		}

		err = createOverview(name, progPath, execPath)
		if err != nil {
			log(ERROR, err.Error())
			log(FAILED, name)
			continue
		}

		log(DONE, name)
	}

	err := finish()
	if err != nil {
		log(ERROR, err.Error())
		log(FAILED, "Failed to create overfiew")
	} else {
		log(DONE, "Created Overview")
	}

	if numberErrors == 0 {
		log(FINISHS, "Finished with 0 errors")
		for _, failed := range resFailure {
			log(FAILED, failed)
		}
	} else {
		log(FINISHF, "Finished with "+strconv.Itoa(numberErrors)+" errors")
	}
}

type logType int

const (
	START logType = iota
	INFO
	CREATE
	EXEC
	ANALYZE
	FAILED
	TIMEOUT
	DONE
	ERROR
	SETUP
	FINISHS
	FINISHF
	CLEANUP
	OVERVIEW
)

func log(lt logType, message string) {
	res := ""
	switch lt {
	case START:
		res = "\n[START   ] " + end + message
	case INFO:
		res = "\n[INFO    ] " + end + message
	case CREATE:
		res = "[CREATE  ] " + end + message
	case EXEC:
		res = "[EXEC    ] " + end + message
	case ANALYZE:
		res = "[ANALYZE ] " + end + message
	case DONE:
		res = grn + "[DONE    ] " + end + message
		resSucc = append(resSucc, message)
	case FAILED:
		res = red + "[FAILED  ] " + end + message
		resFailure = append(resFailure, message)
	case TIMEOUT:
		res = org + "[TIMEOUT ] " + end + message
	case ERROR:
		res = red + "[ERROR   ] " + end + message
		numberErrors++
	case SETUP:
		res = "[SETUP   ] " + message
	case CLEANUP:
		res = "[CLEANUP ] " + message
	case FINISHS:
		res = "\n" + grn + "[FINISH  ] " + end + message
	case FINISHF:
		res = "\n" + red + "[FINISH  ] " + end + message
	case OVERVIEW:
		res = "[OVERVIEW] " + message
	default:
		res = "[UNKOWN  ] "
	}
	fmt.Println(res)
}
