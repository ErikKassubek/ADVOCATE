package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type bugType string

const (
	lingering bugType = "lingering"
	channel   bugType = "channel"
	mutex     bugType = "mutex"
	wait      bugType = "wait"
	cond      bugType = "cond"
	context   bugType = "context"
)

const (
	pathGoCR     = "/drives/D/AdvocateResults/Experiments/progs/GoCR"
	pathGoPie    = "/drives/D/AdvocateResults/Experiments/progs/GoPie"
	pathGFuzz    = "/drives/D/AdvocateResults/Experiments/progs/GFuzz"
	pathAblation = "/drives/D/AdvocateResults/Experiments/ablationStudyLong"
)

var versions = map[string]string{
	"argo-cd":    "v3.1.0",
	"bleve":      "v2.5.3",
	"bosun":      "v0.8.0",
	"caddy":      "v2.10.0",
	"dns":        "v1.1.50",
	"flannel":    "v0.20.2",
	"frp":        "v0.63.0",
	"gin":        "v1.10.1",
	"gofiber":    "v2.40.1",
	"gorums":     "v0.7.0",
	"grpc":       "v1.51.0",
	"hugo":       "v0.148.2",
	"kubernetes": "v1.25.5",
	"nsq":        "v1.3.0",
	"octant":     "v.0.25.1",
	"ollama":     "v0.11.4",
	"pholcus":    "v1.3.4",
	"syncthing":  "v1.22.1",
	"terraform":  "v1.12.2",
	"zinx":       "v1.2.7",
}

type bug struct {
	leak      bugType
	pos       string
	isTimeout bool
}

type bugNumbers struct {
	ch   int
	ctx  int
	mu   int
	wg   int
	cond int
	ling int
}

type elemNumbers struct {
	max   int
	avg   float64
	tests int
}

func (en *elemNumbers) toString() string {
	return fmt.Sprintf("%d,%d,%f", en.tests, en.max, en.avg)
}

type anaData struct {
	bugNum  bugNumbers
	elemNum elemNumbers
}

func (ad *anaData) toString(elems, lingering bool) string {
	if elems {
		return ad.elemNum.toString() + "," + ad.bugNum.toString(lingering, ad.elemNum.max == 0)
	}
	return ad.bugNum.toString(lingering, ad.elemNum.max == 0)
}

type ablationData struct {
	numBugs    int
	percOneRel float64
	percAllRel float64
}

func (ad *ablationData) toString() string {
	return fmt.Sprintf("%d,%f,%f", ad.numBugs, ad.percOneRel, ad.percAllRel)
}

type progData struct {
	goLeak   anaData
	goCR     anaData
	goPie    anaData
	gFuzz    anaData
	ablation ablationData
}

func (pd *progData) toString() string {
	res := ""
	hasPrintedTestNum := false

	if goCR {
		res += pd.goCR.toString(!hasPrintedTestNum, true)
		hasPrintedTestNum = true
	}

	if goLeak {
		if res != "" {
			res += ","
		}
		res += pd.goLeak.toString(!hasPrintedTestNum, false)
		hasPrintedTestNum = true
	}

	if goPie {
		if res != "" {
			res += ","
		}
		res += pd.goPie.toString(!hasPrintedTestNum, true)
		hasPrintedTestNum = true
	}

	if gFuzz {
		if res != "" {
			res += ","
		}
		res += pd.gFuzz.toString(!hasPrintedTestNum, true)
		hasPrintedTestNum = true
	}

	if ablation {
		res = pd.ablation.toString()
	}

	return res
}

func (bn *bugNumbers) total(lingering bool) int {
	res := bn.ch + bn.ctx + bn.mu + bn.wg + bn.cond
	if lingering {
		res += bn.ling
	}
	return res
}

func (bn *bugNumbers) toString(lingering, noElems bool) string {
	if lingering {
		if onlyTotal {
			if noElems {
				return "-,-"
			}
			return fmt.Sprintf("%d,%d", bn.total(false), bn.total(true))
		}
		if noElems {
			return "-,-,-,-,-,-,-,-"
		}
		return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d", bn.ch, bn.ctx, bn.mu, bn.wg, bn.cond, bn.ling, bn.total(false), bn.total(true))
	}

	if onlyTotal {
		if noElems {
			return "-,-"
		}
		return fmt.Sprintf("%d,%d", bn.total(false), bn.total(true))
	}

	if noElems {
		return "-,-,-,-,-,-,-,-"
	}
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d", bn.ch, bn.ctx, bn.mu, bn.wg, bn.cond, bn.total(false), bn.total(true))

}

var foundBugs = make(map[bug]struct{})
var numberElems = make([]int, 0)

var ablation = false
var goLeak = false
var goCR = false
var goPie = false
var gFuzz = false

