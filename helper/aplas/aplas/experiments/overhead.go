package experiments

import (
	"aplas/command"
	"aplas/data"
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var maxElem = 1

func overhead() {
	wg.Add(1)
	defer wg.Done()
	go overheadRout()
	go overheadElemDiff()
	go overheadElemSame()
}

func overheadRout() {
	start()
	defer done()

	outputName := "overheadRout"
	fmt.Printf("Start: %s\n", outputName)

	path, err := data.CreateOutput(outputName)
	if err != nil {
		fmt.Println(outputName + " -> " + err.Error())
	}

	pathBase, err := createProgBase(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	data.CopyConstructed(pathBase)

	for i := 0; i <= maxElem; i++ {
		r := max(2, int(math.Pow(10, math.Sqrt(float64(i)))))
		d := createProgRout(path, r)
		writeCreatedProg(pathBase, fmt.Sprintf("%d", r), d)

		settings := append(defaultSettings, []string{
			"-path",
			pathBase,
			"-timeoutExec",
			"240",
		}...)

		start := time.Now()
		out := bytes.Buffer{}
		err = command.RunCommandGo(&out, fmt.Sprintf("base_%s_%d", outputName, i), pathBase, "test", "./...", "-count=1", "timeout", "240s")
		if err != nil {
			fmt.Println(err.Error())
		}
		dur := time.Since(start)

		writeData(outputName+"Base", strconv.Itoa(i), dur, "")

		start = time.Now()
		out = bytes.Buffer{}
		err = command.RunCommandAdvocate(&out, fmt.Sprintf("%s_%d", outputName, i), settings...)
		if err != nil {
			fmt.Println(err.Error())
		}
		dur = time.Since(start)

		writeData(outputName, strconv.Itoa(i), dur, "")
	}

	fmt.Printf("Finish: %s\n", outputName)
}

func overheadElemSame() {
	start()
	defer done()

	outputName := "overheadElemSame"
	fmt.Printf("Start: %s\n", outputName)

	path, err := data.CreateOutput(outputName)
	if err != nil {
		fmt.Println(outputName + " -> " + err.Error())
	}

	pathBase, err := createProgBase(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	data.CopyConstructed(pathBase)

	for i := 0; i <= maxElem; i++ {
		r := max(2, int(math.Pow(10, math.Sqrt(float64(i)))))
		d := createProgElemSame(path, r)
		writeCreatedProg(pathBase, fmt.Sprintf("%d", r), d)

		settings := append(defaultSettings, []string{
			"-path",
			pathBase,
			"-timeoutExec",
			"240",
		}...)

		start := time.Now()
		out := bytes.Buffer{}
		err = command.RunCommandGo(&out, fmt.Sprintf("base_%s_%d", outputName, i), pathBase, "test", "./...", "-count=1", "timeout", "240s")
		if err != nil {
			fmt.Println(err.Error())
		}
		dur := time.Since(start)

		writeData(outputName+"Base", strconv.Itoa(i), dur, "")

		start = time.Now()
		out = bytes.Buffer{}
		err = command.RunCommandAdvocate(&out, fmt.Sprintf("%s_%d", outputName, i), settings...)
		if err != nil {
			fmt.Println(err.Error())
		}
		dur = time.Since(start)

		writeData(outputName, strconv.Itoa(i), dur, "")
	}

	fmt.Printf("Finish: %s\n", outputName)
}

func overheadElemDiff() {
	start()
	defer done()

	outputName := "overheadDiff"
	fmt.Printf("Start: %s\n", outputName)

	path, err := data.CreateOutput(outputName)
	if err != nil {
		fmt.Println(outputName + " -> " + err.Error())
	}

	pathBase, err := createProgBase(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	data.CopyConstructed(pathBase)

	for i := 0; i <= maxElem; i++ {
		r := max(2, int(math.Pow(10, math.Sqrt(float64(i)))))
		d := createProgElemDiff(path, r)
		writeCreatedProg(pathBase, fmt.Sprintf("%d", r), d)

		settings := append(defaultSettings, []string{
			"-path",
			pathBase,
			"-timeoutExec",
			"600",
		}...)

		start := time.Now()
		out := bytes.Buffer{}
		err = command.RunCommandGo(&out, fmt.Sprintf("base_%s_%d", outputName, i), pathBase, "test", "./...", "-count=1", "timeout", "240s")
		if err != nil {
			fmt.Println(err.Error())
		}
		dur := time.Since(start)

		writeData(outputName+"Base", strconv.Itoa(i), dur, "")

		start = time.Now()
		out = bytes.Buffer{}
		err = command.RunCommandAdvocate(&out, fmt.Sprintf("%s_%d", outputName, i), settings...)
		if err != nil {
			fmt.Println(err.Error())
		}
		dur = time.Since(start)

		writeData(outputName, strconv.Itoa(i), dur, "")
	}

	fmt.Printf("Finish: %s\n", outputName)
}

func createProgRout(path string, elem int) string {
	numElemPerRout := int(math.Pow(10, float64(maxElem)) / float64(elem))
	prog := fmt.Sprintf("package main\n\nimport (\n\"time\"\n\"sync\"\n\"testing\"\n)\n\nfunc Test%d(t *testing.T) {\nc := make(chan int, 10)\nwg := sync.WaitGroup{}\n\n", elem)

	routSend := fmt.Sprintf("wg.Go(func() {\nfor i := 0; i < %d; i++ {\nc <- 1\ntime.Sleep(100 * time.Millisecond)\n}\n})\n\n", numElemPerRout)
	routRecv := fmt.Sprintf("wg.Go(func() {\nfor i := 0; i < %d; i++ {\n<-c\ntime.Sleep(100 * time.Millisecond)\n}\n})\n\n", numElemPerRout)

	for i := 0; i < elem/2; i++ {
		prog += routSend
		prog += routRecv
	}

	prog += "wg.Wait()\n}"

	return prog
}

func createProgElemSame(path string, elem int) string {
	prog := fmt.Sprintf("package main\n\nimport (\n\"time\"\n\"sync\"\n\"testing\"\n)\n\nfunc Test%d(t *testing.T) {\nc := make(chan int, 10)\nwg := sync.WaitGroup{}\n\n", elem)

	routSend := fmt.Sprintf("wg.Go(func() {\nfor i := 0; i < %d; i++ {\nc <- 1\ntime.Sleep(100 * time.Millisecond)\n}\n})\n\n", elem)
	routRecv := fmt.Sprintf("wg.Go(func() {\nfor i := 0; i < %d; i++ {\n<-c\ntime.Sleep(100 * time.Millisecond)\n}\n})\n\n", elem)

	for i := 0; i < 5; i++ {
		prog += routSend
		prog += routRecv
	}

	prog += "wg.Wait()\n}"

	return prog
}

func createProgElemDiff(path string, elem int) string {
	prog := fmt.Sprintf("package main\n\nimport (\n\"time\"\n\"sync\"\n\"testing\"\n)\n\nfunc Test%d(t *testing.T) {\nwg := sync.WaitGroup{}\n\n", elem)

	for i := 0; i < elem; i++ {
		prog += fmt.Sprintf("c%d := make(chan int, 10)\n", i)
	}

	routSend := "wg.Go(func() {\n"
	for i := 0; i < elem; i++ {
		routSend += fmt.Sprintf("c%d <- 1\ntime.Sleep(100 * time.Millisecond)\n\n", i)
	}
	routSend += "\n})\n\n"
	routRecv := "wg.Go(func() {\n"
	for i := 0; i < elem; i++ {
		routRecv += fmt.Sprintf("<-c%d\ntime.Sleep(100 * time.Millisecond)\n\n", i)
	}
	routRecv += "\n})\n\n"

	prog += routSend
	prog += routRecv

	prog += "wg.Wait()\n}"

	return prog
}

func createProgBase(path string) (string, error) {
	pathBase := filepath.Join(path, "prog")
	err := os.Mkdir(pathBase, 0755)
	if err != nil {
		return pathBase, err
	}

	return pathBase, nil
}

func writeCreatedProg(path, id, d string) error {
	fp := filepath.Join(path, "Rout_test.go")
	os.Remove(fp)
	f, err := os.OpenFile(fp,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(d)

	return nil
}