var onlyTotal = false

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Select at least one mode")
		return
	}

	for _, mode := range os.Args[1:] {
		switch mode {
		case "goCR":
			goCR = true
		case "goLeak":
			goLeak = true
		case "goPie":
			goPie = true
		case "gFuzz":
			gFuzz = true
		case "ablation":
			ablation = true
		case "total":
			onlyTotal = true
		default:
			println("Unknown mode ", mode)
			return
		}
	}

	res := make(map[string]progData)

	if ablation {
		readData(pathGoCR, "GoCR", res)
	}
	if goCR {
		readData(pathGoCR, "GoCR", res)
	}
	if goLeak {
		readData(pathGoCR, "GoLeak", res)
	}
	if goPie {
		readData(pathGoPie, "GoPie", res)
	}
	if gFuzz {
		readData(pathGFuzz, "GFuzz", res)
	}

	var line string

	println("\n\n\n")

	printHeader()

	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		data := res[name]
		line = name + "," + versions[name] + "," + data.toString()
		if data.goCR.elemNum.max > 0 {
			fmt.Println(line)
		}
	}
}

func readData(path, mode string, res map[string]progData) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "runner" {
				continue
			}

			fullPath := filepath.Join(path, entry.Name())

			log.Printf("%s %s", mode, fullPath)

			numberTests := 0
			numberTestsWithBugs := 0

			percOneRel, percAllRel := 0.0, 0.0

			foundBugs = make(map[bug]struct{})
			numberElems = make([]int, 0)

			err := filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Check for advocateResult folder
				if d.IsDir() {
					if d.Name() == "advocateResult" {
						numberTests, numberTestsWithBugs, err = countSubfolders(path)
						if err != nil {
							log.Printf("Error counting subfolders in %s: %v\n", path, err)
						}
					} else if strings.HasPrefix(d.Name(), "advocateTrace") {
						handleAdvocateTrace(path)
					}
				}

				// Check for specific files
				if !d.IsDir() {
					if strings.HasPrefix(d.Name(), "leak_") {
						if mode == "GoLeak" && !strings.HasPrefix(d.Name(), "leak_0_") {
							return nil
						}

						err := readBugFile(path)
						if err != nil {
							log.Printf("Error reading log file %s: %v\n", path, err)
						}
					}
					if ablation && strings.HasPrefix(d.Name(), "statsFuzz_") {
						percOneRel, percAllRel, err = readFuzzStats(path)
						if err != nil {
							log.Printf("Error reading fuzz stats %s: %v\n", path, err)
						}
					}
				}
				return nil
			})

			if err != nil {
				log.Fatalf("Error walking the path %q: %v\n", pathGoCR, err)
			}

			bugs := countBugs()
			_, max, avg := countElems()

			elemNum := elemNumbers{
				max, avg, numberTests,
			}

			data := anaData{bugs, elemNum}
			ablDat := ablationData{numberTestsWithBugs, percOneRel, percAllRel}

			if numberTests > 0 {
				d, ok := res[entry.Name()]
				if !ok {
					d = progData{}
				}

				if mode == "GoCR" {
					d.goCR = data
				} else if mode == "GoPie" {
					d.goPie = data
				} else if mode == "GoLeak" {
					d.goLeak = data
				} else if mode == "GFuzz" {
					d.gFuzz = data
				} else if mode == "ablation" {
					d.ablation = ablDat
				} else {
					println("Unknown mode ", mode)
				}
				res[entry.Name()] = d
			}

		}
	}
}

func printHeader() {
	fmt.Println(getHeader())
}

func getHeader() string {
	res := ""
	if !ablation {
		res = "ProjectName,Version,TestsNumber,EventsMaxNumberAmongTests,EventsAverageNumberAmongTests,"
	}

	if goCR {
		if !onlyTotal {
			res += "GoCRBlockingBugNumberChannelSelect,GoCRBlockingBugNumberChannelSelectContext,GoCRBlockingBugNumberMutex,GoCRBlockingBugNumberWaitgroup,GoCRBlockingBugNumberConditional,GoCRLingeringBugNumber"
		}
		if res != "" {
			res += ","
		}
		res += "GoCRTotalBugNumber,GoCRTotalBugNumberWithLingering"
	}

	if goLeak {
		if res != "" && !strings.HasSuffix(res, "-") {
			res += ","
		}
		if !onlyTotal {
			res += "GoLeakBlockingBugNumberChannelSelect,GoLeakBlockingBugNumberChannelSelectContext,GoLeakBlockingBugNumberMutex,GoLeakBlockingBugNumberWaitgroup,GoLeakBlockingBugNumberConditional"
		}
		if !strings.HasSuffix(res, "-") {
			res += ","
		}
		res += "GoLeakTotalBugNumber,GoLeakTotalBugNumberWithLingering"
	}

	if goPie {
		if res != "" && !strings.HasSuffix(res, "-") {
			res += ","
		}
		if !onlyTotal {
			res += "GoPieBlockingBugNumberChannelSelect,GoPieBlockingBugNumberChannelSelectContext,GoPieBlockingBugNumberMutex,GoPieBlockingBugNumberWaitgroup,GoPieBlockingBugNumberConditional,GoPieLingeringBugNumber"
		}
		if !strings.HasSuffix(res, "-") {
			res += ","
		}
		res += "GoPieTotalBugNumber,GoPieTotalBugNumberWithLingering"
	}

	if gFuzz {
		if res != "" && !strings.HasSuffix(res, "-") {
			res += ","
		}
		if !onlyTotal {
			res += "GFuzzBlockingBugNumberChannelSelect,GFuzzBlockingBugNumberChannelSelectContext,GFuzzBlockingBugNumberMutex,GFuzzBlockingBugNumberWaitgroup,GFuzzBlockingBugNumberConditional,GFuzzLingeringBugNumber"
		}
		if !strings.HasSuffix(res, "-") {
			res += ","
		}
		res += "GFuzzTotalBugNumber,GFuzzTotalBugNumberWithLingering"
	}

	if ablation {
		if res != "" && !strings.HasSuffix(res, "-") {
			res += ","
		}
		res += "NumberTestsWithBug,PercentageOneActiveReleased,PercentageAllActiveReleased"
	}

	return res
}

// Count the number of folders in a given directory
func countSubfolders(folderPath string) (int, int, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, 0, err
	}

	count := 0
	bugsCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			count++
			bugsPath := filepath.Join(folderPath, entry.Name(), "bugs")
			info, err := os.Stat(bugsPath)
			if err == nil && info.IsDir() {
				bugsCount++
			}
		}
	}
	return count, bugsCount, nil
}

// Called when "advocate trace" file is found
func handleAdvocateTrace(rootPath string) error {
	totalNumberElements := 0
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() == "trace_info.log" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if len(line) == 0 {
				continue
			}

			if strings.HasPrefix(line, "E") || strings.HasPrefix(line, "N") || strings.HasPrefix(line, "X") {
				continue
			}

			totalNumberElements++
		}

		return nil
	})

	numberElems = append(numberElems, totalNumberElements)

	return err
}

func checkIsTimeOut(dirPath string) bool {
	fp := filepath.Join(dirPath, "trace_info.log")
	data, err := os.ReadFile(fp)
	if err != nil {
		panic(err)
	}
	fields := strings.Split(string(data), "\n")

	for _, field := range fields {
		if strings.HasPrefix(field, "ExitCode") {
			ec := strings.Split(field, "!")[1]
			return ec == "10"
		}
	}

	return false
}

// Called when "results_machine.log" file is found
func readBugFile(rootPath string) error {
	bug := bug{}

	file, err := os.Open(rootPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "# ") {
			switch line {
			case "# Leak: Leak":
				bug.leak = lingering
			case "# Leak: Leak on unbuffered channel", "# Leak: Leak on buffered channel", "# Leak: Leak on buffered Channel", "# Leak: Leak on nil channel", "# Leak: Leak on select":
				bug.leak = channel
			case "# Leak: Leak on sync.Mutex":
				bug.leak = mutex
			case "# Leak: Leak on sync.WaitGroup":
				bug.leak = wait
			case "# Leak: Leak on sync.Cond":
				bug.leak = cond
			case "# Leak: Leak on channel or select on context":
				bug.leak = context
			default:
				fmt.Printf("UNKNOWN LEAK TYPE '%s' IN BUG FILE", line)
			}
		}

		if strings.HasPrefix(line, "->") {
			bug.pos = line
		}

		if strings.HasPrefix(line, "- Trace: ") {
			traceName := strings.TrimPrefix(line, "- Trace: ")
			tracePath := filepath.Join(filepath.Dir(filepath.Dir(rootPath)), traceName)
			bug.isTimeout = checkIsTimeOut(tracePath)
		}
	}

	foundBugs[bug] = struct{}{}

	return scanner.Err()
}

func countBugs() bugNumbers {
	bn := bugNumbers{}

	for key := range foundBugs {
		if !key.isTimeout {
			switch key.leak {
			case lingering:
				bn.ling++
			case channel:
				bn.ch++
			case mutex:
				bn.mu++
			case wait:
				bn.wg++
			case cond:
				bn.cond++
			case context:
				bn.ctx++
			default:
				// fmt.Printf("UNKNOWN LEAK TYPE '%s' IN COUNTING\n", key.leak)
			}
		}
	}

	return bn
}

func countElems() (int, int, float64) {
	min := -1
	max := -1
	total := -1

	if len(numberElems) == 0 {
		return 0, 0, 0
	}

	for _, elems := range numberElems {
		if min == -1 || elems < min {
			min = elems
		}
		if max == -1 || elems > max {
			max = elems
		}
		total += elems
	}

	avg := float64(total) / float64(len(numberElems))
	return min, max, avg
}

func readFuzzStats(path string) (float64, float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	numberTotal := 0.0
	numberOneRel := 0.0
	numberAllRel := 0.0

	for scanner.Scan() {
		line := scanner.Text()

		lineSplit := strings.Split(line, ",")
		if len(lineSplit) != 5 {
			return 0, 0, fmt.Errorf("Expected 5 elements, got %d", len(lineSplit))
		}

		oneRel := lineSplit[3]
		allRel := lineSplit[4]

		numberTotal++

		if allRel == "1" {
			numberAllRel++
			// fix small bug in recording
			oneRel = "1"
		}

		if oneRel == "1" {
			numberOneRel++
		}
	}

	return 100 * numberOneRel / numberTotal, 100 * numberAllRel / numberTotal, nil
}
